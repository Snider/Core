package php

import (
	"fmt"
	"os"

	phppkg "github.com/host-uk/core/pkg/php"
	"github.com/leaanthony/clir"
)

func addPHPPackagesCommands(parent *clir.Command) {
	packagesCmd := parent.NewSubCommand("packages", "Manage local PHP packages")
	packagesCmd.LongDescription("Link and manage local PHP packages for development.\n\n" +
		"Similar to npm link, this adds path repositories to composer.json\n" +
		"for developing packages alongside your project.\n\n" +
		"Commands:\n" +
		"  link    - Link local packages by path\n" +
		"  unlink  - Unlink packages by name\n" +
		"  update  - Update linked packages\n" +
		"  list    - List linked packages")

	addPHPPackagesLinkCommand(packagesCmd)
	addPHPPackagesUnlinkCommand(packagesCmd)
	addPHPPackagesUpdateCommand(packagesCmd)
	addPHPPackagesListCommand(packagesCmd)
}

func addPHPPackagesLinkCommand(parent *clir.Command) {
	linkCmd := parent.NewSubCommand("link", "Link local packages")
	linkCmd.LongDescription("Link local PHP packages for development.\n\n" +
		"Adds path repositories to composer.json with symlink enabled.\n" +
		"The package name is auto-detected from each path's composer.json.\n\n" +
		"Examples:\n" +
		"  core php packages link ../my-package\n" +
		"  core php packages link ../pkg-a ../pkg-b")

	linkCmd.Action(func() error {
		args := linkCmd.OtherArgs()
		if len(args) == 0 {
			return fmt.Errorf("at least one package path is required")
		}

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}

		fmt.Printf("%s Linking packages...\n\n", dimStyle.Render("PHP:"))

		if err := phppkg.LinkPackages(cwd, args); err != nil {
			return fmt.Errorf("failed to link packages: %w", err)
		}

		fmt.Printf("\n%s Packages linked. Run 'composer update' to install.\n", successStyle.Render("Done:"))
		return nil
	})
}

func addPHPPackagesUnlinkCommand(parent *clir.Command) {
	unlinkCmd := parent.NewSubCommand("unlink", "Unlink packages")
	unlinkCmd.LongDescription("Remove linked packages from composer.json.\n\n" +
		"Removes path repositories by package name.\n\n" +
		"Examples:\n" +
		"  core php packages unlink vendor/my-package\n" +
		"  core php packages unlink vendor/pkg-a vendor/pkg-b")

	unlinkCmd.Action(func() error {
		args := unlinkCmd.OtherArgs()
		if len(args) == 0 {
			return fmt.Errorf("at least one package name is required")
		}

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}

		fmt.Printf("%s Unlinking packages...\n\n", dimStyle.Render("PHP:"))

		if err := phppkg.UnlinkPackages(cwd, args); err != nil {
			return fmt.Errorf("failed to unlink packages: %w", err)
		}

		fmt.Printf("\n%s Packages unlinked. Run 'composer update' to remove.\n", successStyle.Render("Done:"))
		return nil
	})
}

func addPHPPackagesUpdateCommand(parent *clir.Command) {
	updateCmd := parent.NewSubCommand("update", "Update linked packages")
	updateCmd.LongDescription("Run composer update for linked packages.\n\n" +
		"If no packages specified, updates all packages.\n\n" +
		"Examples:\n" +
		"  core php packages update\n" +
		"  core php packages update vendor/my-package")

	updateCmd.Action(func() error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}

		args := updateCmd.OtherArgs()

		fmt.Printf("%s Updating packages...\n\n", dimStyle.Render("PHP:"))

		if err := phppkg.UpdatePackages(cwd, args); err != nil {
			return fmt.Errorf("composer update failed: %w", err)
		}

		fmt.Printf("\n%s Packages updated\n", successStyle.Render("Done:"))
		return nil
	})
}

func addPHPPackagesListCommand(parent *clir.Command) {
	listCmd := parent.NewSubCommand("list", "List linked packages")
	listCmd.LongDescription("List all locally linked packages.\n\n" +
		"Shows package name, path, and version for each linked package.")

	listCmd.Action(func() error {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}

		packages, err := phppkg.ListLinkedPackages(cwd)
		if err != nil {
			return fmt.Errorf("failed to list packages: %w", err)
		}

		if len(packages) == 0 {
			fmt.Printf("%s No linked packages found\n", dimStyle.Render("PHP:"))
			return nil
		}

		fmt.Printf("%s Linked packages:\n\n", dimStyle.Render("PHP:"))

		for _, pkg := range packages {
			name := pkg.Name
			if name == "" {
				name = "(unknown)"
			}
			version := pkg.Version
			if version == "" {
				version = "dev"
			}

			fmt.Printf("  %s %s\n", successStyle.Render("*"), name)
			fmt.Printf("    %s %s\n", dimStyle.Render("Path:"), pkg.Path)
			fmt.Printf("    %s %s\n", dimStyle.Render("Version:"), version)
			fmt.Println()
		}

		return nil
	})
}
