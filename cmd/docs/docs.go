// Package docs provides documentation management commands.
package docs

import (
	"github.com/host-uk/core/cmd/shared"
	"github.com/host-uk/core/pkg/i18n"
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
	Short: i18n.T("cmd.docs.short"),
	Long:  i18n.T("cmd.docs.long"),
}

func init() {
	docsCmd.AddCommand(docsSyncCmd)
	docsCmd.AddCommand(docsListCmd)
}
