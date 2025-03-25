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
	"sync"
)

type PackageModule struct{}

func (p PackageModule) Name() string {
	return "Package"
}

// CheckRequirements VERIFIES Package CONFIGURATIONS AND DEPENDENCIES.
func (p PackageModule) CheckRequirements(ctx context.Context) (errors []string, warnings []string, successes []string) {
	// CHECK IF CONTEXT IS CANCELED.
	if ctx.Err() != nil {
		return nil, nil, nil
	}

	packageConfig := config.LoadPackageConfig()
	pm := packageConfig.PackageManager

	if !packageConfig.HasJSON {
		if fi, errModules := os.Stat("node_modules"); os.IsNotExist(errModules) || !fi.IsDir() {
			return nil, nil, nil
		}

		errors = append(errors, "package.json not found.")

		if pm.LockFile != "" {
			warnings = append(warnings, fmt.Sprintf("package.json not found, but %s exists. Ensure package.json is included in your project.", pm.LockFile))
		} else {
			warnings = append(warnings, "Neither package.json nor lock files (package-lock.json, yarn.lock, pnpm-lock.yaml) were found.")
		}

		return errors, warnings, successes
	}

	// HANDLE ENGINES IN package.json.
	engines := map[string]string{
		"node": packageConfig.NodeVersion,
		"npm":  packageConfig.NPMVersion,
		"pnpm": packageConfig.PNPMVersion,
		"yarn": packageConfig.YarnVersion,
	}

	for cmd, requiredVersion := range engines {
		if requiredVersion == "" || (cmd != "node" && cmd != pm.Command) {
			continue
		}

		out, err := exec.CommandContext(ctx, cmd, "--version").Output() //nolint:gosec

		if err != nil {
			warnings = append(warnings, fmt.Sprintf("Could not retrieve version for '%s': %v", cmd, err))
			continue
		}

		installedVersion := strings.TrimSpace(string(out))

		if valid, _ := utils.ValidateVersion(installedVersion, requiredVersion); !valid {
			warnings = append(warnings, fmt.Sprintf("Missing %s%s (%s ⟶ required %s).", utils.Reset, cmd, installedVersion, requiredVersion))
		} else {
			// ENSURE ONLY ONE SUCCESS MESSAGE, PRIORITIZING PNPM OVER NPM AND Yarn.
			if len(successes) == 0 || cmd == "pnpm" || (cmd == "npm" && !strings.Contains(successes[0], "pnpm")) {
				successes = []string{fmt.Sprintf("Installed %s%s (%s ⟶ required %s).", utils.Reset, cmd, installedVersion, requiredVersion)}
			}
		}
	}

	successes = append(successes, "package.json found.")
	installedPackages, err := getInstalledPackages()

	if err != nil {
		warnings = append(warnings, fmt.Sprintf("Error getting installed packages: %v", err))
	}

	for _, dep := range append(packageConfig.Dependencies, packageConfig.DevDependencies...) {
		if version, installed := installedPackages[dep]; installed {
			successes = append(successes, fmt.Sprintf("Installed package %s%s (%s).", utils.Reset, dep, version))
		} else {
			errors = append(errors, fmt.Sprintf("Missing package %s%s, Run `%s install %s`.", utils.Reset, dep, pm.Command, dep))
		}
	}

	return errors, warnings, successes
}

// getInstalledPackages RETRIEVES THE INSTALLED Package DEPENDENCIES.
func getInstalledPackages() (map[string]string, error) {
	installedPackages := make(map[string]string)

	packageConfig := config.LoadPackageConfig()

	if packageConfig.Error != nil {
		return nil, packageConfig.Error
	}

	// COMBINE DEPENDENCIES AND DEV DEPENDENCIES.
	for _, dep := range packageConfig.Dependencies {
		installedPackages[dep] = "unknown"
	}

	for _, devDep := range packageConfig.DevDependencies {
		installedPackages[devDep] = "unknown"
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	for dep := range installedPackages {
		wg.Add(1)

		go func(dep string) {
			defer wg.Done()

			if strings.Contains(dep, "..") || strings.Contains(dep, "/") && !strings.HasPrefix(dep, "@") {
				return
			}

			var path string

			if strings.HasPrefix(dep, "@") {
				parts := strings.SplitN(dep, "/", 2)

				if len(parts) != 2 || strings.Contains(parts[1], "..") || strings.Contains(parts[1], "/") {
					return
				}
				path = filepath.Join("node_modules", parts[0], parts[1], "package.json")
			} else {
				path = filepath.Join("node_modules", dep, "package.json")
			}

			path = filepath.Clean(path)

			if !strings.HasPrefix(path, filepath.Join("node_modules", "")) {
				return
			}

			if data, err := os.ReadFile(path); err == nil {
				var packageInfo struct {
					Version string `json:"version"`
				}

				if json.Unmarshal(data, &packageInfo) == nil && packageInfo.Version != "" {
					mu.Lock()
					installedPackages[dep] = packageInfo.Version
					mu.Unlock()
				}
			}
		}(dep)
	}

	wg.Wait()

	// FALLBACK: SCAN node_modules IF NO INSTALLED PACKAGES FOUND.
	if len(installedPackages) == 0 {
		if entries, err := os.ReadDir("node_modules"); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					name := entry.Name()

					// HANDLE SCOPED PACKAGES (@org/package).
					if strings.HasPrefix(name, "@") {
						if scopedEntries, err := os.ReadDir(filepath.Join("node_modules", name)); err == nil {
							for _, scopedEntry := range scopedEntries {
								if scopedEntry.IsDir() {
									packagePath := filepath.Join("node_modules", name, scopedEntry.Name(), "package.json")

									if data, err := os.ReadFile(packagePath); err == nil {
										var packageInfo struct {
											Version string `json:"version"`
										}

										if json.Unmarshal(data, &packageInfo) == nil && packageInfo.Version != "" {
											installedPackages[name+"/"+scopedEntry.Name()] = packageInfo.Version
										} else {
											installedPackages[name+"/"+scopedEntry.Name()] = "version unknown"
										}
									}
								}
							}
						}
					} else {
						// DEFAULT PACKAGE HANDLING.
						packagePath := filepath.Join("node_modules", name, "package.json")

						if data, err := os.ReadFile(packagePath); err == nil {
							var packageInfo struct {
								Version string `json:"version"`
							}

							if json.Unmarshal(data, &packageInfo) == nil && packageInfo.Version != "" {
								installedPackages[name] = packageInfo.Version
							} else {
								installedPackages[name] = "version unknown"
							}
						}
					}
				}
			}
		}
	}

	return installedPackages, nil
}
