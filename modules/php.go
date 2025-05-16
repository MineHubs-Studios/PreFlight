package modules

import (
	"PreFlight/pm"
	"PreFlight/utils"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
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

	isPieInstalled := false
	if checkPHP84OrHigher(phpVersion) {
		isPieInstalled, _ = checkPieInstalled(ctx)
	}

	composerConfig := pm.LoadComposerConfig()

	if composerConfig.Error != nil {
		errors = append(errors, fmt.Sprintf("Failed to read composer.json: %v", composerConfig.Error))
		return errors, warnings, successes
	}

	// VALIDATE PHP VERSION.
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

	// Get PHP extensions
	installedExtensions, err := getPhpExtensions(ctx)
	if err != nil {
		errors = append(errors, fmt.Sprintf("Failed to check PHP extensions: %v", err))
		return errors, warnings, successes
	}

	// Get Pie extensions if available
	pieExtensions := make(map[string]struct{})
	if isPieInstalled {
		pieExtensions, _ = getPieExtensionsFromPhar(ctx)
	}

	// Create combined map for validation, but keep track of extension source
	extensionSources := make(map[string]string) // maps extension name to source ("php" or "pie")
	for ext := range installedExtensions {
		extensionSources[ext] = "php"
	}
	for ext := range pieExtensions {
		extensionSources[ext] = "pie" // Overwrite if it exists in both
	}

	deprecatedExtensions := map[string]struct{}{
		"imap": {}, "mysql": {}, "recode": {}, "statistics": {}, "wddx": {}, "xml-rpc": {},
	}

	experimentalExtensions := map[string]struct{}{
		"gmagick": {}, "imagemagick": {}, "mqseries": {}, "parle": {}, "rnp": {}, "svm": {}, "svn": {}, "ui": {}, "omq": {},
	}

	// Track extensions to display
	type ExtensionInfo struct {
		Name      string
		Source    string // "php" or "pie"
		IsWarning bool
		Warning   string
	}

	extensionsToShow := make([]ExtensionInfo, 0)

	// 1. Add all Pie extensions
	for ext := range pieExtensions {
		if ext == "" || ext == "Core" || ext == "standard" ||
			ext == "[PHP Modules]" || ext == "[Zend Modules]" {
			continue // Skip empty or header entries
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

	// 2. Add required PHP extensions from composer.json
	if len(composerConfig.PHPExtensions) > 0 {
		for _, ext := range composerConfig.PHPExtensions {
			// Skip if it's already included as a Pie extension
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

			// Check if extension is installed
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

			// Handle PHP 8.4+ split extensions
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

			// Extension is missing
			errors = append(errors, fmt.Sprintf("Missing extension %s%s, Please enable it.", utils.Reset, ext))

		NextExtension:
		}
	}

	// Sort extensions alphabetically by name
	sort.Slice(extensionsToShow, func(i, j int) bool {
		return strings.ToLower(extensionsToShow[i].Name) < strings.ToLower(extensionsToShow[j].Name)
	})

	// Generate extension feedback messages
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

// checkPieInstalled CHECKS IF PIE IS INSTALLED AND AVAILABLE
func checkPieInstalled(ctx context.Context) (bool, error) {
	// Check if pie command exists
	cmd := exec.CommandContext(ctx, "pie", "--version")
	_, err := cmd.Output()
	if err == nil {
		return true, nil
	}

	// Also check if pie.phar file exists in common locations
	searchPaths := []string{
		"./pie.phar",
		"/usr/local/bin/pie.phar",
		"/usr/bin/pie.phar",
	}

	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			return true, nil
		}
	}

	return false, fmt.Errorf("pie not installed or pie.phar not found: %w", err)
}

// findPiePharPath attempts to locate pie.phar in common locations
func findPiePharPath() (string, error) {
	searchPaths := []string{
		"./pie.phar",
		"/usr/local/bin/pie.phar",
		"/usr/bin/pie.phar",
	}

	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	// Try to find using 'which pie' command
	cmd := exec.Command("which", "pie")
	output, err := cmd.Output()
	if err == nil {
		path := strings.TrimSpace(string(output))
		if path != "" {
			// Check if this points to a pie.phar file
			if filepath.Ext(path) == ".phar" {
				return path, nil
			}

			// Check if there's a pie.phar in the same directory
			dir := filepath.Dir(path)
			pharPath := filepath.Join(dir, "pie.phar")
			if _, err := os.Stat(pharPath); err == nil {
				return pharPath, nil
			}
		}
	}

	return "", fmt.Errorf("could not find pie.phar")
}

// getPieExtensionsFromPhar directly extracts extension info from pie.phar
func getPieExtensionsFromPhar(ctx context.Context) (map[string]struct{}, error) {
	extensions := make(map[string]struct{})

	// Method 1: Try using pie -m command first
	cmd := exec.CommandContext(ctx, "pie", "-m")
	output, err := cmd.Output()
	if err == nil {
		for _, ext := range strings.Split(string(output), "\n") {
			if trimmed := strings.TrimSpace(ext); trimmed != "" {
				extensions[trimmed] = struct{}{}
			}
		}
	}

	// Method 2: Use PHP to look inside the pie.phar file (critical for finding xdebug)
	pharPath, pathErr := findPiePharPath()
	if pathErr == nil {
		// This command uses PHP to extract extension info from inside the phar file
		phpCmd := exec.CommandContext(ctx, "php", "-r", fmt.Sprintf(`
			try {
				$phar = new Phar('%s');
				$manifest = $phar->getMetadata();
				if (isset($manifest['extensions'])) {
					echo implode("\n", $manifest['extensions']);
				}
				// Check specifically for xdebug
				if (isset($manifest['xdebug']) || isset($manifest['extensions']['xdebug'])) {
					echo "\nxdebug";
				}
			} catch (Exception $e) {
				exit(1);
			}
		`, pharPath))

		phpOutput, phpErr := phpCmd.Output()
		if phpErr == nil {
			for _, ext := range strings.Split(string(phpOutput), "\n") {
				if trimmed := strings.TrimSpace(ext); trimmed != "" {
					extensions[trimmed] = struct{}{}
				}
			}
		}

		// Method 3: Direct file parsing for extension directories in the PHAR
		// This specifically checks for xdebug and other extension directories
		pharCmd := exec.CommandContext(ctx, "php", "-r", fmt.Sprintf(`
			try {
				$phar = new Phar('%s');
				foreach (new RecursiveIteratorIterator($phar) as $file) {
					$path = $file->getPathname();
					if (strpos($path, 'xdebug') !== false) {
						echo "xdebug\n";
					}
					// Look for other extension pattern directories
					if (preg_match('/extensions\/([a-zA-Z0-9_-]+)\//', $path, $matches)) {
						echo $matches[1] . "\n";
					}
				}
			} catch (Exception $e) {
				exit(1);
			}
		`, pharPath))

		pharOutput, pharErr := pharCmd.Output()
		if pharErr == nil {
			for _, ext := range strings.Split(string(pharOutput), "\n") {
				if trimmed := strings.TrimSpace(ext); trimmed != "" {
					extensions[trimmed] = struct{}{}
				}
			}
		}
	}

	// Explicitly check for common extensions that might be in the PHAR
	commonPieExtensions := []string{"xdebug", "opcache", "pcov"}
	pharPath, _ = findPiePharPath()
	if pharPath != "" {
		for _, ext := range commonPieExtensions {
			// Check if extension directories exist in the PHAR
			checkCmd := exec.CommandContext(ctx, "php", "-r", fmt.Sprintf(`
				try {
					$phar = new Phar('%s');
					if ($phar->offsetExists('extensions/%s') ||
						$phar->offsetExists('ext/%s') ||
						$phar->offsetExists('%s')) {
						echo "found";
					}
				} catch (Exception $e) {
					exit(1);
				}
			`, pharPath, ext, ext, ext))

			checkOutput, checkErr := checkCmd.Output()
			if checkErr == nil && strings.Contains(string(checkOutput), "found") {
				extensions[ext] = struct{}{}
			}
		}
	}

	return extensions, nil
}
