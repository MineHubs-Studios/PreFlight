package config

import (
	"PreFlight/utils"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type PackageJSON struct {
	Engines struct {
		Node string `json:"node"`
		NPM  string `json:"npm,omitempty"`
		PNPM string `json:"pnpm,omitempty"`
		Yarn string `json:"yarn,omitempty"`
	} `json:"engines"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

type PackageConfig struct {
	PackageManager  utils.PackageManager
	NodeVersion     string
	NPMVersion      string
	PNPMVersion     string
	YarnVersion     string
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
	packageConfig.PackageManager = utils.DetectPackageManager("package")

	if _, err := os.Stat("package.json"); err == nil {
		packageConfig.HasJSON = true
	} else {
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
	packageConfig.NPMVersion = strings.TrimSpace(data.Engines.NPM)
	packageConfig.PNPMVersion = strings.TrimSpace(data.Engines.PNPM)
	packageConfig.YarnVersion = strings.TrimSpace(data.Engines.Yarn)

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
