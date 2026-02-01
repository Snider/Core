package docs

import (
	"strings"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
)

// Flag variable for list command
var docsListRegistryPath string

var docsListCmd = &cli.Command{
	Use:   "list",
	Short: i18n.T("cmd.docs.list.short"),
	Long:  i18n.T("cmd.docs.list.long"),
	RunE: func(cmd *cli.Command, args []string) error {
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

	cli.Print("\n%-20s  %-8s  %-8s  %-10s  %s\n",
		headerStyle.Render(i18n.Label("repo")),
		headerStyle.Render(i18n.T("cmd.docs.list.header.readme")),
		headerStyle.Render(i18n.T("cmd.docs.list.header.claude")),
		headerStyle.Render(i18n.T("cmd.docs.list.header.changelog")),
		headerStyle.Render(i18n.T("cmd.docs.list.header.docs")),
	)
	cli.Text(strings.Repeat("─", 70))

	var withDocs, withoutDocs int
	for _, repo := range reg.List() {
		info := scanRepoDocs(repo)

		readme := checkMark(info.Readme != "")
		claude := checkMark(info.ClaudeMd != "")
		changelog := checkMark(info.Changelog != "")

		docsDir := checkMark(false)
		if len(info.DocsFiles) > 0 {
			docsDir = docsFoundStyle.Render(i18n.T("common.count.files", map[string]interface{}{"Count": len(info.DocsFiles)}))
		}

		cli.Print("%-20s  %-8s  %-8s  %-10s  %s\n",
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

	cli.Blank()
	cli.Print("%s %s\n",
		cli.KeyStyle.Render(i18n.Label("coverage")),
		i18n.T("cmd.docs.list.coverage_summary", map[string]interface{}{"WithDocs": withDocs, "WithoutDocs": withoutDocs}),
	)

	return nil
}

func checkMark(ok bool) string {
	if ok {
		return cli.Glyph(":check:")
	}
	return cli.Glyph(":cross:")
}
