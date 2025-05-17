package pm

import (
	"PreFlight/utils"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Common paths for pie.phar.
var piePharPaths = []string{
	"./pie.phar",
	"/usr/local/bin/pie.phar",
	"/usr/bin/pie.phar",
}

// PIEConfig holds information about PIE and its extensions.
type PIEConfig struct {
	IsInstalled bool
	Extensions  []string
	PharPath    string
	Error       error
}

// LoadPIEConfig detects if PIE is installed and gets extension information.
func LoadPIEConfig() PIEConfig {
	ctx := context.Background()
	pieConfig := PIEConfig{}

	// Check if PIE is installed.
	pieConfig.IsInstalled = checkPIEInstalled()

	if !pieConfig.IsInstalled {
		return pieConfig
	}

	// Find a PIE phar path.
	pharPath, err := findPIEPharPath()
	if err == nil {
		pieConfig.PharPath = pharPath
	} else {
		pieConfig.Error = fmt.Errorf("could not locate pie.phar: %w", err)
		return pieConfig
	}

	// Pass pharPath directly to extension reader.
	extensionsMap := getPIEExtensions(ctx, pharPath)

	for ext := range extensionsMap {
		if ext == "" || ext == "Core" || ext == "standard" ||
			ext == "[PHP Modules]" || ext == "[Zend Modules]" {
			continue
		}

		pieConfig.Extensions = append(pieConfig.Extensions, ext)
	}

	utils.SortStrings(pieConfig.Extensions)

	return pieConfig
}

// checkPIEInstalled checks if PIE is installed.
func checkPIEInstalled() bool {
	// Check if the pie command exists.
	if _, err := utils.RunCommand(context.Background(), "pie", "--version"); err == nil {
		return true
	}

	// Check if pie.phar exists in common locations.
	for _, path := range piePharPaths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	return false
}

// findPIEPharPath locates the pie.phar file.
func findPIEPharPath() (string, error) {
	// Check common paths.
	for _, path := range piePharPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	// Try to find using 'which pie' command.
	output, err := utils.RunCommand(context.Background(), "which", "pie")

	if err == nil && output != "" {
		path := strings.TrimSpace(output)

		// Check if this points to a pie.phar file.
		if filepath.Ext(path) == ".phar" {
			return path, nil
		}

		// Check if there's a pie.phar in the same directory.
		dir := filepath.Dir(path)
		pharPath := filepath.Join(dir, "pie.phar")

		if _, err := os.Stat(pharPath); err == nil {
			return pharPath, nil
		}
	}

	return "", fmt.Errorf("could not find pie.phar")
}

// getPIEExtensions retrieves PIE extensions from phar file and fallback mechanisms.
func getPIEExtensions(ctx context.Context, pharPath string) map[string]struct{} {
	extensions := make(map[string]struct{})

	// pie -m
	if output, err := utils.RunCommand(ctx, "pie", "-m"); err == nil {
		for _, ext := range strings.Split(output, "\n") {
			if ext = strings.TrimSpace(ext); ext != "" {
				extensions[ext] = struct{}{}
			}
		}
	}

	if pharPath == "" {
		return extensions
	}

	// metadata
	metadataScript := fmt.Sprintf(`...`, pharPath)
	if output, err := utils.RunCommand(ctx, "php", "-r", metadataScript); err == nil {
		for _, ext := range strings.Split(output, "\n") {
			if ext = strings.TrimSpace(ext); ext != "" {
				extensions[ext] = struct{}{}
			}
		}
	}

	// dir scan
	scanScript := fmt.Sprintf(`...`, pharPath)
	if output, err := utils.RunCommand(ctx, "php", "-r", scanScript); err == nil {
		for _, ext := range strings.Split(output, "\n") {
			if ext = strings.TrimSpace(ext); ext != "" {
				extensions[ext] = struct{}{}
			}
		}
	}

	// common ext scan
	for _, ext := range []string{"xdebug", "opcache", "pcov"} {
		checkScript := fmt.Sprintf(`...`, pharPath, ext, ext, ext)
		if output, err := utils.RunCommand(ctx, "php", "-r", checkScript); err == nil && strings.Contains(output, "found") {
			extensions[ext] = struct{}{}
		}
	}

	return extensions
}
