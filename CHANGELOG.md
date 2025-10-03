# Changelog

### 🚀 Features & Architectural Changes

-   **Modular Monolith Architecture**: The application has been restructured into a "modular monolith". This provides better separation of concerns and an improved development/debugging experience while maintaining a single binary for simple distribution.
    -   Introduced `procguard run api` and `procguard run daemon` commands to run the API/GUI and the background daemon as independent services.
    -   The main `procguard` command now acts as a pure CLI client that communicates with the API service.
    -   Running `procguard.exe` without arguments automatically start the API and daemon services in the background and open the web GUI.

### ♻️ Refactoring

-   **CLI API Client**:
    -   Created a new `internal/client` package to act as an API client for all CLI commands.
    -   Refactored the `procguard block` sub-commands to use the new client, centralizing API communication logic and cleaning up the command implementations.

### 🐛 Bug Fixes

-   **Complete Uninstallation**: Fixed a critical bug where the uninstaller would leave background services running. The `uninstall` command now finds and terminates all running `procguard` processes before removing files.
-   **Code Quality & Linting**:
    -   Resolved all `golangci-lint` issues, including multiple `errcheck` warnings where errors were not being handled.
    -   Fixed a platform-specific build error by correctly isolating Windows-only code.
    -   Removed dead code related to the old startup mechanism.
