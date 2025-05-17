package pm

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
	HasConfig       bool
	Error           error
}

// LoadPackageConfig parses package.json, lock files and returns PackageConfig.
func LoadPackageConfig() PackageConfig {
	packageConfig := PackageConfig{}
	packageConfig.PackageManager = utils.DetectPackageManager("package")

	packageConfig.HasConfig = packageConfig.PackageManager.ConfigFileExists

	// Early return if not applicable.
	if !packageConfig.HasConfig {
		return packageConfig
	}

	// Read and parse package.json.
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

	// Extract information from parsed data.
	parsePackageJSON(&packageConfig, &data)

	return packageConfig
}

// parsePackageJSON extracts information from parsed package.json
func parsePackageJSON(config *PackageConfig, data *PackageJSON) {
	// Extract version requirements from engines section.
	config.NodeVersion = strings.TrimSpace(data.Engines.Node)
	config.NPMVersion = strings.TrimSpace(data.Engines.NPM)
	config.PNPMVersion = strings.TrimSpace(data.Engines.PNPM)
	config.YarnVersion = strings.TrimSpace(data.Engines.Yarn)

	// Extract dependencies.
	config.Dependencies = make([]string, 0, len(data.Dependencies))

	for dep := range data.Dependencies {
		config.Dependencies = append(config.Dependencies, dep)
	}

	utils.SortStrings(config.Dependencies)

	config.DevDependencies = make([]string, 0, len(data.DevDependencies))

	for devDep := range data.DevDependencies {
		config.DevDependencies = append(config.DevDependencies, devDep)
	}

	utils.SortStrings(config.DevDependencies)
}
