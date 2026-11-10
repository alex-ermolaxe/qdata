# Changelog

All notable changes to qdata are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- Renamed the interactive `INFO` command to `HELP` and added a session-independent `help` tool to the MCP server.

## [2.0.1] - 2026-11-09

### Changed

- Unified command definitions across CLI execution, tab completion, and MCP tool exposure.
- Improved tab completion with command-specific context for command arguments and field paths.

### Fixed

- `SCHEMA` now infers and displays array element types, including after `FLATTEN` and `RESET`.
- Command parsing now handles leading and trailing whitespace correctly.

## [2.0.0] - 2026-11-05

### Added

- MCP server mode (`qdata mcp-serve`) for integrating with AI assistants such as GitHub Copilot, Cursor, and Claude Desktop. Includes tools for loading, filtering, selecting and excluding fields, sorting, transforming, and saving data, with support for multiple files through named sessions.
- README documentation for connecting to MCP clients.

## [1.1.5] - 2026-11-04

### Added

- Interactive CLI for querying, filtering, and transforming structured data (JSON, XML, and CSV) with SQL-like syntax.
- Commands: WHERE, SELECT, EXCLUDE, SORT, SHOW, SAVE, FLATTEN, TRANSFORM, RESET, COUNT, SCHEMA, INFO, EXIT, and VERSION.
- Tab completion for keywords, field names, and nested paths.

[Unreleased]: https://github.com/alex-ermolaxe/qdata/compare/v2.0.1...HEAD
[2.0.1]: https://github.com/alex-ermolaxe/qdata/compare/v2.0.0...v2.0.1
[2.0.0]: https://github.com/alex-ermolaxe/qdata/compare/v1.1.5...v2.0.0
[1.1.5]: https://github.com/alex-ermolaxe/qdata/releases/tag/v1.1.5
