package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type PackageJSON struct {
	Engines struct {
		Node string `json:"node"`
	} `json:"engines"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

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

	if !packageConfig.HasJSON {
		return packageConfig
	}

	file, err := os.ReadFile("package.json")

	if err != nil {
		packageConfig.Error = fmt.Errorf("unable to read package.json: %w", err)
		return packageConfig
	}

	var data PackageJSON

	if err := json.Unmarshal(file, &data); err != nil {
		packageConfig.Error = fmt.Errorf("unable to parse package.json: %w", err)
		return packageConfig
	}

	packageConfig.NodeVersion = strings.TrimSpace(data.Engines.Node)
	packageConfig.Dependencies = make([]string, 0, len(data.Dependencies))

	for dep := range data.Dependencies {
		packageConfig.Dependencies = append(packageConfig.Dependencies, dep)
	}

	packageConfig.DevDependencies = make([]string, 0, len(data.DevDependencies))

	for devDep := range data.DevDependencies {
		packageConfig.DevDependencies = append(packageConfig.DevDependencies, devDep)
	}

	return packageConfig
}
