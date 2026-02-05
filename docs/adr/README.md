# Architecture Decision Records (ADR)

This directory contains the Architecture Decision Records for the Core project.

## What is an ADR?

An Architecture Decision Record (ADR) is a document that captures an important architectural decision made along with its context and consequences.

## Why use ADRs?

- **Context:** Helps new contributors understand *why* certain decisions were made.
- **History:** Provides a historical record of the evolution of the project's architecture.
- **Transparency:** Makes the decision-making process transparent and open for discussion.

## ADR Process

1. **Identify a Decision:** When an architectural decision needs to be made, start a new ADR.
2. **Use the Template:** Copy `0000-template.md` to a new file named `NNNN-short-title.md` (e.g., `0001-use-wails-v3.md`).
3. **Draft the ADR:** Fill in the context, drivers, and considered options.
4. **Propose:** Set the status to `proposed` and open a Pull Request for discussion.
5. **Accept/Reject:** Once a consensus is reached, update the status to `accepted` or `rejected` and merge.
6. **Supersede:** If a later decision changes an existing one, update the status of the old ADR to `superseded` and point to the new one.

## ADR Index

| ID | Title | Status | Date |
|---|---|---|---|
| 0000 | [ADR Template](0000-template.md) | N/A | 2025-05-15 |
| 0001 | [Use Wails v3 for GUI](0001-use-wails-v3.md) | accepted | 2025-05-15 |
| 0002 | [IPC Bridge Pattern](0002-ipc-bridge-pattern.md) | accepted | 2025-05-15 |
| 0003 | [Service-Oriented Architecture with Dual-Constructor DI](0003-soa-dual-constructor-di.md) | accepted | 2025-05-15 |
| 0004 | [Storage Abstraction via Medium Interface](0004-storage-abstraction-medium.md) | accepted | 2025-05-15 |
