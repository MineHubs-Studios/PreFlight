package core

import (
	"PreFlight/utils"
)

type DependencyResult struct {
	ComposerDeps []string
	NpmDeps      []string
}

func GetAllDependencies() DependencyResult {
	var result DependencyResult

	_, _, composerDeps, composerFound := utils.ReadComposerJSON()

	if composerFound {
		result.ComposerDeps = composerDeps
	}

	_, packageFound, npmDeps := utils.ReadPackageJSON()

	if packageFound && len(npmDeps) > 0 {
		result.NpmDeps = npmDeps
	}

	return result
}
