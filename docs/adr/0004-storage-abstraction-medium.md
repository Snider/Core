# ADR 0004: Storage Abstraction via Medium Interface

* Status: accepted
* Deciders: Project Maintainers
* Date: 2025-05-15

## Context and Problem Statement

The application needs to support different storage backends (local file system, SFTP, WebDAV, etc.) for its workspace data. Hardcoding file system operations would make it difficult to support remote storage.

## Decision Drivers

* Flexibility in storage backends.
* Ease of testing (mocking storage).
* Uniform API for file operations.

## Considered Options

* Standard `os` package.
* Interface abstraction (`Medium`).
* `spf13/afero` library.

## Decision Outcome

Chosen option: "Interface abstraction (`Medium`)". We defined a custom `Medium` interface in `pkg/io` that abstracts common file operations.

### Positive Consequences

* Application logic is agnostic of where files are actually stored.
* Easy to implement new backends (SFTP, WebDAV).
* Simplified testing using `MockMedium`.

### Negative Consequences

* Small overhead of interface calls.
* Need to ensure all file operations go through the `Medium` interface.
