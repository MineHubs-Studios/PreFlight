package core

import (
	"PreFlight/pm"
	"PreFlight/utils"
	"sort"
	"strings"
)

// DependencyResult HOLDS THE RESULT OF ALL FOUND DEPENDENCIES.
type DependencyResult struct {
	Dependencies map[string][]string
}

// dependencyFetcher IS A FUNCTION SIGNATURE FOR FETCHING DEPENDENCIES.
type dependencyFetcher func() (string, []string, error)

// GetAllDependencies COLLECTS DEPENDENCIES FROM SUPPORTED PACKAGE MANAGERS.
func GetAllDependencies(only ...string) DependencyResult {
	allowed := make(map[string]bool)

	for _, name := range only {
		allowed[name] = true
	}

	result := DependencyResult{
		Dependencies: make(map[string][]string),
	}

	fetchers := map[string]dependencyFetcher{
		"composer": fetchComposerDependencies,
		"package":  fetchPackageDependencies,
		"go":       fetchGoDependencies,
	}

	for name, fetch := range fetchers {
		if len(allowed) > 0 && !allowed[name] {
			continue
		}

		depName, deps, err := fetch()

		if err == nil && len(deps) > 0 {
			sort.Strings(deps)
			result.Dependencies[depName] = deps
		}
	}

	return result
}

// fetchComposerDependencies FETCH Composer DEPENDENCIES.
func fetchComposerDependencies() (string, []string, error) {
	cfg := pm.LoadComposerConfig()

	if !cfg.HasJSON || cfg.Error != nil {
		return "", nil, cfg.Error
	}

	if len(cfg.Dependencies)+len(cfg.DevDependencies) == 0 {
		return "", nil, nil
	}

	deps := append(cfg.Dependencies, cfg.DevDependencies...)

	return "composer", deps, nil
}

// fetchPackageDependencies FETCH Package DEPENDENCIES.
func fetchPackageDependencies() (string, []string, error) {
	cfg := pm.LoadPackageConfig()

	if !cfg.HasJSON || cfg.Error != nil {
		return "", nil, cfg.Error
	}

	if len(cfg.Dependencies)+len(cfg.DevDependencies) == 0 {
		return "", nil, nil
	}

	deps := append(cfg.Dependencies, cfg.DevDependencies...)

	return "package", deps, nil
}

// fetchGoDependencies FETCH Go DEPENDENCIES.
func fetchGoDependencies() (string, []string, error) {
	cfg := pm.LoadGoConfig()

	if !cfg.HasMod || cfg.Error != nil || len(cfg.Modules) == 0 {
		return "", nil, cfg.Error
	}

	return "go", cfg.Modules, nil
}

// PrintDependencies PRINTS THE FOUND DEPENDENCIES.
func PrintDependencies(result DependencyResult) bool {
	ow := utils.NewOutputWriter()

	header := []string{
		utils.Bold + utils.Blue + "\n╭─────────────────────────────────────────╮" + utils.Reset,
		utils.Bold + utils.Blue + "│" + utils.Cyan + utils.Bold + "  🚀 Scanning project for dependencies  " + utils.Reset,
		utils.Bold + utils.Blue + "╰─────────────────────────────────────────╯" + utils.Reset,
	}

	for _, line := range header {
		if !ow.Println(line) {
			return false
		}
	}

	ow.PrintNewLines(1)

	if len(result.Dependencies) == 0 {
		ow.Println(utils.Bold + "📦 No Package Managers:" + utils.Reset)
		ow.Println(utils.Red + "  " + utils.CrossMark + " No package managers detected in this project!" + utils.Reset)
		return true
	}

	// SORT FOR CONSISTENT OUTPUT.
	pmNames := make([]string, 0, len(result.Dependencies))

	for name := range result.Dependencies {
		pmNames = append(pmNames, name)
	}

	sort.Strings(pmNames)

	for _, name := range pmNames {
		deps := result.Dependencies[name]
		displayName := strings.ToUpper(name[:1]) + name[1:]

		ow.Printf("%s%s Dependencies:%s\n", utils.Bold, displayName, utils.Reset)

		if len(deps) == 0 {
			ow.Printf(utils.Red+" "+utils.CrossMark+" No %s dependencies found!%s\n", displayName, utils.Reset)
		} else {
			for _, dep := range deps {
				ow.Printf(utils.Green+" "+utils.CheckMark+" %s%s\n", dep, utils.Reset)
			}
		}

		ow.Println("")
	}

	return true
}
