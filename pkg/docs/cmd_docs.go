// Package docs provides documentation management commands.
package docs

import (
	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/i18n"
)

// Style and utility aliases from shared
var (
	repoNameStyle  = cli.RepoStyle
	successStyle   = cli.SuccessStyle
	errorStyle     = cli.ErrorStyle
	dimStyle       = cli.DimStyle
	headerStyle    = cli.HeaderStyle
	confirm        = cli.Confirm
	docsFoundStyle = cli.SuccessStyle
	docsFileStyle  = cli.InfoStyle
)

var docsCmd = &cli.Command{
	Use:   "docs",
	Short: i18n.T("cmd.docs.short"),
	Long:  i18n.T("cmd.docs.long"),
}

func init() {
	docsCmd.AddCommand(docsSyncCmd)
	docsCmd.AddCommand(docsListCmd)
}
