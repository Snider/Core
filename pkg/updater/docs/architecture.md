# Architecture

The `updater` library is designed to facilitate self-updates for Go applications by replacing the running binary with a newer version downloaded from a remote source.

## Update Mechanisms

The library supports two primary update sources:

1.  **GitHub Releases:** Fetches releases directly from a GitHub repository.
2.  **Generic HTTP:** Fetches update information from a generic HTTP endpoint.

### GitHub Releases

When configured with a GitHub repository URL (e.g., `https://github.com/owner/repo`), the updater uses the GitHub API to find releases.

*   **Channel Support:** You can specify a "channel" (e.g., "stable", "beta"). The updater will filter releases based on this channel.
    *   Ideally, this maps to release tags or pre-release status (though the specific implementation details of how "channel" maps to GitHub release types should be verified in the code).
*   **Pull Request Updates:** The library supports updating to a specific pull request artifact, useful for testing pre-release builds.

### Generic HTTP

When configured with a generic HTTP URL, the updater expects the endpoint to return a JSON object describing the latest version.

**Expected JSON Format:**

```json
{
  "version": "1.2.3",
  "url": "https://your-server.com/path/to/release-asset"
}
```

The updater compares the `version` from the JSON with the current application version. If the remote version is newer, it downloads the binary from the `url`.

## Version Comparison

The library uses Semantic Versioning (SemVer) to compare versions.

*   **Prefix Handling:** The `ForceSemVerPrefix` configuration option allows you to standardize version tags by enforcing a `v` prefix (e.g., `v1.0.0` vs `1.0.0`) for consistent comparison.
*   **Logic:**
    *   If `Remote Version` > `Current Version`: Update available.
    *   If `Remote Version` <= `Current Version`: Up to date.

## Self-Update Process

The actual update process is handled by the `minio/selfupdate` library.

1.  **Download:** The new binary is downloaded from the source.
2.  **Verification:** (Depending on configuration/implementation) Checksums may be verified.
3.  **Apply:** The current executable file is replaced with the new binary.
    *   **Windows:** The old binary is renamed (often to `.old`) before replacement to allow the write operation.
    *   **Linux/macOS:** The file is unlinked and replaced.
4.  **Restart:** The application usually needs to be restarted for the changes to take effect. The `updater` library currently handles the *replacement*, but the *restart* logic is typically left to the application.
