package core

import (
	"context"
	"fmt"
	"strings"
)

var registeredModules []Module
var availableModules = make(map[string]Module)

// Module INTERFACE DEFINES THE CONTRACT FOR SYSTEM CHECK MODULES.
type Module interface {
	Name() string
	CheckRequirements(ctx context.Context, params map[string]interface{}) (errors []string, warnings []string, successes []string)
}

// RegisterModule REGISTERS A NEW MODULE IF IT DOESN'T ALREADY EXIST.
func RegisterModule(module Module, moduleNames ...string) error {
	// Case 1: DIRECT MODULE REGISTRATION.
	if len(moduleNames) == 0 && module != nil {
		for _, existing := range registeredModules {
			if existing.Name() == module.Name() {
				return fmt.Errorf("module with name '%s' is already registered", module.Name())
			}
		}

		registeredModules = append(registeredModules, module)

		return nil
	}

	// Case 2: REGISTRATION BY MODULE NAMES.
	if len(moduleNames) == 0 {
		for _, mod := range availableModules {
			if err := RegisterModule(mod); err != nil {
				return fmt.Errorf("failed to register module: %v", err)
			}
		}

		return nil
	}

	// REGISTER SPECIFIED MODULES.
	for _, name := range moduleNames {
		name = strings.TrimSpace(strings.ToLower(name))

		if mod, exists := availableModules[name]; exists {
			if err := RegisterModule(mod); err != nil {
				return fmt.Errorf("failed to register %s module: %v", name, err)
			}
		} else {
			return fmt.Errorf("unknown module: %s", name)
		}
	}

	return nil
}

// GetModules RETURNS A COPY OF THE REGISTERED MODULES.
func GetModules() []Module {
	modules := make([]Module, len(registeredModules))
	copy(modules, registeredModules)

	return modules
}

func RegisterAvailableModule(name string, module Module) {
	availableModules[strings.ToLower(name)] = module
}

// Reset CLEARS ALL REGISTERED MODULES.
/*func Reset() {
	registeredModules = nil
} */
