package modules

import (
	"PreFlight/pm"
	"PreFlight/utils"
	"context"
	"fmt"
	"regexp"
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

	// Skip this module if Node.js is not installed.
	if err != nil {
		return nil, nil, nil
	}

	composerConfig := pm.LoadComposerConfig()

	if composerConfig.Error != nil {
		errors = append(errors, fmt.Sprintf("Failed to read composer.json: %v", composerConfig.Error))
		return errors, warnings, successes
	}

	// Validate PHP version.
	if composerConfig.PHPVersion != "" {
		isValid, _ := utils.ValidateVersion(phpVersion, composerConfig.PHPVersion)
		eolVersions := []string{"7.4", "8.0"}

		feedback := fmt.Sprintf("Installed %sPHP (%s ⟶ required %s), Built: (%s, %s).", utils.Reset, phpVersion, composerConfig.PHPVersion, buildDate, vcVersion)
		isWarning := false

		for _, eolVersion := range eolVersions {
			if strings.HasPrefix(phpVersion, eolVersion+".") {
				feedback = fmt.Sprintf("Installed %sPHP (%s ⟶ End-of-Life), Consider upgrading!", utils.Reset, phpVersion)
				isWarning = true
				break
			}
		}

		if !isValid {
			errors = append(errors, fmt.Sprintf("Installed %sPHP (%s ⟶ required %s), Built: (%s, %s).", utils.Reset, phpVersion, composerConfig.PHPVersion, buildDate, vcVersion))
		} else if isWarning {
			warnings = append(warnings, feedback)
		} else {
			successes = append(successes, feedback)
		}
	}

	// Check PHP extensions.
	if len(composerConfig.PHPExtensions) > 0 {
		installedExtensions, err := getPhpExtensions(ctx)

		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to check PHP extensions: %v", err))
			return errors, warnings, successes
		}

		deprecatedExtensions := map[string]struct{}{
			"imap": {}, "mysql": {}, "recode": {}, "statistics": {}, "wddx": {}, "xml-rpc": {},
		}

		experimentalExtensions := map[string]struct{}{
			"gmagick": {}, "imagemagick": {}, "mqseries": {}, "parle": {}, "rnp": {}, "svm": {}, "svn": {}, "ui": {}, "omq": {},
		}

		for _, ext := range composerConfig.PHPExtensions {
			if _, exists := installedExtensions[ext]; exists {
				feedback := fmt.Sprintf("Installed extension %s%s.", utils.Reset, ext)
				isWarning := false

				if _, deprecated := deprecatedExtensions[ext]; deprecated {
					feedback = fmt.Sprintf("Installed extension %s(%s ⟶ deprecated), Consider removing or replacing it.", utils.Reset, ext)
					isWarning = true
				} else if _, experimental := experimentalExtensions[ext]; experimental {
					feedback = fmt.Sprintf("Installed extension %s(%s ⟶ experimental), Use with caution.", utils.Reset, ext)
					isWarning = true
				}

				if isWarning {
					warnings = append(warnings, feedback)
				} else {
					successes = append(successes, feedback)
				}

				continue
			}

			// Handle PHP 8.4+ split extensions.
			if checkPHP84OrHigher(phpVersion) {
				pdoExtensions := map[string][]string{
					"pdo": {"pdo_sqlite", "pdo_mysql", "pdo_pgsql", "pdo_oci", "pdo_odbc", "pdo_firebird"},
				}

				if alternatives, isSplitExt := pdoExtensions[ext]; isSplitExt {
					for _, altExt := range alternatives {
						if _, exists := installedExtensions[altExt]; exists {
							successes = append(successes, fmt.Sprintf("Installed extension %s%s (%s).", utils.Reset, ext, altExt))
							goto NextExtension
						}
					}
				}
			}

			errors = append(errors, fmt.Sprintf("Missing extension %s%s, Please enable it.", utils.Reset, ext))

		NextExtension:
		}
	}

	return errors, warnings, successes
}

// getPhpVersion Retrieves the installed PHP version.
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

	if matches := versionRegex.FindStringSubmatch(lines[0]); len(matches) >= 2 {
		phpVersion = matches[1]
	} else {
		return "", "", "", fmt.Errorf("could not parse PHP version from: %s", lines[0])
	}

	buildRegex := regexp.MustCompile(`\(built: ([^)]+)\) \((.*?)\)`)

	if matches := buildRegex.FindStringSubmatch(lines[0]); len(matches) >= 3 {
		buildDate, vcVersion = matches[1], matches[2]
	} else {
		buildDate, vcVersion = "unknown", "unknown"
	}

	return phpVersion, buildDate, vcVersion, nil
}

// getPhpExtensions Retrieves the installed PHP extensions.
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

// checkPHP84OrHigher Determines if the PHP version is 8.4 or higher.
func checkPHP84OrHigher(phpVersion string) bool {
	parts := strings.Split(phpVersion, ".")

	if len(parts) < 2 {
		return false
	}

	major, err1 := strconv.Atoi(parts[0])
	minor, err2 := strconv.Atoi(parts[1])

	return err1 == nil && err2 == nil && (major > 8 || (major == 8 && minor >= 4))
}
