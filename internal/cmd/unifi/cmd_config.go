package unifi

import (
	"fmt"

	"github.com/host-uk/core/pkg/cli"
	uf "github.com/host-uk/core/pkg/unifi"
)

// Config command flags.
var (
	configURL    string
	configUser   string
	configPass   string
	configAPIKey string
	configTest   bool
)

// addConfigCommand adds the 'config' subcommand for UniFi connection setup.
func addConfigCommand(parent *cli.Command) {
	cmd := &cli.Command{
		Use:   "config",
		Short: "Configure UniFi connection",
		Long:  "Set the UniFi controller URL and credentials, or test the current connection.",
		RunE: func(cmd *cli.Command, args []string) error {
			return runConfig()
		},
	}

	cmd.Flags().StringVar(&configURL, "url", "", "UniFi controller URL")
	cmd.Flags().StringVar(&configUser, "user", "", "UniFi username")
	cmd.Flags().StringVar(&configPass, "pass", "", "UniFi password")
	cmd.Flags().StringVar(&configAPIKey, "apikey", "", "UniFi API key")
	cmd.Flags().BoolVar(&configTest, "test", false, "Test the current connection")

	parent.AddCommand(cmd)
}

func runConfig() error {
	// If setting values, save them first
	if configURL != "" || configUser != "" || configPass != "" || configAPIKey != "" {
		if err := uf.SaveConfig(configURL, configUser, configPass, configAPIKey); err != nil {
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
	}

	// If testing, verify the connection
	if configTest {
		return runConfigTest()
	}

	// If no flags, show current config
	if configURL == "" && configUser == "" && configPass == "" && configAPIKey == "" && !configTest {
		return showConfig()
	}

	return nil
}

func showConfig() error {
	url, user, pass, apikey, err := uf.ResolveConfig("", "", "", "")
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

	cli.Blank()

	return nil
}

func runConfigTest() error {
	client, err := uf.NewFromConfig(configURL, configUser, configPass, configAPIKey)
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
