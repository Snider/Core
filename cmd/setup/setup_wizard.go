// setup_wizard.go implements the interactive package selection wizard.
//
// Uses charmbracelet/huh for a rich terminal UI with multi-select checkboxes.
// Falls back to non-interactive mode when not in a TTY or --all is specified.

package setup

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/host-uk/core/pkg/repos"
	"golang.org/x/term"
)

// wizardTheme returns a Dracula-inspired theme matching our CLI styling.
func wizardTheme() *huh.Theme {
	t := huh.ThemeDracula()
	return t
}

// isTerminal returns true if stdin is a terminal.
func isTerminal() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// promptSetupChoice asks the user whether to setup the working directory or create a package.
func promptSetupChoice() (string, error) {
	var choice string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(i18n.T("cmd.setup.wizard.git_repo_title")).
				Description(i18n.T("cmd.setup.wizard.what_to_do")).
				Options(
					huh.NewOption(i18n.T("cmd.setup.wizard.option_setup"), "setup").Selected(true),
					huh.NewOption(i18n.T("cmd.setup.wizard.option_package"), "package"),
				).
				Value(&choice),
		),
	).WithTheme(wizardTheme())

	if err := form.Run(); err != nil {
		return "", err
	}

	return choice, nil
}

// promptProjectName asks the user for a project directory name.
func promptProjectName(defaultName string) (string, error) {
	var name string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(i18n.T("cmd.setup.wizard.project_name_title")).
				Description(i18n.T("cmd.setup.wizard.project_name_desc")).
				Placeholder(defaultName).
				Value(&name),
		),
	).WithTheme(wizardTheme())

	if err := form.Run(); err != nil {
		return "", err
	}

	if name == "" {
		return defaultName, nil
	}
	return name, nil
}

// groupPackagesByType organizes repos by their type for display.
func groupPackagesByType(reposList []*repos.Repo) map[string][]*repos.Repo {
	groups := make(map[string][]*repos.Repo)

	for _, repo := range reposList {
		repoType := repo.Type
		if repoType == "" {
			repoType = "other"
		}
		groups[repoType] = append(groups[repoType], repo)
	}

	// Sort within each group
	for _, group := range groups {
		sort.Slice(group, func(i, j int) bool {
			return group[i].Name < group[j].Name
		})
	}

	return groups
}

// packageOption represents a selectable package in the wizard.
type packageOption struct {
	repo     *repos.Repo
	selected bool
}

// runPackageWizard presents an interactive multi-select UI for package selection.
// Returns the list of selected repo names.
func runPackageWizard(reg *repos.Registry, preselectedTypes []string) ([]string, error) {
	allRepos := reg.List()

	// Build preselection set
	preselect := make(map[string]bool)
	for _, t := range preselectedTypes {
		preselect[strings.TrimSpace(t)] = true
	}

	// Group repos by type for organized display
	groups := groupPackagesByType(allRepos)

	// Build options with preselection
	var options []huh.Option[string]
	typeOrder := []string{"foundation", "module", "product", "template", "other"}

	for _, typeKey := range typeOrder {
		group, ok := groups[typeKey]
		if !ok || len(group) == 0 {
			continue
		}

		// Add type header as a visual separator (empty option)
		typeLabel := strings.ToUpper(typeKey)
		options = append(options, huh.NewOption[string](
			fmt.Sprintf("── %s ──", typeLabel),
			"",
		).Selected(false))

		for _, repo := range group {
			// Skip if clone: false
			if repo.Clone != nil && !*repo.Clone {
				continue
			}

			label := repo.Name
			if repo.Description != "" {
				label = fmt.Sprintf("%s - %s", repo.Name, truncateDesc(repo.Description, 40))
			}

			// Preselect based on type filter or select all if no filter
			selected := len(preselect) == 0 || preselect[repo.Type]

			options = append(options, huh.NewOption[string](label, repo.Name).Selected(selected))
		}
	}

	var selected []string

	// Header styling
	headerStyle := cli.TitleStyle.MarginBottom(1)

	fmt.Println(headerStyle.Render(i18n.T("cmd.setup.wizard.package_selection")))
	fmt.Println(i18n.T("cmd.setup.wizard.selection_hint"))
	fmt.Println()

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title(i18n.T("cmd.setup.wizard.select_packages")).
				Options(options...).
				Value(&selected).
				Filterable(true).
				Height(20),
		),
	).WithTheme(wizardTheme())

	if err := form.Run(); err != nil {
		return nil, err
	}

	// Filter out empty values (type headers)
	var result []string
	for _, name := range selected {
		if name != "" {
			result = append(result, name)
		}
	}

	return result, nil
}

// confirmClone asks for confirmation before cloning.
func confirmClone(count int, target string) (bool, error) {
	var confirmed bool

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(i18n.T("cmd.setup.wizard.confirm_clone", map[string]interface{}{"Count": count, "Target": target})).
				Affirmative(i18n.T("cmd.setup.wizard.confirm_yes")).
				Negative(i18n.T("cmd.setup.wizard.confirm_cancel")).
				Value(&confirmed),
		),
	).WithTheme(wizardTheme())

	if err := form.Run(); err != nil {
		return false, err
	}

	return confirmed, nil
}

// truncateDesc truncates a description to max length with ellipsis.
func truncateDesc(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
