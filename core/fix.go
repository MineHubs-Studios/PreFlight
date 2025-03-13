package core

import (
	"PreFlight/config"
	"PreFlight/modules"
	"PreFlight/utils"
	"context"
	"fmt"
	"os"
	"os/exec"
)

// FixDependencies INSTALL MISSING DEPENDENCIES FOR PHP (Composer) AND JS (NPM).
func FixDependencies(ctx context.Context, force bool) {
	ow := utils.NewOutputWriter()

	if !ow.Println(Bold + Blue + "\n╭─────────────────────────────────────────╮" + Reset) {
		return
	}

	if !ow.Println(Bold + Blue + "│" + Cyan + Bold + "  🚀 Fixing dependencies  " + Reset) {
		return
	}

	if !ow.Println(Bold + Blue + "╰─────────────────────────────────────────╯" + Reset) {
		return
	}

	if !ow.PrintNewLines(1) {
		return
	}

	// FIX COMPOSER DEPENDENCIES.
	if version, err := modules.GetComposerVersion(ctx); err == nil {
		fmt.Printf("🛠 Composer found (version: %s). Running `composer install`...\n", version)

		args := []string{"install"}

		if force {
			args = append(args, "--no-cache")
		}

		cmd := exec.CommandContext(ctx, "composer", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Printf("❌ Composer installation failed: %v\n", err)
		} else {
			fmt.Println("✅ Composer dependencies fixed!")
		}
	} else {
		fmt.Println("⚠️ Composer not found. Skipping PHP dependency fix.")
	}

	// FIX JAVASCRIPT AND TYPESCRIPT DEPENDENCIES.
	packageConfig := config.LoadPackageConfig()

	if packageConfig.HasJSON {
		packageManager := modules.DeterminePackageManager(packageConfig)

		fmt.Printf("🛠 Detected package manager: %s. Running `%s install`...\n", packageManager.Command, packageManager.Command)

		args := []string{"install"}

		if force {
			args = append(args, "--force")
		}

		cmd := exec.CommandContext(ctx, packageManager.Command, args...) //nolint:gosec
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			fmt.Printf("❌ %s installation failed: %v\n", packageManager.Command, err)
		} else {
			fmt.Printf("✅ %s dependencies fixed!\n", packageManager.Command)
		}
	} else {
		fmt.Println("⚠️ package.json not found. Skipping JavaScript dependency fix.")
	}
}
