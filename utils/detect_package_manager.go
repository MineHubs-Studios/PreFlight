package utils

import "os"

// PackageManager REPRESENTS A DETECTED PACKAGE MANAGER.
type PackageManager struct {
	// Name OF THE PACKAGE MANAGER.
	Name string

	// Command TO EXECUTE THE PACKAGE MANAGER.
	Command string

	// LockFile ASSOCIATED WITH THE PACKAGE MANAGER.
	LockFile string
}

// DetectPackageManager IDENTIFIES WHICH PACKAGE MANAGER SHOULD BE USED.
func DetectPackageManager(packageType string) PackageManager {
	switch packageType {
	case "package":
		if _, err := os.Stat("pnpm-lock.yaml"); err == nil {
			return PackageManager{Name: "PNPM", Command: "pnpm", LockFile: "pnpm-lock.yaml"}
		}

		if _, err := os.Stat("yarn.lock"); err == nil {
			return PackageManager{Name: "Yarn", Command: "yarn", LockFile: "yarn.lock"}
		}

		if _, err := os.Stat("package-lock.json"); err == nil {
			return PackageManager{Name: "NPM", Command: "npm", LockFile: "package-lock.json"}
		}

	case "composer":
		if _, err := os.Stat("composer.lock"); err == nil {
			return PackageManager{Name: "Composer", Command: "composer", LockFile: "composer.lock"}
		}

	case "go":
		if _, err := os.Stat("go.mod"); err == nil {
			return PackageManager{Name: "Go Modules", Command: "go", LockFile: "go.mod"}
		}
	}

	// DEFAULT FALLBACK.
	return PackageManager{Name: "NPM", Command: "npm", LockFile: ""}
}
