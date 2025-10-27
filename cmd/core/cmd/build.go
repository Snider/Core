package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// viewState represents the current view of the TUI.
type viewState int

const (
	mainMenuState viewState = iota
	fileSelectState
	buildOutputState
)

type model struct {
	view     viewState
	choices  []string
	cursor   int
	selected map[int]struct{}

	// For file selection
	currentPath  string
	files        []fs.DirEntry
	fileCursor   int
	selectedFile string

	// For build output
	buildLog string
}

func initialModel() model {
	return model{
		view:        mainMenuState,
		choices:     []string{"Wails Build", "Exit"},
		selected:    make(map[int]struct{}),
		currentPath: ".", // Start in current directory for file selection
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

// Messages for asynchronous operations
type filesLoadedMsg []fs.DirEntry
type errorMsg error
type buildFinishedMsg string

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case filesLoadedMsg:
		m.files = msg
		m.fileCursor = 0
		return m, nil
	case errorMsg:
		m.buildLog = fmt.Sprintf("Error: %v", msg)
		m.view = buildOutputState
		return m, nil
	case buildFinishedMsg:
		m.buildLog = string(msg)
		m.view = buildOutputState
		return m, nil
	}

	switch m.view {
	case mainMenuState:
		return updateMainMenu(msg, m)
	case fileSelectState:
		return updateFileSelect(msg, m)
	case buildOutputState:
		return updateBuildOutput(msg, m)
	}
	return m, nil
}

func updateMainMenu(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter":
			switch m.choices[m.cursor] {
			case "Wails Build":
				m.view = fileSelectState
				return m, loadFilesCmd(m.currentPath)
			case "Exit":
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func updateFileSelect(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.view = mainMenuState
			return m, nil
		case "up", "k":
			if m.fileCursor > 0 {
				m.fileCursor--
			}
		case "down", "j":
			if m.fileCursor < len(m.files)-1 {
				m.fileCursor++
			}
		case "enter":
			selectedEntry := m.files[m.fileCursor]
			fullPath := filepath.Join(m.currentPath, selectedEntry.Name())
			if selectedEntry.IsDir() {
				m.currentPath = fullPath
				return m, loadFilesCmd(m.currentPath)
			} else {
				// User selected a file
				m.selectedFile = fullPath
				m.view = buildOutputState
				return m, buildWailsCmd(m.selectedFile)
			}
		case "backspace", "h":
			parentPath := filepath.Dir(m.currentPath)
			if parentPath == m.currentPath { // Already at root or current dir is "."
				return m, nil
			}
			m.currentPath = parentPath
			return m, loadFilesCmd(m.currentPath)
		}
	}
	return m, nil
}

func updateBuildOutput(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.view = mainMenuState
			m.buildLog = "" // Clear build log
			return m, nil
		}
	}
	return m, nil
}

func (m model) View() string {
	sb := strings.Builder{}
	switch m.view {
	case mainMenuState:
		sb.WriteString("Core CLI - Main Menu\n\n")
		for i, choice := range m.choices {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}
			sb.WriteString(fmt.Sprintf("%s %s\n", cursor, choice))
		}
		sb.WriteString("\nPress q to quit.\n")
	case fileSelectState:
		sb.WriteString(fmt.Sprintf("Select an HTML file for Wails build (Current: %s)\n\n", m.currentPath))
		for i, entry := range m.files {
			cursor := " "
			if m.fileCursor == i {
				cursor = ">"
			}
			name := entry.Name()
			if entry.IsDir() {
				name += "/"
			}
			sb.WriteString(fmt.Sprintf("%s %s\n", cursor, name))
		}
		sb.WriteString("\nPress Enter to select/enter, Backspace to go up, Esc to return to main menu, q to quit.\n")
	case buildOutputState:
		sb.WriteString("Wails Build Output:\n\n")
		sb.WriteString(m.buildLog)
		sb.WriteString("\n\nPress Esc to return to main menu, q to quit.\n")
	}
	return sb.String()
}

// --- Commands ---

func loadFilesCmd(path string) tea.Cmd {
	return func() tea.Msg {
		entries, err := os.ReadDir(path)
		if err != nil {
			return errorMsg(fmt.Errorf("failed to read directory %s: %w", path, err))
		}

		// Sort entries: directories first, then files, alphabetically
		sort.Slice(entries, func(i, j int) bool {
			if entries[i].IsDir() && !entries[j].IsDir() {
				return true
			}
			if !entries[i].IsDir() && entries[j].IsDir() {
				return false
			}
			return entries[i].Name() < entries[j].Name()
		})

		return filesLoadedMsg(entries)
	}
}

func buildWailsCmd(htmlPath string) tea.Cmd {
	return func() tea.Msg {
		// Find the wails3 executable
		wailsExec, err := exec.LookPath("wails3")
		if err != nil {
			return errorMsg(fmt.Errorf("wails3 executable not found in PATH: %w", err))
		}

		// Assuming the Wails project is in cmd/core-app relative to the core CLI tool
		wailsProjectDir := "../core-app"

		// Get the directory and base name of the selected HTML file
		assetDir := filepath.Dir(htmlPath)
		assetPath := filepath.Base(htmlPath)

		// Construct the wails3 build command
		// This assumes wails3 build supports overriding assetdir/assetpath via flags.
		cmdArgs := []string{
			"build",
			"-config", filepath.Join(wailsProjectDir, "build", "config.yml"),
			"--assetdir", assetDir,
			"--assetpath", assetPath,
		}

		cmd := exec.Command(wailsExec, cmdArgs...)
		cmd.Dir = wailsProjectDir // Run command from the Wails project directory

		out, err := cmd.CombinedOutput()
		if err != nil {
			return buildFinishedMsg(fmt.Sprintf("Wails build failed: %v\n%s", err, string(out)))
		}

		return buildFinishedMsg(fmt.Sprintf("Wails build successful!\n%s", string(out)))
	}
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build a Wails application",
	Long:  `This command allows you to build a Wails application, optionally selecting a custom HTML entry point.`,
	Run: func(cmd *cobra.Command, args []string) {
		p := tea.NewProgram(initialModel())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Alas, there's been an error: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}
