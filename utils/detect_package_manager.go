package utils

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

// DetectPackageManager identifies which package manager should be used.
func DetectPackageManager(packageType string) PackageManager {
	switch packageType {
	case "package":
		configExists := FileExists("package.json")

		if !configExists && !FileExists("bun.lock") && !FileExists("pnpm-lock.yaml") &&
			!FileExists("yarn.lock") && !FileExists("package-lock.json") {

			return PackageManager{
				Name:             "NPM",
				Command:          "npm",
				LockFile:         "",
				ConfigFileExists: false,
				LockFileExists:   false,
			}
		}

		if FileExists("bun.lock") {
			return PackageManager{
				Name:             "Bun",
				Command:          "bun",
				LockFile:         "bun.lock",
				ConfigFileExists: configExists,
				LockFileExists:   true,
			}
		}

	case "composer":
		configExists := FileExists("composer.json")
		lockExists := FileExists("composer.lock")

		if !configExists && !lockExists {
			return PackageManager{
				Name:             "Composer",
				Command:          "composer",
				LockFile:         "",
				ConfigFileExists: false,
				LockFileExists:   false,
			}
		}

		return PackageManager{
			Name:             "Composer",
			Command:          "composer",
			LockFile:         "composer.lock",
			ConfigFileExists: configExists,
			LockFileExists:   lockExists,
		}

	case "go":
		modExists := FileExists("go.mod")

		if !modExists {
			return PackageManager{
				Name:             "Go Modules",
				Command:          "go",
				LockFile:         "",
				ConfigFileExists: false,
				LockFileExists:   false,
			}
		}

		return PackageManager{
			Name:             "Go Modules",
			Command:          "go",
			LockFile:         "go.mod",
			ConfigFileExists: true,
			LockFileExists:   FileExists("go.sum"),
		}
	}

	// DEFAULT FALLBACK.
	return PackageManager{
		Name:             "NPM",
		Command:          "npm",
		LockFile:         "",
		ConfigFileExists: false,
		LockFileExists:   false,
	}
}
