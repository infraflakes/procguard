# Changelog

## ♻️ Refactoring

-   **Platform-Specific Logic**:
    -   Introduced a new `internal/platform` package to abstract all OS-specific operations.
    -   Consolidated duplicated file-blocking logic into the new package, using build tags to provide the correct implementation for each OS.
    -   Removed the redundant `cmd/unix/block_unix.go` and `cmd/windows/block_win.go` files.

-   **Configuration Management**:
    -   Simplified configuration by unifying the separate Windows and Linux `Config` structs into a single, cross-platform struct in `internal/config/config.go`.
    -   Removed the now-unnecessary `internal/config/config_others.go` and `internal/config/config_windows.go` files.

-   **GUI Handler Organization**:
    -   Decomposed the large `cmd/gui/handlers.go` file into smaller, more focused files based on their API functionality (`auth_handlers.go`, `blocklist_handlers.go`, etc.). This makes the GUI's backend code easier to navigate and maintain.

-   **Systemd Code Structure**:
    -   Broke down the large `installSystemdServiceE` and `removeSystemdServiceE` functions in `cmd/unix/systemd_linux.go` into smaller, single-purpose functions, improving readability.

-   **Separation of Concerns**:
    -   Moved the core business logic for managing the blocklist from the `cmd/block` package to the `internal/blocklist` package.
    -   The `cmd/block` commands now act as a thin layer, calling the centralized logic in the `internal` package.

## 🐛 Bug Fixes

-   **Error Handling**:
    -   Corrected several places in the `internal/blocklist` package where errors from JSON marshaling and unmarshaling were being ignored. This prevents the application from failing silently or corrupting the blocklist data.
