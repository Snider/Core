# Configuration

The `updater` library is highly configurable via the `UpdateServiceConfig` struct.

## UpdateServiceConfig

When creating a new `UpdateService`, you pass a `UpdateServiceConfig` struct. Here are the available fields:

| Field | Type | Description |
| :--- | :--- | :--- |
| `RepoURL` | `string` | The URL to the repository for updates. Can be a GitHub repository URL (e.g., `https://github.com/owner/repo`) or a base URL for a generic HTTP update server. |
| `Channel` | `string` | Specifies the release channel to track (e.g., "stable", "prerelease"). This is **only used for GitHub-based updates**. |
| `CheckOnStartup` | `StartupCheckMode` | Determines the behavior when the service starts. See [Startup Modes](#startup-modes) below. |
| `ForceSemVerPrefix` | `bool` | Toggles whether to enforce a 'v' prefix on version tags for display and comparison. If `true`, a 'v' prefix is added if missing. |
| `ReleaseURLFormat` | `string` | A template for constructing the download URL for a release asset. The placeholder `{tag}` will be replaced with the release tag. |

### Startup Modes

The `CheckOnStartup` field can take one of the following values:

*   `updater.NoCheck`: Disables any checks on startup.
*   `updater.CheckOnStartup`: Checks for updates on startup but does not apply them.
*   `updater.CheckAndUpdateOnStartup`: Checks for and applies updates on startup.

## CLI Flags

If you are using the example CLI provided in `cmd/updater`, the following flags are available:

*   `--check-update`: Check for new updates without applying them.
*   `--do-update`: Perform an update if available.
*   `--channel`: Set the update channel (e.g., stable, beta, alpha). If not set, it's determined from the current version tag.
*   `--force-semver-prefix`: Force 'v' prefix on semver tags (default `true`).
*   `--release-url-format`: A URL format for release assets.
*   `--pull-request`: Update to a specific pull request (integer ID).
