// github_diff.go provides change tracking for dry-run output.

package setup

import (
	"fmt"
	"sort"
	"strings"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
)

// ChangeType indicates the type of change being made.
type ChangeType string

const (
	ChangeCreate ChangeType = "create"
	ChangeUpdate ChangeType = "update"
	ChangeDelete ChangeType = "delete"
	ChangeSkip   ChangeType = "skip"
)

// ChangeCategory groups changes by type.
type ChangeCategory string

const (
	CategoryLabel      ChangeCategory = "label"
	CategoryWebhook    ChangeCategory = "webhook"
	CategoryProtection ChangeCategory = "protection"
	CategorySecurity   ChangeCategory = "security"
)

// Change represents a single change to be made.
type Change struct {
	Category    ChangeCategory
	Type        ChangeType
	Name        string
	Description string
	Details     map[string]string // Key-value details about the change
}

// ChangeSet tracks all changes for a repository.
type ChangeSet struct {
	Repo    string
	Changes []Change
}

// NewChangeSet creates a new change set for a repository.
func NewChangeSet(repo string) *ChangeSet {
	return &ChangeSet{
		Repo:    repo,
		Changes: make([]Change, 0),
	}
}

// Add adds a change to the set.
func (cs *ChangeSet) Add(category ChangeCategory, changeType ChangeType, name, description string) {
	cs.Changes = append(cs.Changes, Change{
		Category:    category,
		Type:        changeType,
		Name:        name,
		Description: description,
		Details:     make(map[string]string),
	})
}

// AddWithDetails adds a change with additional details.
func (cs *ChangeSet) AddWithDetails(category ChangeCategory, changeType ChangeType, name, description string, details map[string]string) {
	cs.Changes = append(cs.Changes, Change{
		Category:    category,
		Type:        changeType,
		Name:        name,
		Description: description,
		Details:     details,
	})
}

// HasChanges returns true if there are any non-skip changes.
func (cs *ChangeSet) HasChanges() bool {
	for _, c := range cs.Changes {
		if c.Type != ChangeSkip {
			return true
		}
	}
	return false
}

// Count returns the number of changes by type.
func (cs *ChangeSet) Count() (creates, updates, deletes, skips int) {
	for _, c := range cs.Changes {
		switch c.Type {
		case ChangeCreate:
			creates++
		case ChangeUpdate:
			updates++
		case ChangeDelete:
			deletes++
		case ChangeSkip:
			skips++
		}
	}
	return
}

// CountByCategory returns changes grouped by category.
func (cs *ChangeSet) CountByCategory() map[ChangeCategory]int {
	counts := make(map[ChangeCategory]int)
	for _, c := range cs.Changes {
		if c.Type != ChangeSkip {
			counts[c.Category]++
		}
	}
	return counts
}

// Print outputs the change set to the console.
func (cs *ChangeSet) Print(verbose bool) {
	creates, updates, deletes, skips := cs.Count()

	// Print header
	fmt.Printf("\n%s %s\n", dimStyle.Render(i18n.Label("repo")), repoNameStyle.Render(cs.Repo))

	if !cs.HasChanges() {
		fmt.Printf("  %s\n", dimStyle.Render(i18n.T("cmd.setup.github.no_changes")))
		return
	}

	// Print summary
	var parts []string
	if creates > 0 {
		parts = append(parts, successStyle.Render(fmt.Sprintf("+%d", creates)))
	}
	if updates > 0 {
		parts = append(parts, warningStyle.Render(fmt.Sprintf("~%d", updates)))
	}
	if deletes > 0 {
		parts = append(parts, errorStyle.Render(fmt.Sprintf("-%d", deletes)))
	}
	if skips > 0 && verbose {
		parts = append(parts, dimStyle.Render(fmt.Sprintf("=%d", skips)))
	}
	fmt.Printf("  %s\n", strings.Join(parts, " "))

	// Print details if verbose
	if verbose {
		cs.printByCategory(CategoryLabel, "Labels")
		cs.printByCategory(CategoryWebhook, "Webhooks")
		cs.printByCategory(CategoryProtection, "Branch protection")
		cs.printByCategory(CategorySecurity, "Security")
	}
}

func (cs *ChangeSet) printByCategory(category ChangeCategory, title string) {
	var categoryChanges []Change
	for _, c := range cs.Changes {
		if c.Category == category && c.Type != ChangeSkip {
			categoryChanges = append(categoryChanges, c)
		}
	}

	if len(categoryChanges) == 0 {
		return
	}

	fmt.Printf("\n  %s:\n", dimStyle.Render(title))
	for _, c := range categoryChanges {
		icon := getChangeIcon(c.Type)
		style := getChangeStyle(c.Type)
		fmt.Printf("    %s %s", style.Render(icon), c.Name)
		if c.Description != "" {
			fmt.Printf(" %s", dimStyle.Render(c.Description))
		}
		fmt.Println()

		// Print details (sorted for deterministic output)
		keys := make([]string, 0, len(c.Details))
		for k := range c.Details {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Printf("      %s: %s\n", dimStyle.Render(k), c.Details[k])
		}
	}
}

func getChangeIcon(t ChangeType) string {
	switch t {
	case ChangeCreate:
		return "+"
	case ChangeUpdate:
		return "~"
	case ChangeDelete:
		return "-"
	default:
		return "="
	}
}

func getChangeStyle(t ChangeType) *cli.AnsiStyle {
	switch t {
	case ChangeCreate:
		return successStyle
	case ChangeUpdate:
		return warningStyle
	case ChangeDelete:
		return errorStyle
	default:
		return dimStyle
	}
}

// Aggregate combines multiple change sets into a summary.
type Aggregate struct {
	Sets []*ChangeSet
}

// NewAggregate creates a new aggregate.
func NewAggregate() *Aggregate {
	return &Aggregate{
		Sets: make([]*ChangeSet, 0),
	}
}

// Add adds a change set to the aggregate.
func (a *Aggregate) Add(cs *ChangeSet) {
	a.Sets = append(a.Sets, cs)
}

// TotalChanges returns the total number of changes across all sets.
func (a *Aggregate) TotalChanges() (creates, updates, deletes, skips int) {
	for _, cs := range a.Sets {
		c, u, d, s := cs.Count()
		creates += c
		updates += u
		deletes += d
		skips += s
	}
	return
}

// ReposWithChanges returns the number of repos that have changes.
func (a *Aggregate) ReposWithChanges() int {
	count := 0
	for _, cs := range a.Sets {
		if cs.HasChanges() {
			count++
		}
	}
	return count
}

// PrintSummary outputs the aggregate summary.
func (a *Aggregate) PrintSummary() {
	creates, updates, deletes, _ := a.TotalChanges()
	reposWithChanges := a.ReposWithChanges()

	fmt.Println()
	fmt.Printf("%s\n", dimStyle.Render(i18n.Label("summary")))
	fmt.Printf("  %s: %d\n", i18n.T("cmd.setup.github.repos_checked"), len(a.Sets))

	if reposWithChanges == 0 {
		fmt.Printf("  %s\n", dimStyle.Render(i18n.T("cmd.setup.github.all_up_to_date")))
		return
	}

	fmt.Printf("  %s: %d\n", i18n.T("cmd.setup.github.repos_with_changes"), reposWithChanges)
	if creates > 0 {
		fmt.Printf("  %s: %s\n", i18n.T("cmd.setup.github.to_create"), successStyle.Render(fmt.Sprintf("%d", creates)))
	}
	if updates > 0 {
		fmt.Printf("  %s: %s\n", i18n.T("cmd.setup.github.to_update"), warningStyle.Render(fmt.Sprintf("%d", updates)))
	}
	if deletes > 0 {
		fmt.Printf("  %s: %s\n", i18n.T("cmd.setup.github.to_delete"), errorStyle.Render(fmt.Sprintf("%d", deletes)))
	}
}
