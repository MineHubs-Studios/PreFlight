package core

import (
	"PreFlight/config"
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

	// USE A SET TO AVOID PACKAGE MANAGER DUPLICATIONS
	pmSet := make(map[string]struct{})

	for _, pm := range packageManagers {
		pmSet[pm] = struct{}{}
	}

	// PROCESS COMPOSER DEPENDENCIES.
	if _, exists := pmSet["composer"]; exists {
		composerConfig := config.LoadComposerConfig()

		if composerConfig.HasJSON && composerConfig.Error == nil {
			composerDeps := append(composerConfig.Dependencies, composerConfig.DevDependencies...)

			if len(composerDeps) > 0 {
				sort.Strings(composerDeps)
				result.Dependencies["composer"] = composerDeps
			} else {
				result.Dependencies["composer"] = []string{}
			}
		} else {
			result.Dependencies["composer"] = []string{}
		}
	}

	// PROCESS NPM/PNPM/Yarn DEPENDENCIES (package.json).
	jsPackageManagers := []string{"npm", "pnpm", "yarn"}
	hasJSPackageManager := false

	for _, pm := range jsPackageManagers {
		if _, exists := pmSet[pm]; exists {
			hasJSPackageManager = true
			break
		}
	}

	if hasJSPackageManager {
		pkgConfig := config.LoadPackageConfig()
		if pkgConfig.HasJSON && pkgConfig.Error == nil {
			allDeps := append(pkgConfig.Dependencies, pkgConfig.DevDependencies...)
			sort.Strings(allDeps)
			for _, pm := range jsPackageManagers {
				if _, exists := pmSet[pm]; exists {
					result.Dependencies[pm] = allDeps
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
		goConfig := config.LoadGoConfig()

		if goConfig.Error == nil && len(goConfig.Modules) > 0 {
			sort.Strings(goConfig.Modules)
			result.Dependencies["go"] = goConfig.Modules
		} else {
			result.Dependencies["go"] = []string{}
		}
	}

	return result
}

// detectAvailablePackageManagers CHECKS WHICH PACKAGE MANAGERS ARE AVAILABLE IN THE PROJECT
func detectAvailablePackageManagers() []string {
	var availablePMs []string

	if _, err := os.Stat("composer.json"); !os.IsNotExist(err) {
		availablePMs = append(availablePMs, "composer")
	}

	packageConfig := config.LoadPackageConfig()
	pm := modules.DeterminePackageManager(packageConfig)

	if packageConfig.HasJSON && pm.Command != "" {
		availablePMs = append(availablePMs, pm.Command)
	}

	if _, err := os.Stat("go.mod"); !os.IsNotExist(err) {
		availablePMs = append(availablePMs, "go")
	}

	return availablePMs
}

// PrintDependencies PRINTS THE FOUND DEPENDENCIES.
func PrintDependencies(result DependencyResult) bool {
	ow := utils.NewOutputWriter()

	if !ow.Println(utils.Bold + utils.Blue + "\n╭─────────────────────────────────────────╮" + utils.Reset) {
		return false
	}

	if !ow.Println(utils.Bold + utils.Blue + "│" + utils.Cyan + utils.Bold + "  🚀 Scanning project for dependencies  " + utils.Reset) {
		return false
	}

	if !ow.Println(utils.Bold + utils.Blue + "╰─────────────────────────────────────────╯" + utils.Reset) {
		return false
	}

	if !ow.PrintNewLines(1) {
		return false
	}

	if len(result.Dependencies) == 0 {
		if !ow.Println(utils.Bold + "📦 No Package Managers:" + utils.Reset) {
			return false
		}

		if !ow.Println(utils.Red + "  " + utils.CrossMark + " No package managers detected in this project!" + utils.Reset) {
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

		if !ow.Printf("%s%s Dependencies:%s\n", utils.Bold, displayName, utils.Reset) {
			return false
		}

		if len(deps) > 0 {
			for _, dep := range deps {
				if !ow.Printf(utils.Green+" "+utils.CheckMark+" %s%s\n", dep, utils.Reset) {
					return false
				}
			}
		} else {
			if !ow.Printf(utils.Red+" "+utils.CrossMark+" No %s dependencies found!%s\n", pm, utils.Reset) {
				return false
			}
		}

		if !ow.Println("") {
			return false
		}
	}

	return true
}
