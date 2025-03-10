package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type ComposerConfig struct {
	PHPVersion      string
	PHPExtensions   []string
	Dependencies    []string
	DevDependencies []string
	HasJSON         bool
	HasLock         bool
	Error           error
}

// LoadComposerConfig PARSES composer.json, composer.lock, AND RETURNS ComposerConfig.
func LoadComposerConfig() ComposerConfig {
	composerConfig := ComposerConfig{}

	if _, err := os.Stat("composer.json"); os.IsNotExist(err) {
		composerConfig.HasJSON = false
	} else {
		composerConfig.HasJSON = true
		file, err := os.ReadFile("composer.json")

		if err != nil {
			composerConfig.Error = fmt.Errorf("could not read composer.json: %w", err)
			return composerConfig
		}

		var data map[string]interface{}

		if err := json.Unmarshal(file, &data); err != nil {
			composerConfig.Error = fmt.Errorf("json parsing composer.json error: %w", err)
			return composerConfig
		}

		if require, ok := data["require"].(map[string]interface{}); ok {
			for dep, version := range require {
				switch {
				case dep == "php":
					composerConfig.PHPVersion = fmt.Sprintf("%v", version)
				case strings.HasPrefix(dep, "ext-"):
					composerConfig.PHPExtensions = append(composerConfig.PHPExtensions, strings.TrimPrefix(dep, "ext-"))
				default:
					composerConfig.Dependencies = append(composerConfig.Dependencies, dep)
				}
			}
		}

		if requireDev, ok := data["require-dev"].(map[string]interface{}); ok {
			for dep := range requireDev {
				composerConfig.DevDependencies = append(composerConfig.DevDependencies, dep)
			}
		}
	}

	if _, err := os.Stat("composer.lock"); os.IsNotExist(err) {
		composerConfig.HasLock = false
	} else {
		composerConfig.HasLock = true
	}

	return composerConfig
}
