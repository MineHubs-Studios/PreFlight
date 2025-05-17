package core

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Module defines the contract for system check modules.
type Module interface {
	Name() string
	CheckRequirements(ctx context.Context) (errors []string, warnings []string, successes []string)
}

var (
	modulesMu      sync.RWMutex
	registeredMods = make(map[string]Module)
	availableMods  = make(map[string]Module)
)

type SortType string

const (
	SortByPriority SortType = "priority"
	SortByName     SortType = "name"
)

var priorities = map[string]int{
	"php": 1, "composer": 2, "node": 3,
	"bun": 4, "yarn": 5, "pnpm": 6, "npm": 7,
}

const defaultPriority = 1000

// RegisterModule registers modules, accepts a module instance or module names.
func RegisterModule(module Module, moduleNames ...string) error {
	modulesMu.Lock()
	defer modulesMu.Unlock()

	if module != nil {
		if _, exists := registeredMods[module.Name()]; exists {
			return fmt.Errorf("module '%s' is already registered", module.Name())
		}
		registeredMods[module.Name()] = module
		return nil
	}

	var errs []string

	if len(moduleNames) == 0 {
		// Register all available modules.
		for name, mod := range availableMods {
			if _, exists := registeredMods[name]; exists {
				errs = append(errs, fmt.Sprintf("module '%s' is already registered", name))
				continue
			}

			registeredMods[name] = mod
		}
	} else {
		// Register specific modules.
		for _, name := range moduleNames {
			mod, exists := availableMods[name]

			if !exists {
				errs = append(errs, fmt.Sprintf("unknown module: %s", name))
				continue
			}

			if _, registered := registeredMods[name]; registered {
				errs = append(errs, fmt.Sprintf("module '%s' is already registered", name))
				continue
			}

			registeredMods[name] = mod
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("registration errors: %s", strings.Join(errs, "; "))
	}

	return nil
}

// GetModules returns a copy of registered modules.
func GetModules() []Module {
	modulesMu.RLock()
	defer modulesMu.RUnlock()

	mods := make([]Module, 0, len(registeredMods))

	for _, m := range registeredMods {
		mods = append(mods, m)
	}

	return mods
}

// RegisterAvailableModule adds a module to available modules.
func RegisterAvailableModule(name string, module Module) {
	if module == nil {
		return
	}

	modulesMu.Lock()
	availableMods[name] = module
	modulesMu.Unlock()
}

// SortModules sorts modules by priority or name.
func SortModules(modules []Module, sortType ...SortType) []Module {
	if len(modules) <= 1 {
		return modules
	}

	result := make([]Module, len(modules))
	copy(result, modules)

	st := SortByPriority

	if len(sortType) > 0 {
		st = sortType[0]
	}

	if st == SortByName {
		sort.SliceStable(result, func(i, j int) bool {
			return result[i].Name() < result[j].Name()
		})
	} else {
		sort.SliceStable(result, func(i, j int) bool {
			pi, ok := priorities[strings.ToLower(result[i].Name())]

			if !ok {
				pi = defaultPriority
			}

			pj, ok := priorities[strings.ToLower(result[j].Name())]

			if !ok {
				pj = defaultPriority
			}

			return pi < pj
		})
	}

	return result
}
