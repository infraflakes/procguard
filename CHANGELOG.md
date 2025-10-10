Major Refactoring: Database & Logging System
   * Event-Driven Logging: Replaced the inefficient, periodic, file-based process logging with an event-driven system using
     a SQLite database.
   * Intelligent Detection: The daemon now intelligently detects process start and end events, preventing duplicate data, reducing
     disk I/O, and providing a precise lifecycle for every process.
     cross-compilation and runtime errors, simplifying the build process for all platforms.
   * API & CLI Integration: Updated all parts of the application, including the API search and the procguard find command, to query
     the new SQLite database.

Architectural Improvements
   * Decoupled Services: The background daemon is now decoupled from the API server. It uses the internal API client to fetch the
     blocklist, making the components more independent and the codebase easier to maintain.
   * Robust Windows Autostart: Replaced the faulty and permission-sensitive Scheduled Task with a more reliable and standard
     registry-based approach (...CurrentVersion\Run).
   * Refined Process Filtering: Improved the logging logic to be less noisy by ignoring child processes that have the same name as
     their parent (e.g., chrome.exe spawning another chrome.exe), resulting in a cleaner, more user-focused log.
   * Centralized Logic: Consolidated all platform-specific code for autostart and process filtering into single, reusable locations
     with clear build tags, improving code organization.

Bug Fixes & Stability
   * Fixed Critical Runtime Errors:
       * Resolved the database is locked (SQLITE_BUSY) race condition on startup.
       * Fixed the sql: database is closed error by correcting the database connection's lifecycle.
   * Corrected Installation Logic:
       * Fixed a bug where the native messaging manifest (procguard.json) on Windows pointed to the wrong executable path.
       * Fixed a bug where the configuration file (spec.json) was not being updated when the autostart setting was changed.

User Experience
   * New "Start with Windows" Setting: Implemented a new section in the GUI's "Settings" page, allowing you to easily enable or
     disable autostart on Windows.
