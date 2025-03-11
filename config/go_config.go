package config

import (
	"fmt"
	"os"
	"strings"
)

type GoConfig struct {
	GoVersion         string
	RequiredGoVersion string
	Modules           []string
	HasMod            bool
	Error             error
}

func LoadGoConfig() GoConfig {
	var goConfig GoConfig

	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		goConfig.HasMod = false
		return goConfig
	}

	goConfig.HasMod = true

	data, err := os.ReadFile("go.mod")

	if err != nil {
		goConfig.Error = fmt.Errorf("could not read go.mod: %w", err)
		return goConfig
	}

	lines := strings.Split(string(data), "\n")

	var insideRequireBlock bool

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "go ") {
			goConfig.RequiredGoVersion = strings.TrimSpace(strings.TrimPrefix(line, "go "))
		}

		if strings.HasPrefix(line, "require (") {
			insideRequireBlock = true
			continue
		}

		if insideRequireBlock {
			if line == ")" {
				insideRequireBlock = false
				continue
			}

			if line != "" {
				fields := strings.Fields(line)

				if len(fields) >= 2 {
					goConfig.Modules = append(goConfig.Modules, fields[0])
				}
			}
		}

		if strings.HasPrefix(line, "require ") && !strings.Contains(line, "(") {
			fields := strings.Fields(line)

			if len(fields) >= 3 && fields[0] == "require" {
				goConfig.Modules = append(goConfig.Modules, fields[1])
			}
		}
	}

	return goConfig
}
