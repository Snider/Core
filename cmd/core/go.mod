module github.com/Snider/Core/cmd/core

go 1.25.5

require (
	github.com/Snider/Core/pkg/git v0.0.0
	github.com/Snider/Core/pkg/repos v0.0.0
	github.com/charmbracelet/lipgloss v1.0.0
	github.com/common-nighthawk/go-figure v0.0.0-20210622060536-734e95fb86be
	github.com/leaanthony/clir v1.7.0
	github.com/leaanthony/debme v1.2.1
	github.com/leaanthony/gosod v1.0.4
	github.com/rivo/tview v0.42.0
	golang.org/x/net v0.49.0
	golang.org/x/text v0.33.0
)

require (
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/charmbracelet/x/ansi v0.4.2 // indirect
	github.com/gdamore/encoding v1.0.1 // indirect
	github.com/gdamore/tcell/v2 v2.8.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/lucasb-eyer/go-colorful v1.2.0 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.16 // indirect
	github.com/muesli/termenv v0.15.2 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	golang.org/x/sys v0.40.0 // indirect
	golang.org/x/term v0.39.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace (
	github.com/Snider/Core => ../../
	github.com/Snider/Core/pkg/git => ../../pkg/git
	github.com/Snider/Core/pkg/repos => ../../pkg/repos
)
