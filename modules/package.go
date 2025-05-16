package modules

import (
	"PreFlight/pm"
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

	packageConfig := pm.LoadPackageConfig()
	pm := packageConfig.PackageManager

	if !packageConfig.HasJSON {
		if fi, errModules := os.Stat("node_modules"); os.IsNotExist(errModules) || !fi.IsDir() {
			return nil, nil, nil
		}

		errors = append(errors, "package.json not found.")

		if pm.LockFile != "" {
			warnings = append(warnings, fmt.Sprintf("package.json not found, but %s exists. Ensure package.json is included in your project.", pm.LockFile))
		} else {
			warnings = append(warnings, "Neither package.json nor lock files (package-lock.json, bun.lock, pnpm-lock.yaml or yarn.lock) were found.")
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
			// ENSURE ONLY ONE SUCCESS MESSAGE, PRIORITIZING Bun FIRST, Yarn SECOND, PNPM THIRD AND NPM LAST.
			if len(successes) == 0 || cmd == "bun" ||
				(cmd == "yarn" && !strings.Contains(successes[0], "bun")) ||
				(cmd == "pnpm" && !strings.Contains(successes[0], "bun") && !strings.Contains(successes[0], "yarn")) ||
				(cmd == "npm" && !strings.Contains(successes[0], "bun") && !strings.Contains(successes[0], "yarn") && !strings.Contains(successes[0], "pnpm")) {
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

	packageConfig := pm.LoadPackageConfig()

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

	validatePath := func(name string) (string, bool) {
		if strings.Contains(name, "..") || (strings.Contains(name, "/") && !strings.HasPrefix(name, "@")) {
			return "", false
		}

		var path string

		if strings.HasPrefix(name, "@") {
			parts := strings.SplitN(name, "/", 2)

			if len(parts) != 2 || strings.Contains(parts[1], "..") || strings.Contains(parts[1], "/") {
				return "", false
			}

			path = filepath.Join("node_modules", parts[0], parts[1], "package.json")
		} else {
			path = filepath.Join("node_modules", name, "package.json")
		}

		path = filepath.Clean(path)

		if !strings.HasPrefix(path, filepath.Join("node_modules", "")) {
			return "", false
		}

		return path, true
	}

	for dep := range installedPackages {
		wg.Add(1)

		go func(dep string) {
			defer wg.Done()

			path, valid := validatePath(dep)

			if !valid {
				return
			}

			data, err := os.ReadFile(path) //nolint:gosec

			if err == nil {
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
									scopedName := name + "/" + scopedEntry.Name()

									if strings.Contains(name, "..") || strings.Contains(scopedEntry.Name(), "..") ||
										strings.Contains(scopedEntry.Name(), "/") {
										continue
									}

									packagePath := filepath.Join("node_modules", name, scopedEntry.Name(), "package.json")
									packagePath = filepath.Clean(packagePath)

									if !strings.HasPrefix(packagePath, filepath.Join("node_modules", "")) {
										continue
									}

									data, err := os.ReadFile(packagePath)

									if err == nil {
										var packageInfo struct {
											Version string `json:"version"`
										}

										if json.Unmarshal(data, &packageInfo) == nil && packageInfo.Version != "" {
											installedPackages[scopedName] = packageInfo.Version
										} else {
											installedPackages[scopedName] = "version unknown"
										}
									}
								}
							}
						}
					} else {
						// DEFAULT PACKAGE HANDLING.
						if strings.Contains(name, "..") || strings.Contains(name, "/") {
							continue
						}

						packagePath := filepath.Join("node_modules", name, "package.json")
						packagePath = filepath.Clean(packagePath)

						if !strings.HasPrefix(packagePath, filepath.Join("node_modules", "")) {
							continue
						}

						data, err := os.ReadFile(packagePath)

						if err == nil {
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
