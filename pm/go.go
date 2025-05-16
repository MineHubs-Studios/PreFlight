package pm

import (
	"PreFlight/utils"
	"fmt"
	"os"
	"strings"
)

type GoConfig struct {
	PackageManager utils.PackageManager
	GoVersion      string
	Modules        []string
	HasMod         bool
	Error          error
}

// LoadGoConfig parses go.mod and returns GoConfig.
func LoadGoConfig() GoConfig {
	var goConfig GoConfig

	goConfig.PackageManager = utils.DetectPackageManager("go")
	goConfig.HasMod = goConfig.PackageManager.ConfigFileExists

	// Early return if not applicable.
	if !goConfig.HasMod {
		return goConfig
	}

	// Read and parse go.mod.
	data, err := os.ReadFile("go.mod")

	if err != nil {
		goConfig.Error = fmt.Errorf("could not read go.mod: %w", err)
		return goConfig
	}

	// Extract information from parsed data.
	parseGoMod(&goConfig, string(data))

	return goConfig
}

// parseGoMod extracts information from go.mod content.
func parseGoMod(config *GoConfig, content string) {
	lines := strings.Split(content, "\n")
	var insideRequireBlock bool

	config.Modules = make([]string, 0, len(lines)/2)

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		// Extract go version.
		if strings.HasPrefix(line, "go ") {
			config.GoVersion = strings.TrimSpace(strings.TrimPrefix(line, "go "))
			continue
		}

		// Handle require blocks.
		if line == "require (" {
			insideRequireBlock = true
			continue
		}

		if insideRequireBlock {
			if line == ")" {
				insideRequireBlock = false
				continue
			}

			fields := strings.Fields(line)

			if len(fields) >= 2 {
				config.Modules = append(config.Modules, fields[0])
			}

			continue
		}

		// Handle single require statements.
		if strings.HasPrefix(line, "require ") && !strings.Contains(line, "(") {
			fields := strings.Fields(line)

			if len(fields) >= 3 && fields[0] == "require" {
				config.Modules = append(config.Modules, fields[1])
			}
		}
	}

	utils.SortStrings(config.Modules)
}
