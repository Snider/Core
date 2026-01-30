package php

import (
	"fmt"
	"os"

	phppkg "github.com/host-uk/core/pkg/php"
	"github.com/spf13/cobra"
)

func addPHPPackagesCommands(parent *cobra.Command) {
	packagesCmd := &cobra.Command{
		Use:   "packages",
		Short: "Manage local PHP packages",
		Long: "Link and manage local PHP packages for development.\n\n" +
			"Similar to npm link, this adds path repositories to composer.json\n" +
			"for developing packages alongside your project.\n\n" +
			"Commands:\n" +
			"  link    - Link local packages by path\n" +
			"  unlink  - Unlink packages by name\n" +
			"  update  - Update linked packages\n" +
			"  list    - List linked packages",
	}
	parent.AddCommand(packagesCmd)

	addPHPPackagesLinkCommand(packagesCmd)
	addPHPPackagesUnlinkCommand(packagesCmd)
	addPHPPackagesUpdateCommand(packagesCmd)
	addPHPPackagesListCommand(packagesCmd)
}

func addPHPPackagesLinkCommand(parent *cobra.Command) {
	linkCmd := &cobra.Command{
		Use:   "link [paths...]",
		Short: "Link local packages",
		Long: "Link local PHP packages for development.\n\n" +
			"Adds path repositories to composer.json with symlink enabled.\n" +
			"The package name is auto-detected from each path's composer.json.\n\n" +
			"Examples:\n" +
			"  core php packages link ../my-package\n" +
			"  core php packages link ../pkg-a ../pkg-b",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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
		},
	}

	parent.AddCommand(linkCmd)
}

func addPHPPackagesUnlinkCommand(parent *cobra.Command) {
	unlinkCmd := &cobra.Command{
		Use:   "unlink [packages...]",
		Short: "Unlink packages",
		Long: "Remove linked packages from composer.json.\n\n" +
			"Removes path repositories by package name.\n\n" +
			"Examples:\n" +
			"  core php packages unlink vendor/my-package\n" +
			"  core php packages unlink vendor/pkg-a vendor/pkg-b",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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
		},
	}

	parent.AddCommand(unlinkCmd)
}

func addPHPPackagesUpdateCommand(parent *cobra.Command) {
	updateCmd := &cobra.Command{
		Use:   "update [packages...]",
		Short: "Update linked packages",
		Long: "Run composer update for linked packages.\n\n" +
			"If no packages specified, updates all packages.\n\n" +
			"Examples:\n" +
			"  core php packages update\n" +
			"  core php packages update vendor/my-package",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get working directory: %w", err)
			}

			fmt.Printf("%s Updating packages...\n\n", dimStyle.Render("PHP:"))

			if err := phppkg.UpdatePackages(cwd, args); err != nil {
				return fmt.Errorf("composer update failed: %w", err)
			}

			fmt.Printf("\n%s Packages updated\n", successStyle.Render("Done:"))
			return nil
		},
	}

	parent.AddCommand(updateCmd)
}

func addPHPPackagesListCommand(parent *cobra.Command) {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List linked packages",
		Long: "List all locally linked packages.\n\n" +
			"Shows package name, path, and version for each linked package.",
		RunE: func(cmd *cobra.Command, args []string) error {
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
		},
	}

	parent.AddCommand(listCmd)
}
