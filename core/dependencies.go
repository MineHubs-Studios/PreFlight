package core

import (
	"PreFlight/config"
	"PreFlight/utils"
	"sort"
	"strings"
)

// DependencyResult HOLDS DATA ABOUT ALL FOUND DEPENDENCIES.
type DependencyResult struct {
	Dependencies map[string][]string
}

// GetAllDependencies COLLECTS ALL DEPENDENCIES BASED ON DETECTED PACKAGE MANAGERS.
func GetAllDependencies() DependencyResult {
	result := DependencyResult{
		Dependencies: make(map[string][]string),
	}

	// DETECT AND PROCESS COMPOSER DEPENDENCIES.
	composerPM := utils.DetectPackageManager("composer")

	if composerPM.LockFile != "" {
		composerConfig := config.LoadComposerConfig()

		if composerConfig.HasJSON && composerConfig.Error == nil {
			composerDeps := append(composerConfig.Dependencies, composerConfig.DevDependencies...)

			if len(composerDeps) > 0 {
				sort.Strings(composerDeps)
				result.Dependencies["composer"] = composerDeps
			}
		}
	}

	// DETECT AND PROCESS PACKAGE DEPENDENCIES (NPM, PNPM, Yarn).
	packagePM := utils.DetectPackageManager("package")

	if packagePM.LockFile != "" {
		packageConfig := config.LoadPackageConfig()

		if packageConfig.HasJSON && packageConfig.Error == nil {
			allDeps := append(packageConfig.Dependencies, packageConfig.DevDependencies...)

			if len(allDeps) > 0 {
				sort.Strings(allDeps)
				result.Dependencies[packagePM.Command] = allDeps
			}
		}
	}

	// DETECT AND PROCESS GO MODULE DEPENDENCIES
	goPM := utils.DetectPackageManager("go")

	if goPM.LockFile != "" {
		goConfig := config.LoadGoConfig()

		if goConfig.Error == nil && len(goConfig.Modules) > 0 {
			sort.Strings(goConfig.Modules)
			result.Dependencies["go"] = goConfig.Modules
		}
	}

	return result
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
