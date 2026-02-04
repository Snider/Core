package help

import (
	"fmt"
	"strings"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/help"
)

func init() {
	cli.RegisterCommands(AddHelpCommands)
}

func AddHelpCommands(root *cli.Command) {
	var searchFlag string

	helpCmd := &cli.Command{
		Use:   "help [topic]",
		Short: "Display help documentation",
		Run: func(cmd *cli.Command, args []string) {
			catalog := help.DefaultCatalog()

			if searchFlag != "" {
				results := catalog.Search(searchFlag)
				if len(results) == 0 {
					fmt.Println("No topics found.")
					return
				}
				fmt.Println("Search Results:")
				for _, res := range results {
					title := res.Topic.Title
					if res.Section != nil {
						title = fmt.Sprintf("%s > %s", res.Topic.Title, res.Section.Title)
					}
					// Use bold for title
					fmt.Printf("  \033[1m%s\033[0m (%s)\n", title, res.Topic.ID)
					if res.Snippet != "" {
						// Highlight markdown bold as ANSI bold for CLI output
						fmt.Printf("    %s\n", replaceMarkdownBold(res.Snippet))
					}
					fmt.Println()
				}
				return
			}

			if len(args) == 0 {
				topics := catalog.List()
				fmt.Println("Available Help Topics:")
				for _, t := range topics {
					fmt.Printf("  %s - %s\n", t.ID, t.Title)
				}
				return
			}

			topic, err := catalog.Get(args[0])
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}

			renderTopic(topic)
		},
	}

	helpCmd.Flags().StringVarP(&searchFlag, "search", "s", "", "Search help topics")
	root.AddCommand(helpCmd)
}

func replaceMarkdownBold(s string) string {
	parts := strings.Split(s, "**")
	var result strings.Builder
	for i, part := range parts {
		result.WriteString(part)
		if i < len(parts)-1 {
			if i%2 == 0 {
				result.WriteString("\033[1m")
			} else {
				result.WriteString("\033[0m")
			}
		}
	}
	return result.String()
}

func renderTopic(t *help.Topic) {
	// Simple ANSI rendering for now
	// Use explicit ANSI codes or just print
	fmt.Printf("\n\033[1;34m%s\033[0m\n", t.Title) // Blue bold title
	fmt.Println("----------------------------------------")
	fmt.Println(t.Content)
	fmt.Println()
}
