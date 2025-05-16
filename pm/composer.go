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
	HasJSON         bool
	HasLock         bool
	Error           error
}

// LoadComposerConfig parses composer.json, composer.lock and returns ComposerConfig.
func LoadComposerConfig() ComposerConfig {
	composerConfig := ComposerConfig{}
	composerConfig.PackageManager = utils.DetectPackageManager("composer")

	composerConfig.HasJSON = composerConfig.PackageManager.ConfigFileExists
	composerConfig.HasLock = composerConfig.PackageManager.LockFileExists

	if !composerConfig.HasJSON {
		return composerConfig
	}

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

	composerConfig.Dependencies = make([]string, 0, len(data.Require))
	composerConfig.PHPExtensions = make([]string, 0, len(data.Require))

	for dep, version := range data.Require {
		switch {
		case dep == "php":
			composerConfig.PHPVersion = version
		case strings.HasPrefix(dep, "ext-"):
			composerConfig.PHPExtensions = append(composerConfig.PHPExtensions, strings.TrimPrefix(dep, "ext-"))
		default:
			composerConfig.Dependencies = append(composerConfig.Dependencies, dep)
		}
	}

	utils.SortStrings(composerConfig.PHPExtensions)
	utils.SortStrings(composerConfig.Dependencies)

	composerConfig.DevDependencies = make([]string, 0, len(data.RequireDev))

	for devDep := range data.RequireDev {
		composerConfig.DevDependencies = append(composerConfig.DevDependencies, devDep)
	}

	utils.SortStrings(composerConfig.DevDependencies)

	return composerConfig
}
