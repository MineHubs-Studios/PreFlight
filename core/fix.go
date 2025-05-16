package core

import (
	"PreFlight/modules"
	"PreFlight/pm"
	"PreFlight/utils"
	"context"
	"fmt"
	"os"
	"os/exec"
)

// FixDependencies INSTALL MISSING DEPENDENCIES FOR PHP (Composer) AND JS (NPM).
func FixDependencies(ctx context.Context, force bool) {
	ow := utils.NewOutputWriter()

	if !ow.Println(utils.Bold + utils.Blue + "\n╭─────────────────────────────────────────╮" + utils.Reset) {
		return
	}

	if !ow.Println(utils.Bold + utils.Blue + "│" + utils.Red + utils.WarningSign + " " +
		"Be careful this is a experimental feature, which means is not stable yet! Continue on your own risk." + utils.Reset) {
		return
	}

	if !ow.Println(utils.Bold + utils.Blue + "│" + utils.Cyan + utils.Bold + "  🚀 Fixing dependencies  " + utils.Reset) {
		return
	}

	if !ow.Println(utils.Bold + utils.Blue + "╰─────────────────────────────────────────╯" + utils.Reset) {
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
		fmt.Println(utils.WarningSign + " Composer not found. Skipping PHP dependency fix.")
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
		fmt.Printf(utils.CrossMark+" Composer installation failed: %v\n", err)
	} else {
		fmt.Println(utils.CheckMark + " Composer dependencies fixed!")
	}
}

// fixJSDependencies HANDLES INSTALLING MISSING JavaScript/TypeScript DEPENDENCIES.
func fixJSDependencies(ctx context.Context, force bool) {
	packageConfig := pm.LoadPackageConfig()

	if !packageConfig.HasConfig {
		fmt.Println(utils.WarningSign + " package.json not found. Skipping JavaScript dependency fix.")
		return
	}

	packageManager := utils.DetectPackageManager("package")

	fmt.Printf("🛠 Detected package manager: %s. Running `%s install`...\n", packageManager.Command, packageManager.Command)

	args := []string{"install"}

	if force {
		args = append(args, "--force")
	}

	cmd := exec.CommandContext(ctx, packageManager.Command, args...) //nolint:gosec
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf(utils.CrossMark+" %s installation failed: %v\n", packageManager.Command, err)
	} else {
		fmt.Printf(utils.CheckMark+" %s dependencies fixed!\n", packageManager.Command)
	}
}
