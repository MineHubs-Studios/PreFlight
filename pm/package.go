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
	HasJSON         bool
	Error           error
}

// LoadPackageConfig parses package.json, lock files and returns PackageConfig.
func LoadPackageConfig() PackageConfig {
	packageConfig := PackageConfig{}
	packageConfig.PackageManager = utils.DetectPackageManager("package")

	packageConfig.HasJSON = packageConfig.PackageManager.ConfigFileExists

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
	packageConfig.NPMVersion = strings.TrimSpace(data.Engines.NPM)
	packageConfig.PNPMVersion = strings.TrimSpace(data.Engines.PNPM)
	packageConfig.YarnVersion = strings.TrimSpace(data.Engines.Yarn)

	packageConfig.Dependencies = make([]string, 0, len(data.Dependencies))

	for dep := range data.Dependencies {
		packageConfig.Dependencies = append(packageConfig.Dependencies, dep)
	}

	utils.SortStrings(packageConfig.Dependencies)

	packageConfig.DevDependencies = make([]string, 0, len(data.DevDependencies))

	for devDep := range data.DevDependencies {
		packageConfig.DevDependencies = append(packageConfig.DevDependencies, devDep)
	}

	utils.SortStrings(packageConfig.DevDependencies)

	return packageConfig
}
