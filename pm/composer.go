package pm

import (
	"PreFlight/utils"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type ComposerJSON struct {
	Require    map[string]string `json:"require"`
	RequireDev map[string]string `json:"require-dev"`
}

type ComposerConfig struct {
	PackageManager  utils.PackageManager
	PHPVersion      string
	PHPExtensions   []string
	Dependencies    []string
	DevDependencies []string
	HasConfig       bool
	HasLock         bool
	Error           error
}

// LoadComposerConfig parses composer.json, composer.lock and returns ComposerConfig.
func LoadComposerConfig() ComposerConfig {
	composerConfig := ComposerConfig{}
	composerConfig.PackageManager = utils.DetectPackageManager("composer")

	composerConfig.HasConfig = composerConfig.PackageManager.ConfigFileExists
	composerConfig.HasLock = composerConfig.PackageManager.LockFileExists

	// Early return if not applicable.
	if !composerConfig.HasConfig {
		return composerConfig
	}

	// Read and parse composer.json.
	file, err := os.ReadFile("composer.json")

	if err != nil {
		composerConfig.Error = fmt.Errorf("unable to read composer.json: %w", err)
		return composerConfig
	}

	var data ComposerJSON

	if err := json.Unmarshal(file, &data); err != nil {
		composerConfig.Error = fmt.Errorf("unable to parse composer.json: %w", err)
		return composerConfig
	}

	// Extract information from parsed data.
	parseComposerJSON(&composerConfig, &data)

	return composerConfig
}

// parseComposerJSON extracts information from parsed composer.json
func parseComposerJSON(config *ComposerConfig, data *ComposerJSON) {
	config.Dependencies = make([]string, 0, len(data.Require))
	config.PHPExtensions = make([]string, 0, len(data.Require))

	// Categorize require entries into PHP version, extensions and dependencies.
	for dep, version := range data.Require {
		switch {
		case dep == "php":
			config.PHPVersion = version
		case strings.HasPrefix(dep, "ext-"):
			config.PHPExtensions = append(config.PHPExtensions, strings.TrimPrefix(dep, "ext-"))
		default:
			config.Dependencies = append(config.Dependencies, dep)
		}
	}

	utils.SortStrings(config.PHPExtensions)
	utils.SortStrings(config.Dependencies)

	config.DevDependencies = make([]string, 0, len(data.RequireDev))

	// Extract dev dependencies separately.
	for devDep := range data.RequireDev {
		config.DevDependencies = append(config.DevDependencies, devDep)
	}

	utils.SortStrings(config.DevDependencies)
}
