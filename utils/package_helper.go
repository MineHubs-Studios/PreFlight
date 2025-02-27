package utils

import (
	"encoding/json"
	"os"
	"strings"
)

// ReadPackageJSON READS package.json TO EXTRACT NODE VERSION, DEPENDENCIES, AND DEV DEPENDENCIES.
func ReadPackageJSON() (string, bool, []string) {
	// CHECK IF package.json EXISTS.
	if _, err := os.Stat("package.json"); os.IsNotExist(err) {
		return "", false, nil
	}

	// READ package.json FILE CONTENT.
	file, err := os.ReadFile("package.json")

	if err != nil {
		return "", false, nil
	}

	// PARSE JSON CONTENT FROM package.json.
	var data map[string]interface{}

	if err := json.Unmarshal(file, &data); err != nil {
		return "", false, nil
	}

	// EXTRACT REQUIRED NODE VERSION FROM "engines" SECTION.
	var requiredNodeVersion string

	if engines, ok := data["engines"].(map[string]interface{}); ok {
		if node, exists := engines["node"].(string); exists {
			requiredNodeVersion = strings.TrimSpace(node)
		}
	}

	// EXTRACT DEPENDENCIES AND DEV DEPENDENCIES.
	var requiredDeps []string

	if dependencies, ok := data["dependencies"].(map[string]interface{}); ok {
		for dep := range dependencies {
			requiredDeps = append(requiredDeps, dep)
		}
	}

	if devDependencies, ok := data["devDependencies"].(map[string]interface{}); ok {
		for dep := range devDependencies {
			requiredDeps = append(requiredDeps, dep)
		}
	}

	return requiredNodeVersion, true, requiredDeps
}
