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

func (p PhpModule) CheckRequirements(ctx context.Context) (errors []string, warnings []string, successes []string) {
	// CHECK IF CONTEXT IS CANCELED.
	if ctx.Err() != nil {
		return nil, nil, nil
	}

	phpVersion, buildDate, vcVersion, err := getPhpVersion(ctx)

	// IF PHP IS NOT INSTALLED, THEN SKIP.
	if err != nil {
		return nil, nil, nil
	}

	composerConfig := config.LoadComposerConfig()

	if composerConfig.Error != nil {
		errors = append(errors, fmt.Sprintf("Failed to read composer.json: %v", composerConfig.Error))
		return errors, warnings, successes
	}

	// CHECK PHP VERSION AGAINST REQUIREMENT.
	if composerConfig.PHPVersion != "" {
		isValid, _ := utils.ValidateVersion(phpVersion, composerConfig.PHPVersion)

		if isValid {
			eolVersions := []string{"7.4", "8.0"}
			successes = append(successes, fmt.Sprintf("Installed %sPHP (%s ⟶ required %s), Built: (%s, %s).", utils.Reset, phpVersion, composerConfig.PHPVersion, buildDate, vcVersion))

			for _, eolVersion := range eolVersions {
				if strings.HasPrefix(phpVersion, eolVersion+".") {
					warnings = append(warnings, fmt.Sprintf("Installed %sPHP (%s ⟶ End-of-Life), consider upgrading!", utils.Reset, phpVersion))
				}
			}
		} else {
			errors = append(errors, fmt.Sprintf("Installed %sPHP (%s ⟶ required %s), Built: (%s, %s).", utils.Reset, phpVersion, composerConfig.PHPVersion, buildDate, vcVersion))
		}
	}

	if len(composerConfig.PHPExtensions) > 0 {
		// GET ALL INSTALLED PHP EXTENSIONS.
		installedExtensions, err := getPhpExtensions(ctx)

		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to check PHP extensions: %v", err))
			return errors, warnings, successes
		}

		// CHECK REQUIRED PHP EXTENSIONS.
		for _, ext := range composerConfig.PHPExtensions {
			if _, exists := installedExtensions[ext]; exists {
				deprecatedExtensions := map[string]struct{}{
					"imap": {}, "mysql": {}, "recode": {}, "statistics": {}, "wddx": {}, "xml-rpc": {},
				}

				experimentalExtensions := map[string]struct{}{
					"gmagick": {}, "imagemagick": {}, "mqseries": {}, "parle": {}, "rnp": {}, "svm": {}, "svn": {}, "ui": {}, "omq": {},
				}

				feedback := fmt.Sprintf("Installed extension %s%s.", utils.Reset, ext)

				// CHECK FOR DEPRECATED AND EXPERIMENTAL EXTENSIONS.
				if _, deprecated := deprecatedExtensions[ext]; deprecated {
					feedback = fmt.Sprintf("Installed extension %s(%s ⟶ deprecated), Consider removing or replacing it.", utils.Reset, ext)
				} else if _, experimental := experimentalExtensions[ext]; experimental {
					feedback = fmt.Sprintf("Installed extension %s(%s ⟶ experimental), Use with caution.", utils.Reset, ext)
				}

				successes = append(successes, feedback)

				continue
			}

			// FOR PHP 8.4+, CHECK IF THIS IS A SPLIT EXTENSION.
			isPHP84OrHigher := checkPHP84OrHigher(phpVersion)

			if isPHP84OrHigher {
				pdoExtensions := map[string][]string{
					"pdo": {"pdo_sqlite", "pdo_mysql", "pdo_pgsql", "pdo_oci", "pdo_odbc", "pdo_firebird"},
				}

				if alternatives, isSplitExt := pdoExtensions[ext]; isSplitExt {
					foundAlternative := false

					for _, altExt := range alternatives {
						if _, exists := installedExtensions[altExt]; exists {
							foundAlternative = true
							successes = append(successes, fmt.Sprintf("Installed extension %s%s (%s).", utils.Reset, ext, altExt))
							break
						}
					}

					if foundAlternative {
						continue
					}
				}
			}

			errors = append(errors, fmt.Sprintf("Missing extension %s%s, Please enable it.", utils.Reset, ext))
		}
	}

	return errors, warnings, successes
}

// getPhpVersion RETURNS THE INSTALLED PHP VERSION OR AN ERROR.
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

	// EXTRACT PHP VERSION.
	versionRegex := regexp.MustCompile(`PHP (\d+\.\d+\.\d+)`)
	versionMatches := versionRegex.FindStringSubmatch(lines[0])

	if len(versionMatches) < 2 {
		return "", "", "", fmt.Errorf("could not parse PHP version from: %s", lines[0])
	}

	phpVersion = versionMatches[1]

	// EXTRACT BUILD DATE AND VC++ VERSION.
	buildRegex := regexp.MustCompile(`\(built: ([^)]+)\) \((.*?)\)`)
	buildMatches := buildRegex.FindStringSubmatch(lines[0])

	if len(buildMatches) >= 3 {
		buildDate = buildMatches[1]
		vcVersion = buildMatches[2]
	} else {
		buildDate = "unknown"
		vcVersion = "unknown"
	}

	return phpVersion, buildDate, vcVersion, nil
}

// getPhpExtensions RETURNS A MAP OF ALL INSTALLED PHP EXTENSIONS.
func getPhpExtensions(ctx context.Context) (map[string]struct{}, error) {
	cmd := exec.CommandContext(ctx, "php", "-m")
	output, err := cmd.Output()

	if err != nil {
		return nil, fmt.Errorf("failed to run php -m: %w", err)
	}

	PHPExtensions := make(map[string]struct{})

	for _, ext := range strings.Split(string(output), "\n") {
		if trimmed := strings.TrimSpace(ext); trimmed != "" {
			PHPExtensions[trimmed] = struct{}{}
		}
	}

	return PHPExtensions, nil
}

// checkPHP84OrHigher CHECKS IF THE PHP VERSION IS 8.4 OR HIGHER.
func checkPHP84OrHigher(phpVersion string) bool {
	parts := strings.Split(phpVersion, ".")

	if len(parts) < 2 {
		return false
	}

	major, err := strconv.Atoi(parts[0])

	if err != nil {
		return false
	}

	minor, err := strconv.Atoi(parts[1])

	if err != nil {
		return false
	}

	return major > 8 || (major == 8 && minor >= 4)
}
