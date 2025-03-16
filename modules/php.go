package modules

import (
	"PreFlight/config"
	"PreFlight/utils"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type PhpModule struct{}

func (p PhpModule) Name() string {
	return "PHP"
}

// CheckRequirements VERIFIES PHP CONFIGURATIONS AND EXTENSIONS.
func (p PhpModule) CheckRequirements(ctx context.Context) (errors []string, warnings []string, successes []string) {
	// CHECK IF CONTEXT IS CANCELED.
	if ctx.Err() != nil {
		return nil, nil, nil
	}

	phpVersion, buildDate, vcVersion, err := getPhpVersion(ctx)

	// SKIP MODULE IF PHP IS NOT INSTALLED.
	if err != nil {
		return nil, nil, nil
	}

	composerConfig := config.LoadComposerConfig()

	if composerConfig.Error != nil {
		errors = append(errors, fmt.Sprintf("Failed to read composer.json: %v", composerConfig.Error))
		return errors, warnings, successes
	}

	// VALIDATE PHP VERSION.
	if composerConfig.PHPVersion != "" {
		isValid, _ := utils.ValidateVersion(phpVersion, composerConfig.PHPVersion)
		eolVersions := []string{"7.4", "8.0"}

		if isValid {
			successes = append(successes, fmt.Sprintf("Installed %sPHP (%s ⟶ required %s), Built: (%s, %s).", utils.Reset, phpVersion, composerConfig.PHPVersion, buildDate, vcVersion))

			// Check for End-of-Life (EOL) PHP versions.
			for _, eolVersion := range eolVersions {
				if strings.HasPrefix(phpVersion, eolVersion+".") {
					warnings = append(warnings, fmt.Sprintf("Installed %sPHP (%s ⟶ End-of-Life), consider upgrading!", utils.Reset, phpVersion))
					break
				}
			}
		} else {
			errors = append(errors, fmt.Sprintf("Installed %sPHP (%s ⟶ required %s), Built: (%s, %s).", utils.Reset, phpVersion, composerConfig.PHPVersion, buildDate, vcVersion))
		}
	}

	// CHECK PHP EXTENSIONS.
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
				message := fmt.Sprintf("Installed extension %s%s.", utils.Reset, ext)

				if _, deprecated := deprecatedExtensions[ext]; deprecated {
					message = fmt.Sprintf("Installed extension %s(%s ⟶ deprecated), Consider removing or replacing it.", utils.Reset, ext)
				} else if _, experimental := experimentalExtensions[ext]; experimental {
					message = fmt.Sprintf("Installed extension %s(%s ⟶ experimental), Use with caution.", utils.Reset, ext)
				}

				successes = append(successes, message)
				continue
			}

			// Handle PHP 8.4+ split extensions
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

// getPhpVersion RETRIEVES THE INSTALLED PHP VERSION.
func getPhpVersion(ctx context.Context) (phpVersion, buildDate, vcVersion string, err error) {
	cmd := exec.CommandContext(ctx, "php", "--version")
	output, err := cmd.Output()

	if err != nil {
		return "", "", "", fmt.Errorf("failed to run php --version: %w", err)
	}

	lines := strings.Split(string(output), "\n")

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

// getPhpExtensions RETRIEVES THE INSTALLED PHP EXTENSIONS.
func getPhpExtensions(ctx context.Context) (map[string]struct{}, error) {
	cmd := exec.CommandContext(ctx, "php", "-m")
	output, err := cmd.Output()

	if err != nil {
		return nil, fmt.Errorf("failed to run php -m: %w", err)
	}

	extensions := make(map[string]struct{})

	for _, ext := range strings.Split(string(output), "\n") {
		if trimmed := strings.TrimSpace(ext); trimmed != "" {
			extensions[trimmed] = struct{}{}
		}
	}

	return extensions, nil
}

// checkPHP84OrHigher DETERMINES IF THE PHP VERSION IS 8.4 OR HIGHER.
func checkPHP84OrHigher(phpVersion string) bool {
	parts := strings.Split(phpVersion, ".")

	if len(parts) < 2 {
		return false
	}

	major, err1 := strconv.Atoi(parts[0])
	minor, err2 := strconv.Atoi(parts[1])

	return err1 == nil && err2 == nil && (major > 8 || (major == 8 && minor >= 4))
}
