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

	fixComposerDependencies(ctx, force)
	fixJSDependencies(ctx, force)
}

// fixComposerDependencies HANDLES INSTALLING MISSING Composer DEPENDENCIES.
func fixComposerDependencies(ctx context.Context, force bool) {
	version, err := modules.GetComposerVersion(ctx)
	if err != nil {
		fmt.Println("⚠️ Composer not found. Skipping PHP dependency fix.")
		return
	}

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
}

// fixJSDependencies HANDLES INSTALLING MISSING JavaScript/TypeScript DEPENDENCIES.
func fixJSDependencies(ctx context.Context, force bool) {
	packageConfig := config.LoadPackageConfig()

	if !packageConfig.HasJSON {
		fmt.Println("⚠️ package.json not found. Skipping JavaScript dependency fix.")
		return
	}

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
}
