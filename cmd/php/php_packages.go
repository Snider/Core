package php

import (
	"fmt"
	"os"

	"github.com/host-uk/core/pkg/i18n"
	phppkg "github.com/host-uk/core/pkg/php"
	"github.com/spf13/cobra"
)

func addPHPPackagesCommands(parent *cobra.Command) {
	packagesCmd := &cobra.Command{
		Use:   "packages",
		Short: i18n.T("cmd.php.packages.short"),
		Long:  i18n.T("cmd.php.packages.long"),
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
		Short: i18n.T("cmd.php.packages.link.short"),
		Long:  i18n.T("cmd.php.packages.link.long"),
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("%s: %w", i18n.T("common.error.failed", map[string]any{"Action": "get working directory"}), err)
			}

			fmt.Printf("%s %s\n\n", dimStyle.Render(i18n.T("cmd.php.label.php")), i18n.T("cmd.php.packages.link.linking"))

			if err := phppkg.LinkPackages(cwd, args); err != nil {
				return fmt.Errorf("%s: %w", i18n.T("common.error.failed", map[string]any{"Action": "link packages"}), err)
			}

			fmt.Printf("\n%s %s\n", successStyle.Render(i18n.T("common.label.done")), i18n.T("cmd.php.packages.link.done"))
			return nil
		},
	}

	parent.AddCommand(linkCmd)
}

func addPHPPackagesUnlinkCommand(parent *cobra.Command) {
	unlinkCmd := &cobra.Command{
		Use:   "unlink [packages...]",
		Short: i18n.T("cmd.php.packages.unlink.short"),
		Long:  i18n.T("cmd.php.packages.unlink.long"),
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("%s: %w", i18n.T("common.error.failed", map[string]any{"Action": "get working directory"}), err)
			}

			fmt.Printf("%s %s\n\n", dimStyle.Render(i18n.T("cmd.php.label.php")), i18n.T("cmd.php.packages.unlink.unlinking"))

			if err := phppkg.UnlinkPackages(cwd, args); err != nil {
				return fmt.Errorf("%s: %w", i18n.T("common.error.failed", map[string]any{"Action": "unlink packages"}), err)
			}

			fmt.Printf("\n%s %s\n", successStyle.Render(i18n.T("common.label.done")), i18n.T("cmd.php.packages.unlink.done"))
			return nil
		},
	}

	parent.AddCommand(unlinkCmd)
}

func addPHPPackagesUpdateCommand(parent *cobra.Command) {
	updateCmd := &cobra.Command{
		Use:   "update [packages...]",
		Short: i18n.T("cmd.php.packages.update.short"),
		Long:  i18n.T("cmd.php.packages.update.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("%s: %w", i18n.T("common.error.failed", map[string]any{"Action": "get working directory"}), err)
			}

			fmt.Printf("%s %s\n\n", dimStyle.Render(i18n.T("cmd.php.label.php")), i18n.T("cmd.php.packages.update.updating"))

			if err := phppkg.UpdatePackages(cwd, args); err != nil {
				return fmt.Errorf("%s: %w", i18n.T("cmd.php.error.update_packages"), err)
			}

			fmt.Printf("\n%s %s\n", successStyle.Render(i18n.T("common.label.done")), i18n.T("cmd.php.packages.update.done"))
			return nil
		},
	}

	parent.AddCommand(updateCmd)
}

func addPHPPackagesListCommand(parent *cobra.Command) {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: i18n.T("cmd.php.packages.list.short"),
		Long:  i18n.T("cmd.php.packages.list.long"),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("%s: %w", i18n.T("common.error.failed", map[string]any{"Action": "get working directory"}), err)
			}

			packages, err := phppkg.ListLinkedPackages(cwd)
			if err != nil {
				return fmt.Errorf("%s: %w", i18n.T("common.error.failed", map[string]any{"Action": "list packages"}), err)
			}

			if len(packages) == 0 {
				fmt.Printf("%s %s\n", dimStyle.Render(i18n.T("cmd.php.label.php")), i18n.T("cmd.php.packages.list.none_found"))
				return nil
			}

			fmt.Printf("%s %s\n\n", dimStyle.Render(i18n.T("cmd.php.label.php")), i18n.T("cmd.php.packages.list.linked"))

			for _, pkg := range packages {
				name := pkg.Name
				if name == "" {
					name = i18n.T("cmd.php.packages.list.unknown")
				}
				version := pkg.Version
				if version == "" {
					version = "dev"
				}

				fmt.Printf("  %s %s\n", successStyle.Render("*"), name)
				fmt.Printf("    %s %s\n", dimStyle.Render(i18n.T("common.label.path")), pkg.Path)
				fmt.Printf("    %s %s\n", dimStyle.Render(i18n.T("common.label.version")), version)
				fmt.Println()
			}

			return nil
		},
	}

	parent.AddCommand(listCmd)
}
