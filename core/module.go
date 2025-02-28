package core

import (
	"context"
	"fmt"
)

var registeredModules []Module

// Module INTERFACE DEFINES THE CONTRACT FOR SYSTEM CHECK MODULES.
type Module interface {
	Name() string
	CheckRequirements(ctx context.Context, params map[string]interface{}) (errors []string, warnings []string, successes []string)
}

// RegisterModule REGISTERS A NEW MODULE IF IT DOESN'T ALREADY EXIST.
func RegisterModule(module Module) error {
	if module == nil {
		return fmt.Errorf("cannot register nil module")
	}

	for _, existingModule := range registeredModules {
		if existingModule.Name() == module.Name() {
			return fmt.Errorf("module with name '%s' is already registered", module.Name())
		}
	}

	registeredModules = append(registeredModules, module)

	return nil
}

// GetModules RETURNS A COPY OF THE REGISTERED MODULES.
func GetModules() []Module {
	modules := make([]Module, len(registeredModules))
	copy(modules, registeredModules)

	return modules
}

// Reset CLEARS ALL REGISTERED MODULES.
/*func Reset() {
	registeredModules = nil
} */
