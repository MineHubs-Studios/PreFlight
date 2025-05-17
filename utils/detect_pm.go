package utils

import "strings"

// PackageManager represents a detected package manager.
type PackageManager struct {
	// Name of the package manager.
	Name string

	// Command to execute the package manager.
	Command string

	// LockFile associated with the package manager.
	LockFile string

	// ConfigFileExists indicates if the main config file exists.
	ConfigFileExists bool

	// LockFileExists indicates if the lock file exists.
	LockFileExists bool
}

// packageManagerConfig defines detection rules for a package manager.
type packageManagerConfig struct {
	name       string
	command    string
	configFile string
	lockFile   string
}

// DetectPackageManager identifies which package manager should be used.
func DetectPackageManager(packageType string) PackageManager {
	configs := map[string]packageManagerConfig{
		"package":  {"NPM", "npm", "package.json", "package-lock.json"},
		"composer": {"Composer", "composer", "composer.json", "composer.lock"},
		"go":       {"Go Modules", "go", "go.mod", "go.sum"},
	}

	config, found := configs[packageType]

	if !found {
		return PackageManager{
			Name:             "NPM",
			Command:          "npm",
			LockFile:         "",
			ConfigFileExists: false,
			LockFileExists:   false,
		}
	}

	configExists := FileExists(config.configFile)
	lockExists := FileExists(config.lockFile)

	if packageType == "package" {
		alternatives := map[string]string{
			"bun.lock":       "Bun",
			"pnpm-lock.yaml": "PNPM",
			"yarn.lock":      "Yarn",
		}

		for lockFile, name := range alternatives {
			if FileExists(lockFile) {
				return PackageManager{
					Name:             name,
					Command:          strings.ToLower(name),
					LockFile:         lockFile,
					ConfigFileExists: configExists,
					LockFileExists:   true,
				}
			}
		}
	}

	if !configExists && !lockExists {
		return PackageManager{
			Name:             config.name,
			Command:          config.command,
			LockFile:         "",
			ConfigFileExists: false,
			LockFileExists:   false,
		}
	}

	return PackageManager{
		Name:             config.name,
		Command:          config.command,
		LockFile:         config.lockFile,
		ConfigFileExists: configExists,
		LockFileExists:   lockExists,
	}
}
