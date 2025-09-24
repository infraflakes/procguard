# ProcGuard

ProcGuard is a cross-platform tool for monitoring and controlling processes on your system, featuring a simple web-based GUI.

## Main Usage (GUI Mode)

The easiest way to use ProcGuard is to run the executable directly with no commands:

```bash
./procguard
# On Windows, double-click procguard.exe
```

This will:
1.  **Check for an existing instance.** If ProcGuard is already running, it will simply open a new browser tab to the dashboard and exit.
2.  **Start the backend daemon.** If this is the first instance, it will start the background monitoring process.
3.  **Launch the GUI.** It will start a local web server and automatically open your default browser to the dashboard.

### GUI Features
- **Search:** Search the process logs for specific application names.
- **Blocklist Management:** View all currently blocked applications, select multiple apps, and unblock them in a single action.

## CLI Commands

While GUI mode is the primary way to use the application, the following CLI commands are still available for scripting and advanced usage.

### `procguard` (no command)
Launches the smart-starting GUI and daemon. If an instance is already running, it just opens the browser.

### `procguard gui`
Explicitly starts the GUI and web server. This is now the same as the default command.

### `procguard daemon`
Runs the ProcGuard daemon in the background. The daemon performs two main tasks:
1.  Logs running processes to `~/.cache/procguard/events.log`.
2.  Kills any running process that is on the blocklist.

### `procguard find <name>`
Search the process log for a specific program name from the command line.

### `procguard block`
Manage the process blocklist from the command line.
- `procguard block add <name>`: Add a program to the blocklist.
- `procguard block rm <name>`: Remove a program from the blocklist.
- `procguard block list`: Show the current blocklist.
- `procguard block clear`: Clear the entire blocklist.
- `procguard block save <file>`: Save the current blocklist to a file.
- `procguard block load <file>`: Load a blocklist from a file.

### `procguard systemd` (Linux only)
Manage the systemd user service for the ProcGuard daemon.
- `procguard systemd install`: Install and enable the systemd user service.
- `procguard systemd remove`: Disable and remove the systemd user service.

## Platform-Specific Behavior

-   **Automatic Setup (Windows):** On its first launch, ProcGuard will automatically create an entry in the Windows Task Scheduler to ensure the daemon runs on logon.
-   **Persistent Service (Linux):** For similar persistent behavior on Linux, use the `procguard systemd install` command.
-   **File Blocking:**
    -   On **Windows**, blocking an executable renames it to have a `.blocked` extension.
    -   On **Linux/macOS**, blocking an executable removes its execute permission.

## Configuration

ProcGuard uses a configuration file located at `~/.cache/procguard/spec.json`. As of the latest version, this file's primary role is to track the state of the systemd service on Linux:

```json
{
  "systemd_installed": false
}
```
Other features, such as the Windows autostart task, are now handled automatically by the application and do not require manual configuration.