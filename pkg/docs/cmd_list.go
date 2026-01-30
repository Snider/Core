package docs

import (
	"fmt"
	"strings"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/spf13/cobra"
)

// Flag variable for list command
var docsListRegistryPath string

var docsListCmd = &cobra.Command{
	Use:   "list",
	Short: i18n.T("cmd.docs.list.short"),
	Long:  i18n.T("cmd.docs.list.long"),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDocsList(docsListRegistryPath)
	},
}

func init() {
	docsListCmd.Flags().StringVar(&docsListRegistryPath, "registry", "", i18n.T("common.flag.registry"))
}

func runDocsList(registryPath string) error {
	reg, _, err := loadRegistry(registryPath)
	if err != nil {
		return err
	}

	fmt.Printf("\n%-20s  %-8s  %-8s  %-10s  %s\n",
		headerStyle.Render(i18n.T("common.label.repo")),
		headerStyle.Render(i18n.T("cmd.docs.list.header.readme")),
		headerStyle.Render(i18n.T("cmd.docs.list.header.claude")),
		headerStyle.Render(i18n.T("cmd.docs.list.header.changelog")),
		headerStyle.Render(i18n.T("cmd.docs.list.header.docs")),
	)
	fmt.Println(strings.Repeat("─", 70))

	var withDocs, withoutDocs int
	for _, repo := range reg.List() {
		info := scanRepoDocs(repo)

		readme := cli.CheckMark(info.Readme != "")
		claude := cli.CheckMark(info.ClaudeMd != "")
		changelog := cli.CheckMark(info.Changelog != "")

		docsDir := cli.CheckMark(false)
		if len(info.DocsFiles) > 0 {
			docsDir = docsFoundStyle.Render(i18n.T("common.count.files", map[string]interface{}{"Count": len(info.DocsFiles)}))
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
	fmt.Printf("%s %s\n",
		cli.Label(i18n.T("common.label.coverage")),
		i18n.T("cmd.docs.list.coverage_summary", map[string]interface{}{"WithDocs": withDocs, "WithoutDocs": withoutDocs}),
	)

	return nil
}
