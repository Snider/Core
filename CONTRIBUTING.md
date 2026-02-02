# Contributing to Core

First off, thank you for considering contributing to Core. It's people like you that make Core such a great tool.

Following these guidelines helps to communicate that you respect the time of the developers managing and developing this open source project. In return, they should reciprocate that respect in addressing your issue, assessing changes, and helping you finalize your pull requests.

## Code of Conduct

This project and everyone participating in it is governed by the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/version/2/1/0/code_of_conduct.html). By participating, you are expected to uphold this code. Please report unacceptable behavior to the project maintainers.

## How Can I Contribute?

### Reporting Bugs

This section guides you through submitting a bug report for Core. Following these guidelines helps maintainers and the community understand your report, reproduce the behavior, and find related reports.

*   **Use a clear and descriptive title** for the issue to identify the problem.
*   **Describe the exact steps which reproduce the problem** in as many details as possible.
*   **Provide specific examples to demonstrate the steps.** Include links to files or GitHub projects, or copy/pasteable snippets, which you use in those examples. If you're providing snippets in the issue, use Markdown code blocks.
*   **Describe the behavior you observed after following the steps** and point out what exactly is the problem with that behavior.
*   **Explain which behavior you expected to see instead and why.**
*   **Include screenshots and animated GIFs** which show you following the described steps and clearly demonstrate the problem.

### Suggesting Enhancements

This section guides you through submitting an enhancement suggestion for Core, including completely new features and minor improvements to existing functionality.

*   **Use a clear and descriptive title** for the issue to identify the suggestion.
*   **Provide a step-by-step description of the suggested enhancement** in as many details as possible.
*   **Provide specific examples to demonstrate the steps.** Include copy/pasteable snippets which you use in those examples, and Markdown code blocks.
*   **Describe the current behavior and explain which behavior you expected to see instead and why.**
*   **Explain why this enhancement would be useful** to most Core users.

### Pull Requests

The process described here has several goals:

- Maintain Core's quality
- Fix problems that are important to users
- Engage the community in working toward the best possible Core
- A speedy response to your pull request

Please follow these steps to have your contribution considered by the maintainers:

1.  Follow all instructions in the template
2.  Follow the [styleguides](#styleguides)
3.  After you submit your pull request, verify that all status checks are passing <details><summary>What if the status checks are failing?</summary>If a status check is failing, and you believe that the failure is unrelated to your change, please leave a comment on the pull request explaining why you believe the failure is unrelated. A maintainer will re-run the status check for you. If we conclude that the failure was a false positive, then we will open an issue to track that problem with our CI infrastructure.</details>

While the prerequisites for contributing to Core are spelled out here, we suggest that you familiarize yourself with the project's [README.md](README.md) before contributing.

## Styleguides

### Git Commit Messages

*   Use the present tense ("Add feature" not "Added feature")
*   Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
*   Limit the first line to 72 characters or less
*   Reference issues and pull requests liberally after the first line
*   When only changing documentation, include `[ci skip]` in the commit title
*   Consider starting the commit message with an applicable emoji:
    *   🎨 `:art:` when improving the format/structure of the code
    *   🐎 `:racehorse:` when improving performance
    *   🚱 `:non-potable_water:` when plugging memory leaks
    *   📝 `:memo:` when writing docs
    *   🐧 `:penguin:` when fixing something on Linux
    *   🍎 `:apple:` when fixing something on macOS
    *   🏁 `:checkered_flag:` when fixing something on Windows
    *   🐛 `:bug:` when fixing a bug
    *   🔥 `:fire:` when removing code or files
    *   💚 `:green_heart:` when fixing the CI build
    *   ✅ `:white_check_mark:` when adding tests
    *   🔒 `:lock:` when dealing with security
    *   ⬆️ `:arrow_up:` when upgrading dependencies
    *   ⬇️ `:arrow_down:` when downgrading dependencies
    *   👕 `:shirt:` when removing linter warnings

### Go Styleguide

All Go code must adhere to the style guides outlined in [Effective Go](https://go.dev/doc/effective_go.html) and the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).

- Use `gofmt` to format your code.
- Use `golint` to lint your code.

## Releasing

The release process is automated. When a pull request is merged, a new version is automatically released to the appropriate channel.
