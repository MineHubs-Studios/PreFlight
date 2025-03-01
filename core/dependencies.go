package core

import (
	"PreFlight/utils"
	"fmt"
)

type DependencyResult struct {
	Dependencies map[string][]string
}

func GetAllDependencies(packageManagers []string) DependencyResult {
	result := DependencyResult{
		Dependencies: make(map[string][]string),
	}

	if len(packageManagers) == 0 {
		packageManagers = []string{"composer", "npm", "pnpm"}
	}

	for _, pm := range packageManagers {
		switch pm {
		case "composer":
			_, _, composerDeps, _ := utils.ReadComposerJSON()
			result.Dependencies["composer"] = composerDeps
		case "npm", "pnpm":
			_, _, npmDeps := utils.ReadPackageJSON()
			result.Dependencies["npm"] = npmDeps
		}
	}

	return result
}

func PrintDependencies(result DependencyResult) {
	fmt.Println(Bold + "🔍 Scanning project for dependencies...")
	fmt.Println("")

	if len(result.Dependencies) == 0 {
		fmt.Println(Red + " " + CrossMark + " No dependencies found!" + Reset)
		return
	}

	for pm, deps := range result.Dependencies {
		fmt.Printf("%s%s Dependencies:%s\n", Bold, pm, Reset)

		if len(deps) > 0 {
			for _, dep := range deps {
				fmt.Printf(Green+" "+CheckMark+" %s\n", dep+Reset)
			}
		} else {
			fmt.Printf(Red+" "+CrossMark+" No %s dependencies found!%s\n",
				pm, Reset)
		}

		fmt.Println("")
	}
}
