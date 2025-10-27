package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/c-bata/go-prompt"
	"github.com/charmbracelet/lipgloss"
	"mvdan.cc/sh/v3/syntax"
	// "github.com/rivo/tview" // Disabled for now due to console issues
	figure "github.com/common-nighthawk/go-figure"
	"github.com/spf13/cobra"
)

// Define some global lipgloss styles for a Tailwind dark theme
var (
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e2e8f0")). // Tailwind gray‑200
			Bold(true).
			Padding(0, 1)

	coreStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3b82f6")). // Tailwind blue-500
			Bold(true)

	subPkgStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e2e8f0")). // Tailwind gray-200
			Bold(true)

	descriptionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#e2e8f0")). // Tailwind gray-200
		//Background(lipgloss.Color("#1a202c")). // Tailwind gray-900
		Padding(1, 2)

	linkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3b82f6")). // Tailwind blue-500
			Underline(true)

	taglineStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#e2e8f0")).
			PaddingTop(2).PaddingLeft(8).PaddingBottom(1). // vertical spacing around the tagline
			Align(lipgloss.Center)                         // centre it under the big words

)

var rootCmd = &cobra.Command{
	Use:   "core",
	Short: coreStyle.Render("Core") + subPkgStyle.Render(".Framework") + " is a CLI tool for managing applications",
	Long: descriptionStyle.Render(
		"\n" +
			"                      " + coreStyle.Render("Core") + subPkgStyle.Render(".Framework") + "    " + coreStyle.Render("Version V0.0.1") + "\n\n\n" +
			"For more information: " + linkStyle.Render("https://core.help") + " and " + linkStyle.Render("https://lt.hn") + "\n\n" +
			"managing various aspects of Core.Framework applications."),
	Run: func(cmd *cobra.Command, args []string) {
		// Default behaviour if no subcommand is given
		err := cmd.Help()
		if err != nil {
			return
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Here you can define global flags and configuration that apply to all commands.
	if os.Getenv("CORE_DEV_TOOLS") == "true" {
		initDevTools()
	}

	coreFig := figure.NewFigure("Core", "big", true)
	frameworkFig := figure.NewFigure("Framework", "big", true)

	// Apply the colour styles to the Figlet output.
	coreBlock := coreStyle.Render(coreFig.String())
	frameworkBlock := subPkgStyle.Render(frameworkFig.String())

	gap := lipgloss.NewStyle().Width(4).Render("")
	bigWords := lipgloss.JoinHorizontal(lipgloss.Top, gap, coreBlock, gap, frameworkBlock)
	tagline := taglineStyle.Render(
		"the birthplace of what Web3 will become",
	)

	output := lipgloss.JoinVertical(lipgloss.Left,
		bigWords, // big “Core   Framework”
		tagline,  // bottom tagline, centred
	)
	// Print the whole thing.
	fmt.Print(output)
}

func initDevTools() {
	// Example usage for survey
	name := ""
	surveyPrompt := &survey.Input{
		Message: "What is your name?",
	}
	_ = survey.AskOne(surveyPrompt, &name)
	fmt.Printf("Hello, %s (from survey)!\n", name)

	// Example usage for go-prompt
	fmt.Println("Starting go-prompt (type 'exit' to quit)...")
	executor := func(in string) {
		fmt.Printf("You typed: %s\n", in)
		if in == "exit" {
			os.Exit(0)
		}
	}
	completer := func(d prompt.Document) []prompt.Suggest {
		s := []prompt.Suggest{
			{Text: "hello", Description: "Say hello"},
			{Text: "world", Description: "Say world"},
			{Text: "exit", Description: "Exit the prompt"},
		}
		return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
	}
	// go-prompt will block, so we run it in a goroutine for this example
	// In a real app, you'd likely manage its lifecycle more carefully.
	// The prompt.New function returns a Prompt object, which then has a Run method.
	// For a simple blocking run, you can call .Run() directly on the New object.
	go prompt.New(executor, completer, prompt.OptionPrefix(">>> ")).Run()

	// Example usage for lipgloss
	// Using the global descriptionStyle for consistency
	fmt.Println(descriptionStyle.Render("Hello from Lipgloss!"))

	// Example usage for mvdan.cc/sh/v3/syntax
	parser := syntax.NewParser()
	word, err := parser.Parse(strings.NewReader("echo hello world"), "<stdin>")
	if err != nil {
		fmt.Printf("Error parsing shell command: %v\n", err)
	} else {
		// For demonstration, just print the parsed statement.
		fmt.Printf("Parsed shell command: %v\n", word.Stmts[0])
	}

	// The tview example is disabled for now due to console issues.
	// Uncomment the following block to re-enable it.
	/*
		app := tview.NewApplication()
		box := tview.NewBox().
			SetBorder(true).
			SetTitle("Hello from tview!")

		go func() {
			if err := app.SetRoot(box, true).Run(); err != nil {
				panic(err)
			}
		}()

		time.Sleep(2 * time.Second)
	*/
}
