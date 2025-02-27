package core

import (
	"PreFlight/modules"
)

type DependencyResult struct {
	ComposerDeps []string
	NpmDeps      []string
}

func GetAllDependencies() DependencyResult {
	var result DependencyResult

	_, _, composerDeps, composerFound := modules.ReadComposerJSON()

	if composerFound {
		result.ComposerDeps = composerDeps
	}

	_, packageFound, npmDeps := modules.ReadPackageJSON()

	if packageFound && len(npmDeps) > 0 {
		result.NpmDeps = npmDeps
	}

	return result
}
