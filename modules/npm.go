package modules

import (
	"PreFlight/config"
	"PreFlight/utils"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type NpmModule struct{}

func (n NpmModule) Name() string {
	return "NPM"
}

type PackageManager struct {
	Command  string // Command TO RUN (npm, pnpm, yarn)
	LockFile string // ASSOCIATED LOCK FILE.
}

// DeterminePackageManager IDENTIFIES WHICH PACKAGE MANAGER TO USE.
func DeterminePackageManager(pkgConfig config.PackageConfig) PackageManager {
	switch {
	case pkgConfig.HasPnpmLock:
		return PackageManager{Command: "pnpm", LockFile: "pnpm-lock.yaml"}
	case pkgConfig.HasYarnLock:
		return PackageManager{Command: "yarn", LockFile: "yarn.lock"}
	case pkgConfig.HasPackageLock:
		return PackageManager{Command: "npm", LockFile: "package-lock.json"}
	default:
		return PackageManager{Command: "npm", LockFile: ""}
	}
}

// CheckRequirements CHECK THE REQUIREMENTS FOR THE NPM MODULE.
func (n NpmModule) CheckRequirements(ctx context.Context) (errors []string, warnings []string, successes []string) {
	// CHECK IF CONTEXT IS CANCELED.
	if ctx.Err() != nil {
		return nil, nil, nil
	}

	packageConfig := config.LoadPackageConfig()

	// IF package.json, LOCK FILES OR node_modules ARE NOT FOUND, THEN SKIP.
	if !packageConfig.HasJSON && !packageConfig.HasPackageLock && !packageConfig.HasYarnLock && !packageConfig.HasPnpmLock {
		if fi, errModules := os.Stat("node_modules"); os.IsNotExist(errModules) || !fi.IsDir() {
			return nil, nil, nil
		}
	}

	pm := DeterminePackageManager(packageConfig)

	// HANDLE ERRORS FROM LOADING CONFIG.
	if packageConfig.Error != nil {
		errors = append(errors, fmt.Sprintf("Failed to load package configuration: %v", packageConfig.Error))
		return errors, warnings, successes
	}

	if !packageConfig.HasJSON {
		errors = append(errors, "package.json not found.")

		if pm.LockFile != "" {
			warnings = append(warnings, fmt.Sprintf("package.json not found, but %s exists. Ensure package.json is included in your project.", pm.LockFile))
		} else {
			warnings = append(warnings, "Neither package.json nor lock files (package-lock.json, yarn.lock, pnpm-lock.yaml) are found.")
		}

		return errors, warnings, successes
	}

	// HANDLE ENGINES IN package.json.
	enginesConfig := []struct {
		Cmd     string
		Name    string
		Version string
	}{
		{"node", "Node", packageConfig.NodeVersion},
		{"npm", "NPM", packageConfig.NPMVersion},
		{"pnpm", "PNPM", packageConfig.PNPMVersion},
		{"yarn", "Yarn", packageConfig.YarnVersion},
	}

	for _, engine := range enginesConfig {
		if engine.Version == "" || (engine.Cmd != "node" && engine.Cmd != pm.Command) {
			continue
		}

		validCmd := false

		for _, validEngine := range enginesConfig {
			if engine.Cmd == validEngine.Cmd {
				validCmd = true
				break
			}
		}

		if !validCmd {
			warnings = append(warnings, fmt.Sprintf("Skipping potentially unsafe command: '%s'", engine.Cmd))
			continue
		}

		out, err := exec.CommandContext(ctx, engine.Cmd, "--version").Output() //nolint:gosec

		if err != nil {
			warnings = append(warnings, fmt.Sprintf("Could not retrieve version for '%s': %v", engine.Cmd, err))
			continue
		}

		installed := strings.TrimSpace(string(out))

		if valid, msg := utils.ValidateVersion(installed, engine.Version); !valid {
			warnings = append(warnings, fmt.Sprintf("%s version mismatch. %s", engine.Name, msg))
		} else {
			// Ensure only one success message, prioritizing pnpm over npm and yarn
			if len(successes) == 0 || engine.Cmd == "pnpm" || (engine.Cmd == "npm" && !strings.Contains(successes[0], "pnpm")) {
				successes = []string{fmt.Sprintf("%s version meets the engines requirement (%s).", engine.Cmd, installed)}
			}
		}
	}

	successes = append(successes, "package.json found.")

	// GET ALL INSTALLED PACKAGES.
	installedPackages, err := getInstalledPackages()

	if err != nil {
		warnings = append(warnings, fmt.Sprintf("Error getting installed packages: %v", err))
	}

	packageDeps := append(packageConfig.Dependencies, packageConfig.DevDependencies...)

	// CHECK REQUIRED PACKAGES.
	for _, dep := range packageDeps {
		if installedPackages[dep] {
			successes = append(successes, fmt.Sprintf("NPM package %s is installed.", dep))
		} else {
			errors = append(errors, fmt.Sprintf("NPM package %s is missing. Run `%s install %s`.", dep, pm.Command, dep))
		}
	}

	return errors, warnings, successes
}

// getInstalledPackages RETURNS A MAP OF ALL INSTALLED PACKAGES.
func getInstalledPackages() (map[string]bool, error) {
	installedPackages := make(map[string]bool)

	// LOAD DEPENDENCIES FROM package.json
	packageJSON, err := os.ReadFile("package.json")
	if err != nil {
		return nil, err
	}

	var packageData struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}

	if err := json.Unmarshal(packageJSON, &packageData); err != nil {
		return nil, err
	}

	// GET ALL DECLARED DEPENDENCIES.
	packageDeps := make(map[string]struct{})

	for name := range packageData.Dependencies {
		packageDeps[name] = struct{}{}
	}

	for name := range packageData.DevDependencies {
		packageDeps[name] = struct{}{}
	}

	// CHECK IF EACH DEPENDENCY IS PRESENT IN node_modules.
	for dep := range packageDeps {
		path := filepath.Join("node_modules", filepath.FromSlash(dep))

		if _, err := os.Stat(path); err == nil {
			installedPackages[dep] = true
		}
	}

	// FALLBACK: SCAN node_modules DIRECTLY IF NO DEPENDENCIES ARE FOUND ABOVE.
	if len(installedPackages) == 0 {
		entries, err := os.ReadDir("node_modules")

		if err != nil {
			return nil, err
		}

		for _, entry := range entries {
			if entry.IsDir() {
				name := entry.Name()

				if strings.HasPrefix(name, "@") {
					scopedEntries, err := os.ReadDir(filepath.Join("node_modules", name))

					if err == nil {
						for _, scopedEntry := range scopedEntries {
							if scopedEntry.IsDir() {
								packageName := name + "/" + scopedEntry.Name()
								installedPackages[packageName] = true
							}
						}
					}
				} else {
					installedPackages[name] = true
				}
			}
		}
	}

	return installedPackages, nil
}
