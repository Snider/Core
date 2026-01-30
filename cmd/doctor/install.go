package doctor

import (
	"fmt"
	"runtime"
)

// printInstallInstructions prints OS-specific installation instructions
func printInstallInstructions() {
	switch runtime.GOOS {
	case "darwin":
		fmt.Println("  brew install git gh php composer node pnpm docker")
		fmt.Println("  brew install --cask claude")
	case "linux":
		fmt.Println("  # Install via your package manager or:")
		fmt.Println("  # Git: apt install git")
		fmt.Println("  # GitHub CLI: https://cli.github.com/")
		fmt.Println("  # PHP: apt install php8.3-cli")
		fmt.Println("  # Node: https://nodejs.org/")
		fmt.Println("  # pnpm: npm install -g pnpm")
	default:
		fmt.Println("  See documentation for your OS")
	}
}
