package modules

import (
	"PreFlight/utils"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type NpmModule struct{}

func (n NpmModule) Name() string {
	return "NPM"
}

func (n NpmModule) IsApplicable(ctx context.Context) bool {
	if ctx.Err() != nil {
		return false
	}

	// CHECK IF PACKAGE.JSON EXISTS.
	if _, err := os.Stat("package.json"); err == nil {
		return true
	}

	// DETERMINE PACKAGE MANAGER AND CHECK FOR LOCK FILE.
	pm := DeterminePackageManager()

	if pm.LockFile != "" {
		return true
	}

	// CHECK IF NODE_MODULES EXISTS.
	if _, err := os.Stat("node_modules"); err == nil {
		return true
	}

	return false
}

type PackageManager struct {
	Command  string // Command TO RUN (npm, pnpm, yarn)
	LockFile string // ASSOCIATED LOCK FILE.
}

func DeterminePackageManager() PackageManager {
	if _, err := os.Stat("pnpm-lock.yaml"); err == nil {
		return PackageManager{Command: "pnpm", LockFile: "pnpm-lock.yaml"}
	}

	if _, err := os.Stat("yarn.lock"); err == nil {
		return PackageManager{Command: "yarn", LockFile: "yarn.lock"}
	}

	if _, err := os.Stat("package-lock.json"); err == nil {
		return PackageManager{Command: "npm", LockFile: "package-lock.json"}
	}

	// DEFAULT TO NPM WITH NO LOCK FILE.
	return PackageManager{Command: "npm", LockFile: ""}
}

// CheckRequirements CHECK THE REQUIREMENTS FOR THE NPM MODULE.
func (n NpmModule) CheckRequirements(ctx context.Context, params map[string]interface{}) (errors []string, warnings []string, successes []string) {
	// CHECK IF CONTEXT IS CANCELED.
	if ctx.Err() != nil {
		return nil, nil, nil
	}

	// READ package.json TO EXTRACT DEPENDENCIES.
	_, requiredDeps, found := utils.ReadPackageJSON()

	// DETERMINE WHICH PACKAGE MANAGER TO USE.
	pm := DeterminePackageManager()

	// HANDLE MISSING package.json.
	if !found {
		errors = append(errors, "package.json not found.")

		if pm.LockFile != "" {
			warnings = append(warnings, fmt.Sprintf("package.json not found, but %s exists. Ensure package.json is included in your project.", pm.LockFile))
		} else {
			warnings = append(warnings, "Neither package.json nor lock files (package-lock.json, yarn.lock, pnpm-lock.yaml) are found.")
		}

		return errors, warnings, successes
	}

	successes = append(successes, "package.json found.")

	// GET ALL INSTALLED PACKAGES.
	installedPackages, err := getInstalledPackages()

	if err != nil {
		warnings = append(warnings, fmt.Sprintf("Error getting installed packages: %v", err))
	}

	// CHECK REQUIRED PACKAGES.
	for _, dep := range requiredDeps {
		if isInstalled, exists := installedPackages[dep]; exists && isInstalled {
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
	packageJSON, err := os.ReadFile("package.json")

	if err == nil {
		var pkgData map[string]interface{}

		if err = json.Unmarshal(packageJSON, &pkgData); err == nil {
			// GET ALL DEPENDENCIES AND DEV DEPENDENCIES.
			allDeps := make(map[string]interface{})

			if deps, ok := pkgData["dependencies"].(map[string]interface{}); ok {
				for name, version := range deps {
					allDeps[name] = version
				}
			}

			if devDeps, ok := pkgData["devDependencies"].(map[string]interface{}); ok {
				for name, version := range devDeps {
					allDeps[name] = version
				}
			}

			// CHECK EACH DEPENDENCY IN NODE_MODULES.
			for name := range allDeps {
				var path string

				if strings.HasPrefix(name, "@") {
					// HANDLE SCOPED PACKAGES.
					parts := strings.SplitN(name, "/", 2)

					if len(parts) == 2 {
						path = filepath.Join("node_modules", parts[0], parts[1])
					}
				} else {
					path = filepath.Join("node_modules", name)
				}

				if path != "" {
					if _, err := os.Stat(path); err == nil {
						installedPackages[name] = true
					}
				}
			}
		}
	}

	// IF WE COULDN'T READ package.json, TRY TO SCAN NODE_MODULES DIRECTLY
	if len(installedPackages) == 0 {
		entries, err := os.ReadDir("node_modules")

		if err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					name := entry.Name()

					// HANDLE @scope DIRECTORIES
					if strings.HasPrefix(name, "@") {
						scopedEntries, err := os.ReadDir(filepath.Join("node_modules", name))

						if err == nil {
							for _, scopedEntry := range scopedEntries {
								if scopedEntry.IsDir() {
									installedPackages[name+"/"+scopedEntry.Name()] = true
								}
							}
						}
					} else {
						installedPackages[name] = true
					}
				}
			}
		}
	}

	return installedPackages, nil
}
