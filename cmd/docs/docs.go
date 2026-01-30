// Package docs provides documentation management commands.
package docs

import (
	"github.com/host-uk/core/cmd/shared"
	"github.com/spf13/cobra"
)

// Style and utility aliases from shared
var (
	repoNameStyle    = shared.RepoNameStyle
	successStyle     = shared.SuccessStyle
	errorStyle       = shared.ErrorStyle
	dimStyle         = shared.DimStyle
	headerStyle      = shared.HeaderStyle
	confirm          = shared.Confirm
	docsFoundStyle   = shared.SuccessStyle
	docsMissingStyle = shared.DimStyle
	docsFileStyle    = shared.InfoStyle
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Documentation management",
	Long: `Manage documentation across all repos.
Scan for docs, check coverage, and sync to core-php/docs/packages/.`,
}

func init() {
	docsCmd.AddCommand(docsSyncCmd)
	docsCmd.AddCommand(docsListCmd)
}
