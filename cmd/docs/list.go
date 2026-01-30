package docs

import (
	"fmt"
	"strings"

	"github.com/leaanthony/clir"
)

func addDocsListCommand(parent *clir.Command) {
	var registryPath string

	listCmd := parent.NewSubCommand("list", "List documentation across repos")
	listCmd.StringFlag("registry", "Path to repos.yaml", &registryPath)

	listCmd.Action(func() error {
		return runDocsList(registryPath)
	})
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

		readme := docsMissingStyle.Render("—")
		if info.Readme != "" {
			readme = docsFoundStyle.Render("✓")
		}

		claude := docsMissingStyle.Render("—")
		if info.ClaudeMd != "" {
			claude = docsFoundStyle.Render("✓")
		}

		changelog := docsMissingStyle.Render("—")
		if info.Changelog != "" {
			changelog = docsFoundStyle.Render("✓")
		}

		docsDir := docsMissingStyle.Render("—")
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
		dimStyle.Render("Coverage:"),
		withDocs,
		withoutDocs,
	)

	return nil
}
