package unifi

import (
	"fmt"

	"github.com/host-uk/core/pkg/cli"
	uf "github.com/host-uk/core/pkg/unifi"
)

// Config command flags.
var (
	configURL      string
	configUser     string
	configPass     string
	configAPIKey   string
	configInsecure bool
	configTest     bool
)

// addConfigCommand adds the 'config' subcommand for UniFi connection setup.
func addConfigCommand(parent *cli.Command) {
	cmd := &cli.Command{
		Use:   "config",
		Short: "Configure UniFi connection",
		Long:  "Set the UniFi controller URL and credentials, or test the current connection.",
		RunE: func(cmd *cli.Command, args []string) error {
			return runConfig(cmd)
		},
	}

	cmd.Flags().StringVar(&configURL, "url", "", "UniFi controller URL")
	cmd.Flags().StringVar(&configUser, "user", "", "UniFi username")
	cmd.Flags().StringVar(&configPass, "pass", "", "UniFi password")
	cmd.Flags().StringVar(&configAPIKey, "apikey", "", "UniFi API key")
	cmd.Flags().BoolVar(&configInsecure, "insecure", false, "Allow insecure TLS connections (e.g. self-signed certs)")
	cmd.Flags().BoolVar(&configTest, "test", false, "Test the current connection")

	parent.AddCommand(cmd)
}

func runConfig(cmd *cli.Command) error {
	var insecure *bool
	if cmd.Flags().Changed("insecure") {
		insecure = &configInsecure
	}

	// If setting values, save them first
	if configURL != "" || configUser != "" || configPass != "" || configAPIKey != "" || insecure != nil {
		if err := uf.SaveConfig(configURL, configUser, configPass, configAPIKey, insecure); err != nil {
			return err
		}

		if configURL != "" {
			cli.Success(fmt.Sprintf("UniFi URL set to %s", configURL))
		}
		if configUser != "" {
			cli.Success("UniFi username saved")
		}
		if configPass != "" {
			cli.Success("UniFi password saved")
		}
		if configAPIKey != "" {
			cli.Success("UniFi API key saved")
		}
		if insecure != nil {
			cli.Success(fmt.Sprintf("UniFi insecure mode set to %v", *insecure))
		}
	}

	// If testing, verify the connection
	if configTest {
		return runConfigTest(cmd)
	}

	// If no flags, show current config
	if configURL == "" && configUser == "" && configPass == "" && configAPIKey == "" && !configInsecure && !configTest {
		return showConfig()
	}

	return nil
}

func showConfig() error {
	url, user, pass, apikey, insecure, err := uf.ResolveConfig("", "", "", "", nil)
	if err != nil {
		return err
	}

	cli.Blank()
	cli.Print("  %s %s\n", dimStyle.Render("URL:"), valueStyle.Render(url))

	if user != "" {
		cli.Print("  %s %s\n", dimStyle.Render("User:"), valueStyle.Render(user))
	} else {
		cli.Print("  %s %s\n", dimStyle.Render("User:"), warningStyle.Render("not set"))
	}

	if pass != "" {
		cli.Print("  %s %s\n", dimStyle.Render("Pass:"), valueStyle.Render("****"))
	} else {
		cli.Print("  %s %s\n", dimStyle.Render("Pass:"), warningStyle.Render("not set"))
	}

	if apikey != "" {
		masked := apikey
		if len(apikey) >= 8 {
			masked = apikey[:4] + "..." + apikey[len(apikey)-4:]
		}
		cli.Print("  %s %s\n", dimStyle.Render("API Key:"), valueStyle.Render(masked))
	} else {
		cli.Print("  %s %s\n", dimStyle.Render("API Key:"), warningStyle.Render("not set"))
	}

	if insecure {
		cli.Print("  %s %s\n", dimStyle.Render("Insecure:"), warningStyle.Render("enabled"))
	} else {
		cli.Print("  %s %s\n", dimStyle.Render("Insecure:"), successStyle.Render("disabled"))
	}

	cli.Blank()

	return nil
}

func runConfigTest(cmd *cli.Command) error {
	var insecure *bool
	if cmd.Flags().Changed("insecure") {
		insecure = &configInsecure
	}

	client, err := uf.NewFromConfig(configURL, configUser, configPass, configAPIKey, insecure)
	if err != nil {
		return err
	}

	sites, err := client.GetSites()
	if err != nil {
		cli.Error("Connection failed")
		return cli.WrapVerb(err, "connect to", "UniFi controller")
	}

	cli.Blank()
	cli.Success(fmt.Sprintf("Connected to %s", client.URL()))
	cli.Print("  %s %s\n", dimStyle.Render("Sites:"), numberStyle.Render(fmt.Sprintf("%d", len(sites))))
	for _, s := range sites {
		cli.Print("    %s %s\n", valueStyle.Render(s.Name), dimStyle.Render(s.Desc))
	}
	cli.Blank()

	return nil
}
