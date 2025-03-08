package core

import (
	"PreFlight/modules"
	"PreFlight/utils"
	"os"
	"sort"
	"strings"
)

// DependencyResult HOLDS DATA ABOUT ALL FOUND DEPENDENCIES.
type DependencyResult struct {
	Dependencies map[string][]string
}

// GetAllDependencies COLLECTS ALL DEPENDENCIES BASED ON SPECIFIED PACKAGE MANAGER.
func GetAllDependencies(packageManagers []string) DependencyResult {
	result := DependencyResult{
		Dependencies: make(map[string][]string),
	}

	// DETECT WHICH PACKAGE MANAGERS ARE ACTUALLY USED IN THE PROJECT
	availablePackageManagers := detectAvailablePackageManagers()

	if len(availablePackageManagers) == 0 {
		return result
	}

	// USE THE DEFAULT PACKAGE MANAGER IF NONE SPECIFIED
	if len(packageManagers) == 0 {
		packageManagers = availablePackageManagers
	} else {
		// FILTER REQUESTED PACKAGE MANAGERS TO ONLY THOSE THAT EXIST IN PROJECT
		filteredPMs := make([]string, 0)

		for _, pm := range packageManagers {
			for _, availPM := range availablePackageManagers {
				if pm == availPM {
					filteredPMs = append(filteredPMs, pm)
					break
				}
			}
		}

		packageManagers = filteredPMs
	}

	// USE A SET TO AVOID PACKAGE MANAGER DUPLICATIONS.
	pmSet := make(map[string]struct{})

	for _, pm := range packageManagers {
		pmSet[pm] = struct{}{}
	}

	// PROCESS COMPOSER DEPENDENCIES.
	if _, exists := pmSet["composer"]; exists {
		_, _, composerDeps, success := utils.ReadComposerJSON()

		if success && len(composerDeps) > 0 {
			// SORT FOR CONSISTENT OUTPUT.
			sort.Strings(composerDeps)
			result.Dependencies["composer"] = composerDeps
		} else {
			result.Dependencies["composer"] = []string{}
		}
	}

	// PROCESS NPM/PNPM/Yarn DEPENDENCIES (THEY ALL USE package.json).
	jsPackageManagers := []string{"npm", "pnpm", "yarn"}
	hasJSPackageManager := false

	for _, pm := range jsPackageManagers {
		if _, exists := pmSet[pm]; exists {
			hasJSPackageManager = true
			break
		}
	}

	if hasJSPackageManager {
		_, npmDeps, success := utils.ReadPackageJSON()

		if success && len(npmDeps) > 0 {
			// SORT FOR CONSISTENT OUTPUT.
			sort.Strings(npmDeps)

			// SAVE UNDER APPROPRIATE KEYS.
			for _, pm := range jsPackageManagers {
				if _, exists := pmSet[pm]; exists {
					result.Dependencies[pm] = npmDeps
				}
			}
		} else {
			for _, pm := range jsPackageManagers {
				if _, exists := pmSet[pm]; exists {
					result.Dependencies[pm] = []string{}
				}
			}
		}
	}

	// PROCESS GO MODULE DEPENDENCIES
	if _, exists := pmSet["go"]; exists {
		goDeps, err := modules.GetRequiredGoModules()

		if err == nil && len(goDeps) > 0 {
			sort.Strings(goDeps)
			result.Dependencies["go"] = goDeps
		} else {
			result.Dependencies["go"] = []string{}
		}
	}

	return result
}

// detectAvailablePackageManagers CHECKS WHICH PACKAGE MANAGERS ARE AVAILABLE IN THE PROJECT
func detectAvailablePackageManagers() []string {
	availablePMs := make([]string, 0)

	// CHECK FOR Composer (composer.json).
	if _, err := os.Stat("composer.json"); !os.IsNotExist(err) {
		availablePMs = append(availablePMs, "composer")
	}

	// CHECK FOR JS PACKAGE MANAGERS (package.json)
	if _, err := os.Stat("package.json"); !os.IsNotExist(err) {
		pm := modules.DeterminePackageManager()

		if pm.Command != "" {
			availablePMs = append(availablePMs, pm.Command)
		}
	}

	// CHECK FOR Go MODULES (go.mod)
	if _, err := os.Stat("go.mod"); !os.IsNotExist(err) {
		availablePMs = append(availablePMs, "go")
	}

	return availablePMs
}

// PrintDependencies PRINTS THE FOUND DEPENDENCIES.
func PrintDependencies(result DependencyResult) bool {
	ow := utils.NewOutputWriter()

	if !ow.Println(Bold + "🔍 Scanning project for dependencies...") {
		return false
	}

	if !ow.Println("") {
		return false
	}

	if len(result.Dependencies) == 0 {
		if !ow.Println(Bold + "📦 No Package Managers:" + Reset) {
			return false
		}

		if !ow.Println(Red + "  " + CrossMark + " No package managers detected in this project!" + Reset) {
			return false
		}

		return true
	}

	// SORT PACKAGE NAMES FOR CONSISTENT OUTPUT.
	packageManagers := make([]string, 0, len(result.Dependencies))

	for pm := range result.Dependencies {
		packageManagers = append(packageManagers, pm)
	}

	sort.Strings(packageManagers)

	for _, pm := range packageManagers {
		deps := result.Dependencies[pm]

		// CONVERT TO A TITLE CASE FOR NICE FORMATTING (npm -> NPM, composer -> Composer)
		displayName := strings.ToUpper(pm[:1]) + pm[1:]

		if !ow.Printf("%s%s Dependencies:%s\n", Bold, displayName, Reset) {
			return false
		}

		if len(deps) > 0 {
			for _, dep := range deps {
				if !ow.Printf(Green+" "+CheckMark+" %s%s\n", dep, Reset) {
					return false
				}
			}
		} else {
			if !ow.Printf(Red+" "+CrossMark+" No %s dependencies found!%s\n", pm, Reset) {
				return false
			}
		}

		if !ow.Println("") {
			return false
		}
	}

	return true
}
