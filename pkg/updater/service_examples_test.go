package updater_test

import (
	"fmt"
	"log"

	"github.com/host-uk/core/pkg/updater"
)

func ExampleNewUpdateService() {
	// Mock the update check functions to prevent actual updates during tests
	updater.CheckForUpdates = func(owner, repo, channel string, forceSemVerPrefix bool, releaseURLFormat string) error {
		fmt.Println("CheckForUpdates called")
		return nil
	}
	defer func() {
		updater.CheckForUpdates = nil // Restore original function
	}()

	config := updater.UpdateServiceConfig{
		RepoURL:        "https://github.com/owner/repo",
		Channel:        "stable",
		CheckOnStartup: updater.CheckAndUpdateOnStartup,
	}
	updateService, err := updater.NewUpdateService(config)
	if err != nil {
		log.Fatalf("Failed to create update service: %v", err)
	}
	if err := updateService.Start(); err != nil {
		log.Printf("Update check failed: %v", err)
	}
	// Output: CheckForUpdates called
}

func ExampleParseRepoURL() {
	owner, repo, err := updater.ParseRepoURL("https://github.com/owner/repo")
	if err != nil {
		log.Fatalf("Failed to parse repo URL: %v", err)
	}
	fmt.Printf("Owner: %s, Repo: %s", owner, repo)
	// Output: Owner: owner, Repo: repo
}
