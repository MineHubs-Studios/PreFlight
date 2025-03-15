package utils

import "os"

// PackageManager represents a detected package manager.
type PackageManager struct {
	Name     string // Package manager name
	Command  string // Command to execute
	LockFile string // Associated lock file
}

// DetectPackageManager identifies which package manager should be used.
func DetectPackageManager(packageType string) PackageManager {
	switch packageType {
	case "package": // JS Package Managers (Only one can be selected)
		if _, err := os.Stat("pnpm-lock.yaml"); err == nil {
			return PackageManager{Name: "PNPM", Command: "pnpm", LockFile: "pnpm-lock.yaml"}
		}
		if _, err := os.Stat("yarn.lock"); err == nil {
			return PackageManager{Name: "Yarn", Command: "yarn", LockFile: "yarn.lock"}
		}
		if _, err := os.Stat("package-lock.json"); err == nil {
			return PackageManager{Name: "NPM", Command: "npm", LockFile: "package-lock.json"}
		}

	case "composer": // Composer (PHP)
		if _, err := os.Stat("composer.lock"); err == nil {
			return PackageManager{Name: "Composer", Command: "composer", LockFile: "composer.lock"}
		}

	case "go": // Go Modules
		if _, err := os.Stat("go.mod"); err == nil {
			return PackageManager{Name: "Go Modules", Command: "go", LockFile: "go.mod"}
		}
	}

	// Default fallback (useful for JS managers)
	return PackageManager{Name: "NPM", Command: "npm", LockFile: ""}
}
