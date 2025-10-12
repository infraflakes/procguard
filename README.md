# ProcGuard

ProcGuard is a Windows-based tool for monitoring and controlling processes and web activity on your system. It is composed of a daemon that runs in the background, an API server with a web-based GUI, and a browser extension for web monitoring.

## Features

- **Process Monitoring:** Logs all running processes and their activity.
- **Application Blocking:** Block any application from running.
- **Web Activity Monitoring:** Logs all visited websites.
- **Website Blocking:** Block any website from being accessed.
- **Web-based GUI:** A simple and intuitive web interface to view logs and manage blocklists.
- **Browser Extension:** A Chrome extension for web monitoring and blocking.

## Dependencies

To build and run ProcGuard, you will need:

- **Go:** The programming language used for the backend.
- **`golangci-lint`:** (Optional) For running linters and ensuring code quality.

## Build Guide

To build the application, you can use the provided `Makefile`:

```bash
make build
```

This will generate the main executable `ProcGuardSvc.exe` in the `build/bin` directory.

## Usage

To run the application, simply execute the `ProcGuardSvc.exe` file located in the `build/bin` directory. This will start the background daemon and the API server.

The web GUI will be available at `http://127.0.0.1:58141`.

To install the browser extension, you will need to load it manually in Chrome from the `extension` directory.
