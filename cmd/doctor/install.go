package doctor

import (
	"fmt"
	"runtime"

	"github.com/host-uk/core/pkg/i18n"
)

// printInstallInstructions prints OS-specific installation instructions
func printInstallInstructions() {
	switch runtime.GOOS {
	case "darwin":
		fmt.Printf("  %s\n", i18n.T("cmd.doctor.install_macos"))
		fmt.Printf("  %s\n", i18n.T("cmd.doctor.install_macos_cask"))
	case "linux":
		fmt.Printf("  %s\n", i18n.T("cmd.doctor.install_linux_header"))
		fmt.Printf("  %s\n", i18n.T("cmd.doctor.install_linux_git"))
		fmt.Printf("  %s\n", i18n.T("cmd.doctor.install_linux_gh"))
		fmt.Printf("  %s\n", i18n.T("cmd.doctor.install_linux_php"))
		fmt.Printf("  %s\n", i18n.T("cmd.doctor.install_linux_node"))
		fmt.Printf("  %s\n", i18n.T("cmd.doctor.install_linux_pnpm"))
	default:
		fmt.Printf("  %s\n", i18n.T("cmd.doctor.install_other"))
	}
}
