package deploy

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/host-uk/core/pkg/cli"
	"github.com/host-uk/core/pkg/deploy/coolify"
	"github.com/host-uk/core/pkg/i18n"
	"github.com/spf13/cobra"
)

var (
	coolifyURL   string
	coolifyToken string
	outputJSON   bool
)

// Cmd is the root deploy command.
var Cmd = &cobra.Command{
	Use:   "deploy",
	Short: i18n.T("cmd.deploy.short"),
	Long:  i18n.T("cmd.deploy.long"),
}

var serversCmd = &cobra.Command{
	Use:   "servers",
	Short: "List Coolify servers",
	RunE:  runListServers,
}

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "List Coolify projects",
	RunE:  runListProjects,
}

var appsCmd = &cobra.Command{
	Use:   "apps",
	Short: "List Coolify applications",
	RunE:  runListApps,
}

var dbsCmd = &cobra.Command{
	Use:     "databases",
	Short:   "List Coolify databases",
	Aliases: []string{"dbs", "db"},
	RunE:    runListDatabases,
}

var servicesCmd = &cobra.Command{
	Use:   "services",
	Short: "List Coolify services",
	RunE:  runListServices,
}

var teamCmd = &cobra.Command{
	Use:   "team",
	Short: "Show current team info",
	RunE:  runTeam,
}

var callCmd = &cobra.Command{
	Use:   "call <operation> [params-json]",
	Short: "Call any Coolify API operation",
	Args:  cobra.RangeArgs(1, 2),
	RunE:  runCall,
}

func init() {
	// Global flags
	Cmd.PersistentFlags().StringVar(&coolifyURL, "url", os.Getenv("COOLIFY_URL"), "Coolify API URL")
	Cmd.PersistentFlags().StringVar(&coolifyToken, "token", os.Getenv("COOLIFY_TOKEN"), "Coolify API token")
	Cmd.PersistentFlags().BoolVar(&outputJSON, "json", false, "Output as JSON")

	// Add subcommands
	Cmd.AddCommand(serversCmd)
	Cmd.AddCommand(projectsCmd)
	Cmd.AddCommand(appsCmd)
	Cmd.AddCommand(dbsCmd)
	Cmd.AddCommand(servicesCmd)
	Cmd.AddCommand(teamCmd)
	Cmd.AddCommand(callCmd)
}

func getClient() (*coolify.Client, error) {
	cfg := coolify.Config{
		BaseURL:   coolifyURL,
		APIToken:  coolifyToken,
		Timeout:   30,
		VerifySSL: true,
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = os.Getenv("COOLIFY_URL")
	}
	if cfg.APIToken == "" {
		cfg.APIToken = os.Getenv("COOLIFY_TOKEN")
	}

	return coolify.NewClient(cfg)
}

func outputResult(data any) error {
	if outputJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(data)
	}

	// Pretty print based on type
	switch v := data.(type) {
	case []map[string]any:
		for _, item := range v {
			printItem(item)
		}
	case map[string]any:
		printItem(v)
	default:
		fmt.Printf("%v\n", data)
	}
	return nil
}

func printItem(item map[string]any) {
	// Common fields to display
	if uuid, ok := item["uuid"].(string); ok {
		fmt.Printf("%s  ", cli.DimStyle.Render(uuid[:8]))
	}
	if name, ok := item["name"].(string); ok {
		fmt.Printf("%s", cli.TitleStyle.Render(name))
	}
	if desc, ok := item["description"].(string); ok && desc != "" {
		fmt.Printf("  %s", cli.DimStyle.Render(desc))
	}
	if status, ok := item["status"].(string); ok {
		switch status {
		case "running":
			fmt.Printf("  %s", cli.SuccessStyle.Render("●"))
		case "stopped":
			fmt.Printf("  %s", cli.ErrorStyle.Render("○"))
		default:
			fmt.Printf("  %s", cli.DimStyle.Render(status))
		}
	}
	fmt.Println()
}

func runListServers(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	servers, err := client.ListServers(context.Background())
	if err != nil {
		return err
	}

	if len(servers) == 0 {
		fmt.Println("No servers found")
		return nil
	}

	return outputResult(servers)
}

func runListProjects(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	projects, err := client.ListProjects(context.Background())
	if err != nil {
		return err
	}

	if len(projects) == 0 {
		fmt.Println("No projects found")
		return nil
	}

	return outputResult(projects)
}

func runListApps(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	apps, err := client.ListApplications(context.Background())
	if err != nil {
		return err
	}

	if len(apps) == 0 {
		fmt.Println("No applications found")
		return nil
	}

	return outputResult(apps)
}

func runListDatabases(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	dbs, err := client.ListDatabases(context.Background())
	if err != nil {
		return err
	}

	if len(dbs) == 0 {
		fmt.Println("No databases found")
		return nil
	}

	return outputResult(dbs)
}

func runListServices(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	services, err := client.ListServices(context.Background())
	if err != nil {
		return err
	}

	if len(services) == 0 {
		fmt.Println("No services found")
		return nil
	}

	return outputResult(services)
}

func runTeam(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return err
	}

	team, err := client.GetTeam(context.Background())
	if err != nil {
		return err
	}

	return outputResult(team)
}

func runCall(cmd *cobra.Command, args []string) error {
	client, err := getClient()
	if err != nil {
		return cli.WrapVerb(err, "initialize", "client")
	}

	operation := args[0]
	var params map[string]any
	if len(args) > 1 {
		if err := json.Unmarshal([]byte(args[1]), &params); err != nil {
			return fmt.Errorf("invalid JSON params: %w", err)
		}
	}

	result, err := client.Call(context.Background(), operation, params)
	if err != nil {
		return err
	}

	return outputResult(result)
}
