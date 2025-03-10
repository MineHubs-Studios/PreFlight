package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type PackageConfig struct {
	NodeVersion     string
	Dependencies    []string
	DevDependencies []string
	HasJSON         bool
	HasPackageLock  bool
	HasYarnLock     bool
	HasPnpmLock     bool
	Error           error
}

// LoadPackageConfig PARSES package.json, LOCK FILES AND RETURNS PackageConfig.
func LoadPackageConfig() PackageConfig {
	packageConfig := PackageConfig{}

	if _, err := os.Stat("package.json"); err == nil {
		packageConfig.HasJSON = true
	}

	if _, err := os.Stat("package-lock.json"); err == nil {
		packageConfig.HasPackageLock = true
	}

	if _, err := os.Stat("yarn.lock"); err == nil {
		packageConfig.HasYarnLock = true
	}

	if _, err := os.Stat("pnpm-lock.yaml"); err == nil {
		packageConfig.HasPnpmLock = true
	}

	if packageConfig.HasJSON {
		file, err := os.ReadFile("package.json")

		if err != nil {
			packageConfig.Error = fmt.Errorf("could not read package.json: %w", err)
			return packageConfig
		}

		var data map[string]interface{}

		if err := json.Unmarshal(file, &data); err != nil {
			packageConfig.Error = fmt.Errorf("json parsing package.json error: %w", err)
			return packageConfig
		}

		if engines, ok := data["engines"].(map[string]interface{}); ok {
			if node, exists := engines["node"].(string); exists {
				packageConfig.NodeVersion = strings.TrimSpace(node)
			}
		}

		if deps, ok := data["dependencies"].(map[string]interface{}); ok {
			for dep := range deps {
				packageConfig.Dependencies = append(packageConfig.Dependencies, dep)
			}
		}

		if devDeps, ok := data["devDependencies"].(map[string]interface{}); ok {
			for dep := range devDeps {
				packageConfig.DevDependencies = append(packageConfig.DevDependencies, dep)
			}
		}
	}

	return packageConfig
}
