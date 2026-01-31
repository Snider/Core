// cmd_wizard.go implements the interactive package selection wizard.
package setup

import (
	"fmt"
	"os"
	"sort"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/repos"
	"golang.org/x/term"
)

// isTerminal returns true if stdin is a terminal.
func isTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// promptSetupChoice asks the user whether to setup the working directory or create a package.
func promptSetupChoice() (string, error) {
	fmt.Println(cli.TitleStyle.Render(i18n.T("cmd.setup.wizard.git_repo_title")))
	fmt.Println(i18n.T("cmd.setup.wizard.what_to_do"))

	choice, err := cli.Select("Choose action", []string{"setup", "package"})
	if err != nil {
		return "", err
	}
	return choice, nil
}

// promptProjectName asks the user for a project directory name.
func promptProjectName(defaultName string) (string, error) {
	fmt.Println(cli.TitleStyle.Render(i18n.T("cmd.setup.wizard.project_name_title")))
	return cli.Prompt(i18n.T("cmd.setup.wizard.project_name_desc"), defaultName)
}

// runPackageWizard presents an interactive multi-select UI for package selection.
func runPackageWizard(reg *repos.Registry, preselectedTypes []string) ([]string, error) {
	allRepos := reg.List()

	// Build options
	var options []string
	
	// Sort by name
	sort.Slice(allRepos, func(i, j int) bool {
		return allRepos[i].Name < allRepos[j].Name
	})

	for _, repo := range allRepos {
		if repo.Clone != nil && !*repo.Clone {
			continue
		}
		// Format: name (type)
		label := fmt.Sprintf("%s (%s)", repo.Name, repo.Type)
		options = append(options, label)
	}

	fmt.Println(cli.TitleStyle.Render(i18n.T("cmd.setup.wizard.package_selection")))
	fmt.Println(i18n.T("cmd.setup.wizard.selection_hint"))

	selectedLabels, err := cli.MultiSelect(i18n.T("cmd.setup.wizard.select_packages"), options)
	if err != nil {
		return nil, err
	}

	// Extract names from labels
	var selected []string
	for _, label := range selectedLabels {
		// Basic parsing assuming "name (type)" format
		// Find last space
		var name string
		// Since we constructed it, we know it ends with (type)
		// but repo name might have spaces? Repos usually don't.
		// Let's iterate repos to find match
		for _, repo := range allRepos {
			if label == fmt.Sprintf("%s (%s)", repo.Name, repo.Type) {
				name = repo.Name
				break
			}
		}
		if name != "" {
			selected = append(selected, name)
		}
	}
	return selected, nil
}

// confirmClone asks for confirmation before cloning.
func confirmClone(count int, target string) (bool, error) {
	confirmed := cli.Confirm(i18n.T("cmd.setup.wizard.confirm_clone", map[string]interface{}{"Count": count, "Target": target}))
	return confirmed, nil
}