package crypt

import (
	"context"
	"fmt"

	core "github.com/Snider/Core"
	"github.com/Snider/Core/crypt/lib/openpgp"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// createServerKeyPair is a package-level variable that can be swapped for testing.
var createServerKeyPair = openpgp.CreateServerKeyPair

// ServiceStartup Startup is called when the app starts. It handles one-time cryptographic setup.
func (s *API) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	// Define the directory for server keys based on the central config.
	//serverKeysDir := filepath.Join(s.config.DataDir, "server_keys")
	//if err := filesystem.EnsureDir(filesystem.Local, serverKeysDir); err != nil {
	//	return fmt.Errorf("failed to create server keys directory: %w", err)
	//}
	//
	//// Check for server key pair using the configured path.
	//serverKeyPath := filepath.Join(serverKeysDir, "server.lthn.pub")
	//if !filesystem.IsFile(filesystem.Local, serverKeyPath) {
	//	log.Println("Creating server key pair...")
	//	if err := createServerKeyPair(serverKeysDir); err != nil {
	//		return fmt.Errorf("failed to create server key pair: %w", err)
	//	}
	//	log.Println("Server key pair created.")
	//}
	return nil
}

// Register creates a new crypt service and registers it with the core.
var instance *API

func Register(c *core.Core) error {
	instance = &API{
		core: c,
	}
	if err := c.RegisterModule("crypt", instance); err != nil {
		return err
	}
	c.RegisterAction(handleActionCall)
	return nil
}

func handleActionCall(c *core.Core, msg core.Message) error {
	switch m := msg.(type) {

	case core.ActionServiceStartup:
		err := instance.ServiceStartup(context.Background(), application.ServiceOptions{})
		if err != nil {
			return err
		}
		c.App.Logger.Info("Crypt service started")
		return nil
	default:
		c.App.Logger.Error("Unknown message type", "type", fmt.Sprintf("%T", m))
		return nil
	}
}
