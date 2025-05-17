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
	// modulesMutex protects registeredModules and availableModules from concurrent modifications.
	modulesMutex sync.RWMutex

	// registeredModules contains the active module.
	registeredModules = make(map[string]Module)

	// availableModules contains all known modules that are registered.
	availableModules = make(map[string]Module)
)

// SortType defines the sorting method to be applied to modules.
type SortType string

const (
	// SortByPriority sorts modules based on a predefined priority order.
	SortByPriority SortType = "priority"

	// SortByName sorts modules alphabetically by name.
	SortByName SortType = "name"
)

var defaultPriority = map[string]int{
	"php":      1,
	"composer": 2,
	"node":     3,
	"bun":      4,
	"yarn":     5,
	"pnpm":     6,
	"npm":      7,
}

const fallbackPriority = 1000

// RegisterModule registers a new module if it doesn't already exist.
func RegisterModule(module Module, moduleNames ...string) error {
	modulesMutex.Lock()
	defer modulesMutex.Unlock()

	if module != nil {
		return registerSingleModule(module)
	}

	if len(moduleNames) == 0 {
		return registerAllModules()
	}

	return registerSpecificModules(moduleNames)
}

// registerSingleModule register a single module.
func registerSingleModule(module Module) error {
	if _, exists := registeredModules[module.Name()]; exists {
		return fmt.Errorf("module with name '%s' is already registered", module.Name())
	}

	registeredModules[module.Name()] = module
	return nil
}

// registerAllModules register all available modules.
func registerAllModules() error {
	var errs []string

	for name, mod := range availableModules {
		if _, exists := registeredModules[name]; !exists {
			registeredModules[name] = mod
		} else {
			errs = append(errs, fmt.Sprintf("module '%s' is already registered", name))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("error registering modules: %s", strings.Join(errs, "; "))
	}

	return nil
}

// registerSpecificModules register specific modules by name.
func registerSpecificModules(moduleNames []string) error {
	var errs []string

	for _, name := range moduleNames {
		mod, exists := availableModules[name]

		if !exists {
			errs = append(errs, fmt.Sprintf("unknown module: %s", name))
			continue
		}

		if _, registered := registeredModules[name]; registered {
			errs = append(errs, fmt.Sprintf("module with name '%s' is already registered", name))
			continue
		}

		registeredModules[name] = mod
	}

	if len(errs) > 0 {
		return fmt.Errorf("error registering modules: %s", strings.Join(errs, "; "))
	}

	return nil
}

// GetModules returns a copy of the registered modules.
func GetModules() []Module {
	modulesMutex.RLock()
	defer modulesMutex.RUnlock()

	modules := make([]Module, 0, len(registeredModules))

	for _, module := range registeredModules {
		modules = append(modules, module)
	}

	return modules
}

// RegisterAvailableModule adds a module to the list of available modules.
func RegisterAvailableModule(name string, module Module) {
	if module == nil {
		return
	}

	modulesMutex.Lock()
	defer modulesMutex.Unlock()

	availableModules[name] = module
}

// SortModules sorts modules based on the specified sort type.
func SortModules(modules []Module, sortType ...SortType) []Module {
	sortedModules := make([]Module, len(modules))
	copy(sortedModules, modules)

	actualSortType := SortByPriority

	if len(sortType) > 0 {
		actualSortType = sortType[0]
	}

	switch actualSortType {
	case SortByName:
		sort.SliceStable(sortedModules, func(i, j int) bool {
			return sortedModules[i].Name() < sortedModules[j].Name()
		})
	default:
		sort.SliceStable(sortedModules, func(i, j int) bool {
			return getPriority(sortedModules[i].Name()) < getPriority(sortedModules[j].Name())
		})
	}

	return sortedModules
}

// getPriority retrieves the priority of a module by its name.
func getPriority(name string) int {
	if p, ok := defaultPriority[strings.ToLower(name)]; ok {
		return p
	}

	return fallbackPriority
}
