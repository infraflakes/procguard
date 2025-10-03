# ProcGuard

ProcGuard is a cross-platform tool for monitoring and controlling processes on your system. It is composed of three main components: a daemon, an API server with a web-based GUI, and a command-line interface (CLI).

## Usage

ProcGuard is designed as a modular system. You run the services you need (the API server and the daemon) and then interact with them through the web GUI or the CLI.

### 1. Running the Services

You need to run the `api` and `daemon` services in separate terminal windows.

**To run the API server (which includes the web GUI):**

```bash
procguard run api
```

This will start the server, and you can access the GUI by navigating to `http://127.0.0.1:58141` in your web browser.

**To run the background daemon:**

```bash
procguard run daemon
```

This will start the background process that monitors and blocks applications based on the blocklist provided by the API service.

### 2. Using the Web GUI

Once the `api` service is running, you can open your browser to `http://127.0.0.1:58141` to access the GUI.

**GUI Features:**
- **Search:** Search the process logs for specific application names.
- **Blocklist Management:** View all currently blocked applications, add new ones, and unblock them.

### 3. Using the CLI

The `procguard` command-line tool allows you to interact with the system from your terminal. It communicates with the `api` service, so make sure the API server is running before using these commands.

#### `procguard find <name>`
Search the process log for a specific program name.

#### `procguard block`
Manage the process blocklist.
- `procguard block add <name>`: Add a program to the blocklist.
- `procguard block rm <name>`: Remove a program from the blocklist.
- `procguard block list`: Show the current blocklist.
- `procguard block clear`: Clear the entire blocklist.
- `procguard block save <file>`: Save the current blocklist to a file.
- `procguard block load <file>`: Load a blocklist from a file.

#### `procguard systemd` (Linux only)
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
