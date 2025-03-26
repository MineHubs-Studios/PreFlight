package core

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Module DEFINES THE CONTRACT FOR SYSTEM CHECK MODULES.
type Module interface {
	Name() string
	CheckRequirements(ctx context.Context) (errors []string, warnings []string, successes []string)
}

var (
	// modulesMutex PROTECTS registeredModules AND availableModules FROM CONCURRENT MODIFICATIONS.
	modulesMutex sync.RWMutex

	// registeredModules CONTAINS THE ACTIVE MODULES.
	registeredModules = make(map[string]Module)

	// availableModules CONTAINS ALL KNOWN MODULES THAT ARE REGISTERED.
	availableModules = make(map[string]Module)
)

// SortType DEFINES THE SORTING METHOD TO BE APPLIED TO MODULES.
type SortType string

const (
	// SortByPriority SORTS MODULES BASED ON A PREDEFINED PRIORITY ORDER.
	SortByPriority SortType = "priority"

	// SortByName SORTS MODULES ALPHABETICALLY BY NAME.
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

// RegisterModule REGISTERS A NEW MODULE IF IT DOESN'T ALREADY EXIST.
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

// registerSingleModule HELPER FUNCTION TO REGISTER A SINGLE MODULE.
func registerSingleModule(module Module) error {
	if _, exists := registeredModules[module.Name()]; exists {
		return fmt.Errorf("module with name '%s' is already registered", module.Name())
	}

	registeredModules[module.Name()] = module
	return nil
}

// registerAllModules HELPER FUNCTION TO REGISTER ALL AVAILABLE MODULES.
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

// registerSpecificModules HELPER FUNCTION TO REGISTER SPECIFIC MODULES BY NAME.
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

// GetModules RETURNS A COPY OF THE REGISTERED MODULES.
func GetModules() []Module {
	modulesMutex.RLock()
	defer modulesMutex.RUnlock()

	modules := make([]Module, 0, len(registeredModules))

	for _, module := range registeredModules {
		modules = append(modules, module)
	}

	return modules
}

// RegisterAvailableModule ADDS A MODULE TO THE LIST OF AVAILABLE MODULES.
func RegisterAvailableModule(name string, module Module) {
	if module == nil {
		return
	}

	modulesMutex.Lock()
	defer modulesMutex.Unlock()

	availableModules[name] = module
}

// SortModules SORTS MODULES BASED ON THE SPECIFIED SORT TYPE.
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

func getPriority(name string) int {
	if p, ok := defaultPriority[strings.ToLower(name)]; ok {
		return p
	}

	return fallbackPriority
}
