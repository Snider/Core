// Package docs provides documentation management commands.
package docs

import (
	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/spf13/cobra"
)

// Style and utility aliases from shared
var (
	repoNameStyle    = cli.RepoNameStyle
	successStyle     = cli.SuccessStyle
	errorStyle       = cli.ErrorStyle
	dimStyle         = cli.DimStyle
	headerStyle      = cli.HeaderStyle
	confirm          = cli.Confirm
	docsFoundStyle   = cli.SuccessStyle
	docsMissingStyle = cli.DimStyle
	docsFileStyle    = cli.InfoStyle
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
