package modules

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type ComposerModule struct{}

func (c ComposerModule) Name() string {
	return "composer"
}

func (c ComposerModule) CheckRequirements(context map[string]interface{}) (errors []string, warnings []string, successes []string) {
	// CHECK IF PHP IS INSTALLED.
	phpInstalled, phpVersion := isPhpInstalled()

	if phpInstalled {
		successes = append(successes, fmt.Sprintf("PHP is installed with version: %s.", phpVersion))
	} else {
		errors = append(errors, "PHP is not installed. Please install PHP to use Composer.")
	}

	// READ composer.json AND PARSE THE REQUIRED INFORMATION.
	_, phpExtensions, composerDeps, found := ReadComposerJSON()

	// HANDLE MISSING composer.json.
	if !found {
		errors = append(errors, "composer.json not found.")

		// CHECK FOR composer.lock IF composer.json IS MISSING.
		if _, err := os.Stat("composer.lock"); err == nil {
			warnings = append(warnings, "composer.lock exists. Ensure composer.json is included.")
		} else {
			warnings = append(warnings, "No composer.lock found.")
		}

		return errors, warnings, successes
	}

	successes = append(successes, "composer.json found.")

	// CHECK PHP EXTENSIONS.
	for _, ext := range phpExtensions {
		if !CheckPHPExtension(ext) {
			errors = append(errors, fmt.Sprintf("PHP extension %s is missing. Please enable it.", ext))
		} else {
			successes = append(successes, fmt.Sprintf("PHP extension %s is installed.", ext))
		}
	}

	// CHECK COMPOSER DEPENDENCIES.
	for _, dep := range composerDeps {
		if !CheckComposerPackage(dep) {
			errors = append(errors, fmt.Sprintf("Composer package %s is missing. Run `composer require %s`.", dep, dep))
		} else {
			successes = append(successes, fmt.Sprintf("Composer package %s is installed.", dep))
		}
	}

	return errors, warnings, successes
}

// CHECK IF PHP IS INSTALLED AND RETURN THE VERSION IF AVAILABLE.
func isPhpInstalled() (bool, string) {
	cmd := exec.Command("php", "--version")
	output, err := cmd.Output()

	if err != nil {
		return false, ""
	}

	lines := strings.Split(string(output), "\n")

	if len(lines) > 0 {
		versionLine := lines[0]
		return true, versionLine
	}

	return true, ""
}

// CheckPHPExtension CHECK IF A SPECIFIC PHP EXTENSION IS INSTALLED.
func CheckPHPExtension(extension string) bool {
	cmd := exec.Command("php", "-m")
	output, err := cmd.Output()

	if err != nil {
		return false
	}

	for _, ext := range strings.Split(string(output), "\n") {
		if strings.TrimSpace(ext) == extension {
			return true
		}
	}

	return false
}

// ReadComposerJSON READ composer.json, PARSE REQUIRED PHP VERSION, EXTENSIONS, AND DEPENDENCIES.
func ReadComposerJSON() (string, []string, []string, bool) {
	var phpVersion string
	var phpExtensions []string
	var composerDeps []string

	// CHECK IF composer.json EXISTS.
	if _, err := os.Stat("composer.json"); os.IsNotExist(err) {
		return "", phpExtensions, composerDeps, false
	}

	// READ THE composer.json FILE.
	file, err := os.ReadFile("composer.json")

	if err != nil {
		return "", phpExtensions, composerDeps, false
	}

	// PARSE JSON CONTENT FROM composer.json.
	var data map[string]interface{}

	if err := json.Unmarshal(file, &data); err != nil {
		return "", phpExtensions, composerDeps, false
	}

	// EXTRACT "require" AND "require-dev" SECTIONS.
	if require, ok := data["require"].(map[string]interface{}); ok {
		for dep, version := range require {
			if dep == "php" {
				phpVersion = fmt.Sprintf("%v", version)
			} else if strings.HasPrefix(dep, "ext-") {
				phpExtensions = append(phpExtensions, strings.TrimPrefix(dep, "ext-"))
			} else {
				composerDeps = append(composerDeps, dep)
			}
		}
	}

	if requireDev, ok := data["require-dev"].(map[string]interface{}); ok {
		for dep := range requireDev {
			composerDeps = append(composerDeps, dep)
		}
	}

	return phpVersion, phpExtensions, composerDeps, true
}

// CheckComposerPackage CHECK IF A SPECIFIC COMPOSER PACKAGE IS INSTALLED.
func CheckComposerPackage(packageName string) bool {
	cmd := exec.Command("composer", "show", packageName)
	err := cmd.Run()

	if err != nil {
		return false
	}

	return true
}
