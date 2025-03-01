package utils

import (
	"encoding/json"
	"os"
	"strings"
)

// ReadPackageJSON READS package.json TO EXTRACT NODE VERSION, DEPENDENCIES, AND DEV DEPENDENCIES.
func ReadPackageJSON() (string, []string, bool) {
	var nodeVersion string
	var npmDeps []string

	// CHECK IF package.json EXISTS.
	if _, err := os.Stat("composer.json"); os.IsNotExist(err) {
		return nodeVersion, npmDeps, false
	}

	// READ package.json FILE.
	file, err := os.ReadFile("package.json")

	if err != nil {
		return nodeVersion, npmDeps, false
	}

	// PARSE JSON CONTENT FROM package.json.
	var data map[string]interface{}

	if err := json.Unmarshal(file, &data); err != nil {
		return nodeVersion, npmDeps, false
	}

	// EXTRACT REQUIRED NODE VERSION FROM "engines" SECTION.
	if engines, ok := data["engines"].(map[string]interface{}); ok {
		if node, exists := engines["node"].(string); exists {
			nodeVersion = strings.TrimSpace(node)
		}
	}

	// EXTRACT DEPENDENCIES AND DEV DEPENDENCIES.
	if dependencies, ok := data["dependencies"].(map[string]interface{}); ok {
		for dep := range dependencies {
			npmDeps = append(npmDeps, dep)
		}
	}

	if devDependencies, ok := data["devDependencies"].(map[string]interface{}); ok {
		for dep := range devDependencies {
			npmDeps = append(npmDeps, dep)
		}
	}

	return nodeVersion, npmDeps, true
}
