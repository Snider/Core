package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/common-nighthawk/go-figure"
	"github.com/leaanthony/clir"

	"github.com/Snider/Core/cmd/core/cmd/dev"
)

// IsDevMode indicates if the application is running in development or CI mode.
var IsDevMode bool

// Define some global lipgloss styles for a Tailwind dark theme
var (
	coreStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3b82f6")). // Tailwind blue-500
			Bold(true)

	subPkgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e2e8f0")). // Tailwind gray-200
			Bold(true)

	descriptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#e2e8f0")). // Tailwind gray-200
				Padding(1, 2)

	linkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3b82f6")). // Tailwind blue-500
			Underline(true)

	taglineStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e2e8f0")).
			PaddingTop(2).PaddingLeft(8).PaddingBottom(1).
			Align(lipgloss.Center)
)

// Execute creates the root CLI application and runs it.
func Execute() error {
	// Determine if in dev mode
	executablePath, err := os.Executable()
	if err == nil {
		// Check if the executable path contains the build path for dev mode
		if strings.Contains(executablePath, "Core/cmd/core/bin/") {
			IsDevMode = true
		}
	}

	// Check for DEV or CI environment variables
	if os.Getenv("DEV") != "" || os.Getenv("CI") != "" {
		IsDevMode = true
	}

	// Create a new clir instance, removing the description and version to avoid the default header.
	app := clir.NewCli("core", "", "")

	// Recreate the header with better alignment.
	title := coreStyle.Render("Core") + subPkgStyle.Render(".Framework")
	version := coreStyle.Render("Version V0.0.1")
	titleLine := lipgloss.JoinHorizontal(lipgloss.Top, title, lipgloss.NewStyle().Width(4).Render(""), version)

	linksLine := "For more information: " + linkStyle.Render("https://core.help") + " and " + linkStyle.Render("https://lt.hn")
	descLine := "managing various aspects of Core.Framework applications."

	headerBlock := lipgloss.JoinVertical(lipgloss.Center,
		titleLine,
		"", // blank line
		"", // blank line
		linksLine,
		"", // blank line
		descLine,
	)

	// Set the long description using a centered container.
	app.LongDescription(lipgloss.NewStyle().Padding(1, 2).Align(lipgloss.Center).Render(headerBlock))

	// Default action when no command is given is to show the banner and then the help.
	app.Action(func() error {
		showBanner()
		app.PrintHelp()
		return nil
	})

	// Add the top-level commands
	devCmd := app.NewSubCommand("dev", "Development tools for Core Framework")
	dev.AddAPICommands(devCmd)
	dev.AddTestGenCommand(devCmd)
	dev.AddSyncCommand(devCmd)
	dev.AddDocCommand(devCmd)
	dev.AddBuildCommand(devCmd)
	// Run the application
	return app.Run()
}

// showBanner generates and prints the ASCII art banner.
func showBanner() {
	if !IsDevMode {
		return
	}
	coreFig := figure.NewFigure("Core", "big", true)
	frameworkFig := figure.NewFigure("Framework", "big", true)

	coreBlock := coreStyle.Render(coreFig.String())
	frameworkBlock := subPkgStyle.Render(frameworkFig.String())

	gap := lipgloss.NewStyle().Width(4).Render("")
	bigWords := lipgloss.JoinHorizontal(lipgloss.Top, gap, coreBlock, gap, frameworkBlock)
	tagline := taglineStyle.Render(
		"the birthplace of what Web3 will become",
	)

	output := lipgloss.JoinVertical(lipgloss.Left,
		bigWords,
		tagline,
	)

	fmt.Print(output)
}
