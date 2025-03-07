package core

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// Module DEFINES THE CONTRACT FOR SYSTEM CHECK MODULES.
type Module interface {
	Name() string
	CheckRequirements(ctx context.Context, params map[string]interface{}) (errors []string, warnings []string, successes []string)
}

var (
	// modulesMutex PROTECTS registeredModules AND availableModules FROM CONCURRENT MODIFICATIONS.
	modulesMutex sync.RWMutex

	// registeredModules CONTAINS THE ACTIVE MODULES.
	registeredModules = make(map[string]Module)

	// availableModules CONTAINS ALL KNOWN MODULES THAT ARE REGISTERED.
	availableModules = make(map[string]Module)
)

// RegisterModule REGISTERS A NEW MODULE IF IT DOESN'T ALREADY EXIST.
func RegisterModule(module Module, moduleNames ...string) error {
	modulesMutex.Lock()
	defer modulesMutex.Unlock()

	// Case 1: DIRECT MODULE REGISTRATION.
	if module != nil {
		name := strings.ToLower(module.Name())

		if _, exists := registeredModules[name]; exists {
			return fmt.Errorf("modul med navnet '%s' er allerede registreret", module.Name())
		}

		registeredModules[name] = module

		return nil
	}

	// Case 2: NO PARAMETERS, REGISTER ALL AVAILABLE MODULES.
	if len(moduleNames) == 0 {
		var errs []string

		for name, mod := range availableModules {
			if _, exists := registeredModules[name]; !exists {
				registeredModules[name] = mod
			} else {
				errs = append(errs, fmt.Sprintf("modul '%s' er allerede registreret", name))
			}
		}

		if len(errs) > 0 {
			return fmt.Errorf("fejl ved registrering af moduler: %s", strings.Join(errs, "; "))
		}

		return nil
	}

	// Case 3: REGISTER SPECIFIED MODULES.
	var errs []string

	for _, name := range moduleNames {
		normalizedName := strings.TrimSpace(strings.ToLower(name))
		mod, exists := availableModules[normalizedName]

		if !exists {
			errs = append(errs, fmt.Sprintf("ukendt modul: %s", name))
			continue
		}

		if _, registered := registeredModules[normalizedName]; registered {
			errs = append(errs, fmt.Sprintf("modul '%s' er allerede registreret", name))
			continue
		}

		registeredModules[normalizedName] = mod
	}

	if len(errs) > 0 {
		return fmt.Errorf("fejl ved registrering af moduler: %s", strings.Join(errs, "; "))
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

	normalizedName := strings.ToLower(name)

	modulesMutex.Lock()
	defer modulesMutex.Unlock()

	availableModules[normalizedName] = module
}

// Reset CLEARS ALL REGISTERED MODULES.
/* func Reset() {
	modulesMutex.Lock()
	defer modulesMutex.Unlock()

	// EMPTY THE MAP BY CREATING A NEW ONE.
	registeredModules = make(map[string]Module)
} */
