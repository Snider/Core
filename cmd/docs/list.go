package docs

import (
	"fmt"
	"strings"

	"github.com/host-uk/core/cmd/shared"
	"github.com/spf13/cobra"
)

// Flag variable for list command
var docsListRegistryPath string

var docsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List documentation across repos",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDocsList(docsListRegistryPath)
	},
}

func init() {
	docsListCmd.Flags().StringVar(&docsListRegistryPath, "registry", "", "Path to repos.yaml")
}

func runDocsList(registryPath string) error {
	reg, _, err := loadRegistry(registryPath)
	if err != nil {
		return err
	}

	fmt.Printf("\n%-20s  %-8s  %-8s  %-10s  %s\n",
		headerStyle.Render("Repo"),
		headerStyle.Render("README"),
		headerStyle.Render("CLAUDE"),
		headerStyle.Render("CHANGELOG"),
		headerStyle.Render("docs/"),
	)
	fmt.Println(strings.Repeat("─", 70))

	var withDocs, withoutDocs int
	for _, repo := range reg.List() {
		info := scanRepoDocs(repo)

		readme := shared.CheckMark(info.Readme != "")
		claude := shared.CheckMark(info.ClaudeMd != "")
		changelog := shared.CheckMark(info.Changelog != "")

		docsDir := shared.CheckMark(false)
		if len(info.DocsFiles) > 0 {
			docsDir = docsFoundStyle.Render(fmt.Sprintf("%d files", len(info.DocsFiles)))
		}

		fmt.Printf("%-20s  %-8s  %-8s  %-10s  %s\n",
			repoNameStyle.Render(info.Name),
			readme,
			claude,
			changelog,
			docsDir,
		)

		if info.HasDocs {
			withDocs++
		} else {
			withoutDocs++
		}
	}

	fmt.Println()
	fmt.Printf("%s %d with docs, %d without\n",
		shared.Label("Coverage"),
		withDocs,
		withoutDocs,
	)

	return nil
}
