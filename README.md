# ProcGuard

ProcGuard is a cross-platform tool for monitoring and controlling processes on your system.

## Features

- **Process Logging:** Run a daemon to log all running processes at regular intervals.
- **Blocklist Enforcement:** The daemon can automatically kill any process that matches an entry in your blocklist.
- **Process Blocking:** Prevent executables from running by blocking them at the file system level.
- **Systemd Integration (Linux):** Easily install and manage ProcGuard as a systemd user service.
- **Configuration:** A simple JSON configuration file to manage the application's state.

## Commands

### `procguard daemon`

Run the ProcGuard daemon in the background. The daemon performs two main tasks:

1.  Logs all running processes to `~/.cache/procguard/events.log`.
2.  If a running process name matches an entry in the blocklist, the daemon will kill it.

### `procguard find <name>`

Search the process log for a specific program name. The search is case-insensitive.

### `procguard block`

Manage the process blocklist.

- **`procguard block add <name>`**: Add a program to the blocklist.
- **`procguard block rm <name>`**: Remove a program from the blocklist.
- **`procguard block list`**: Show the current blocklist.
- **`procguard block clear`**: Clear the entire blocklist.
- **`procguard block save <file>`**: Save the current blocklist to a file.
- **`procguard block load <file>`**: Load a blocklist from a file and merge it with the existing list, removing duplicates.
- **`procguard block find`**: Find all `.blocked` files.

### `procguard systemd` (Linux only)

Manage the systemd user service for the ProcGuard daemon.

- **`procguard systemd install`**: Install and enable the systemd user service. This will create a service file in `~/.config/systemd/user/` and enable it.
- **`procguard systemd remove`**: Disable and remove the systemd user service.

## Platform-Specific Behavior

ProcGuard is designed to be cross-platform, but some behaviors differ depending on your operating system.

-   **Executable Blocking:**
    -   On **Windows**, blocking an executable renames it to have a `.blocked` extension.
    -   On **Linux and macOS**, blocking an executable removes its execute permission (`chmod 0644`).

-   **Systemd Integration:**
    -   The `procguard systemd` command and its subcommands are only available on **Linux**.

## Configuration

ProcGuard uses a configuration file located at `~/.cache/procguard/spec.json` to store its state.

Currently, it stores whether the systemd service is installed:

```json
{
  "systemd_installed": false
}
```

This file is managed automatically by the `systemd` commands.
