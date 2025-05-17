package modules

import (
	"PreFlight/pm"
	"PreFlight/utils"
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type PhpModule struct{}

func (p PhpModule) Name() string {
	return "PHP"
}

// CheckRequirements verifies PHP configurations and extensions.
func (p PhpModule) CheckRequirements(ctx context.Context) (errors []string, warnings []string, successes []string) {
	// Check if context is canceled.
	if ctx.Err() != nil {
		return nil, nil, nil
	}

	phpVersion, buildDate, vcVersion, err := getPhpVersion(ctx)

	// Skip this module if PHP is not installed.
	if err != nil {
		return nil, nil, nil
	}

	// Get composer configuration
	composerConfig := pm.LoadComposerConfig()

	if composerConfig.Error != nil {
		errors = append(errors, fmt.Sprintf("Failed to read composer.json: %v", composerConfig.Error))
		return errors, warnings, successes
	}

	// Validate PHP version.
	if composerConfig.PHPVersion != "" {
		isValid, _ := utils.ValidateVersion(phpVersion, composerConfig.PHPVersion)

		eolVersions := map[string]bool{
			"7.4": true, "8.0": true,
		}

		feedback := fmt.Sprintf("Installed %sPHP (%s ⟶ required %s), Built: (%s, %s).", utils.Reset, phpVersion, composerConfig.PHPVersion, buildDate, vcVersion)
		versionPrefix := strings.Split(phpVersion, ".")[0] + "." + strings.Split(phpVersion, ".")[1]

		if eolVersions[versionPrefix] {
			warnings = append(warnings, fmt.Sprintf("Installed %sPHP (%s ⟶ End-of-Life), Consider upgrading!", utils.Reset, phpVersion))

			if isValid {
				warnings = append(warnings, feedback)
			}
		} else if !isValid {
			errors = append(errors, fmt.Sprintf("Installed %sPHP (%s ⟶ required %s), Built: (%s, %s).", utils.Reset, phpVersion, composerConfig.PHPVersion, buildDate, vcVersion))
		} else {
			successes = append(successes, feedback)
		}
	}

	// Get PHP extensions.
	installedExtensions, err := getPhpExtensions(ctx)

	if err != nil {
		errors = append(errors, fmt.Sprintf("Failed to check PHP extensions: %v", err))
		return errors, warnings, successes
	}

	// Get PIE extensions if it's installed and a PHP version is 8.4+.
	pieConfig := pm.LoadPIEConfig()

	// Create extension source map.
	extensionSources := make(map[string]string)

	for ext := range installedExtensions {
		extensionSources[ext] = "php"
	}

	// Convert PIE's Extensions slice to a map for an easier lookup.
	pieExtensions := make(map[string]struct{})

	if checkPHP84OrHigher(phpVersion) && pieConfig.IsInstalled {
		for _, ext := range pieConfig.Extensions {
			pieExtensions[ext] = struct{}{}
			extensionSources[ext] = "pie"
		}
	}

	// Define deprecated and experimental extensions inline.
	deprecatedExtensions := map[string]struct{}{
		"imap": {}, "mysql": {}, "recode": {}, "statistics": {}, "wddx": {}, "xml-rpc": {},
	}

	experimentalExtensions := map[string]struct{}{
		"gmagick": {}, "imagemagick": {}, "mqseries": {}, "parle": {}, "rnp": {},
		"svm": {}, "svn": {}, "ui": {}, "omq": {},
	}

	// Track extensions to display.
	type ExtensionInfo struct {
		Name      string
		Source    string
		IsWarning bool
		Warning   string
	}

	extensionsToShow := make([]ExtensionInfo, 0, len(pieExtensions)+len(composerConfig.PHPExtensions))

	// Add PIE extensions first.
	for ext := range pieExtensions {
		if ext == "" || ext == "Core" || ext == "standard" ||
			ext == "[PHP Modules]" || ext == "[Zend Modules]" {
			continue
		}

		isWarning := false
		warningMsg := ""

		if _, deprecated := deprecatedExtensions[ext]; deprecated {
			isWarning = true
			warningMsg = fmt.Sprintf("(%s ⟶ deprecated), Consider removing or replacing it.", ext)
		} else if _, experimental := experimentalExtensions[ext]; experimental {
			isWarning = true
			warningMsg = fmt.Sprintf("(%s ⟶ experimental), Use with caution.", ext)
		}

		extensionsToShow = append(extensionsToShow, ExtensionInfo{
			Name:      ext,
			Source:    "pie",
			IsWarning: isWarning,
			Warning:   warningMsg,
		})
	}

	// Process required extensions.
	if len(composerConfig.PHPExtensions) > 0 {
		for _, ext := range composerConfig.PHPExtensions {
			// Skip if already included from PIE.
			alreadyIncluded := false

			for _, info := range extensionsToShow {
				if info.Name == ext {
					alreadyIncluded = true
					break
				}
			}

			if alreadyIncluded {
				continue
			}

			// Check if the extension is installed.
			source, exists := extensionSources[ext]
			if exists {
				isWarning := false
				warningMsg := ""

				if _, deprecated := deprecatedExtensions[ext]; deprecated {
					isWarning = true
					warningMsg = fmt.Sprintf("(%s ⟶ deprecated), Consider removing or replacing it.", ext)
				} else if _, experimental := experimentalExtensions[ext]; experimental {
					isWarning = true
					warningMsg = fmt.Sprintf("(%s ⟶ experimental), Use with caution.", ext)
				}

				extensionsToShow = append(extensionsToShow, ExtensionInfo{
					Name:      ext,
					Source:    source,
					IsWarning: isWarning,
					Warning:   warningMsg,
				})
				continue
			}

			// Handle PHP 8.4+ split extensions.
			if checkPHP84OrHigher(phpVersion) {
				pdoExtensions := map[string][]string{
					"pdo": {"pdo_sqlite", "pdo_mysql", "pdo_pgsql", "pdo_oci", "pdo_odbc", "pdo_firebird"},
				}

				if alternatives, isSplitExt := pdoExtensions[ext]; isSplitExt {
					for _, altExt := range alternatives {
						if _, exists := extensionSources[altExt]; exists {
							extensionsToShow = append(extensionsToShow, ExtensionInfo{
								Name:      ext,
								Source:    "php",
								IsWarning: false,
								Warning:   fmt.Sprintf("(%s)", altExt),
							})
							goto NextExtension
						}
					}
				}
			}

			// Extension is missing.
			errors = append(errors, fmt.Sprintf("Missing extension %s%s, Please enable it.", utils.Reset, ext))

		NextExtension:
		}
	}

	// Sort extensions alphabetically by name.
	sort.Slice(extensionsToShow, func(i, j int) bool {
		return strings.ToLower(extensionsToShow[i].Name) < strings.ToLower(extensionsToShow[j].Name)
	})

	// Generate extension feedback messages.
	for _, extInfo := range extensionsToShow {
		var feedback string

		if extInfo.Source == "pie" {
			if extInfo.IsWarning {
				feedback = fmt.Sprintf("Installed extension %s%s %s", utils.Reset, extInfo.Name, extInfo.Warning)
				warnings = append(warnings, feedback)
			} else {
				feedback = fmt.Sprintf("Installed extension %s%s.", utils.Reset, extInfo.Name)
				successes = append(successes, feedback)
			}
		} else {
			if extInfo.IsWarning {
				feedback = fmt.Sprintf("Installed extension %s%s %s", utils.Reset, extInfo.Name, extInfo.Warning)
				warnings = append(warnings, feedback)
			} else if extInfo.Warning != "" {
				feedback = fmt.Sprintf("Installed extension %s%s %s.", utils.Reset, extInfo.Name, extInfo.Warning)
				successes = append(successes, feedback)
			} else {
				feedback = fmt.Sprintf("Installed extension %s%s.", utils.Reset, extInfo.Name)
				successes = append(successes, feedback)
			}
		}
	}

	return errors, warnings, successes
}

// getPhpVersion retrieves the installed PHP version.
func getPhpVersion(ctx context.Context) (phpVersion, buildDate, vcVersion string, err error) {
	output, err := utils.RunCommand(ctx, "php", "--version")

	if err != nil {
		return "", "", "", fmt.Errorf("failed to run php --version: %w", err)
	}

	lines := strings.Split(output, "\n")

	if len(lines) == 0 {
		return "", "", "", fmt.Errorf("unexpected output from php --version")
	}

	versionRegex := regexp.MustCompile(`PHP (\d+\.\d+\.\d+)`)
	buildRegex := regexp.MustCompile(`\(built: ([^)]+)\) \((.*?)\)`)

	if matches := versionRegex.FindStringSubmatch(lines[0]); len(matches) >= 2 {
		phpVersion = matches[1]
	} else {
		return "", "", "", fmt.Errorf("could not parse PHP version from: %s", lines[0])
	}

	if matches := buildRegex.FindStringSubmatch(lines[0]); len(matches) >= 3 {
		buildDate, vcVersion = matches[1], matches[2]
	} else {
		buildDate, vcVersion = "unknown", "unknown"
	}

	return phpVersion, buildDate, vcVersion, nil
}

// getPhpExtensions retrieves the installed PHP extensions.
func getPhpExtensions(ctx context.Context) (map[string]struct{}, error) {
	output, err := utils.RunCommand(ctx, "php", "-m")

	if err != nil {
		return nil, fmt.Errorf("failed to run php -m: %w", err)
	}

	extensions := make(map[string]struct{})

	for _, ext := range strings.Split(output, "\n") {
		if trimmed := strings.TrimSpace(ext); trimmed != "" {
			extensions[trimmed] = struct{}{}
		}
	}

	return extensions, nil
}

// checkPHP84OrHigher determines if the PHP version is 8.4 or higher.
func checkPHP84OrHigher(phpVersion string) bool {
	parts := strings.Split(phpVersion, ".")

	if len(parts) < 2 {
		return false
	}

	major, err1 := strconv.Atoi(parts[0])
	minor, err2 := strconv.Atoi(parts[1])

	return err1 == nil && err2 == nil && (major > 8 || (major == 8 && minor >= 4))
}
