# Agent Guide

This file provides instructions for AI agents working on this project.

## Core Responsibilities

- **Keep README.md Updated:** Whenever you add new features, change flags, or modify the application's architecture, you MUST update the `README.md` file to reflect these changes.
- **Maintain Versioning:** Increment the version number in `README.md` (e.g., from v2 to v3) when significant new functionality is added.
- **Document New Flags:** Ensure every new CLI flag is added to the "Flags" table in `README.md`.
- **Explain New Modes:** If a new mode (like Web UI or a new migration type) is added, provide a dedicated section in `README.md` with usage examples.
- **Update Tasks:** For every new feature, create a corresponding task in `docs/` with clear implementation steps.

## Code Standards

- Use structured logging (`log/slog`) for all significant events.
- Ensure all new features are covered by the dry-run mode if applicable.
- Follow the existing internal package structure (`internal/config`, `internal/db`, `internal/migration`, `internal/web`).
