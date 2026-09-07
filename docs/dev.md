# Developer Guide

> [!NOTE]
> This guide covers development of the maintained mrext tools. If you're building a new media application or integration, use the [Zaparoo Core API](https://zaparoo.org/docs/core/api/), even for MiSTer-only projects. See [mrext and Zaparoo](../MIGRATE.md) for API mappings and the MiSTer-specific functions available through mrext.

MiSTer Extensions is a single Go project that outputs multiple individual binary applications. It provides modular tools for MiSTer menu workflows and device management, backed by shared packages rather than duplicate implementations. Zaparoo Core is not required on the device.

Changes to these tools should preserve existing configuration, folder layouts, command-line contracts, and controller-driven workflows. Remote's API remains supported for existing clients and its specialist MiSTer functions. For new media integrations, use Core's daemon rather than building on mrext's internal indexing, tracking, or launch implementation.

Applications should:
- Be usable with only a controller for at least the core functionality
- Be installable by copying a single binary and running it from the Scripts menu
- Not require any external dependencies, including applications shipped with the MiSTer Linux image
- Minimise writes to the SD card and disk space usage
- Minimise polluting the filesystem so they're easy to uninstall
- Minimise CPU usage when running as a daemon

## Development Environment

Applications and shared backend packages are written in pure Go. Remote also embeds a React/TypeScript web UI from `web/remote`. Mage coordinates both build systems. MiSTer ARM32 cross-compilation uses Go directly; no C compiler, native libraries, ARM container, or Docker installation is required.

Most applications use a lot of MiSTer-specific paths and files to function. They will mostly work on a desktop with a `/media/fat` directory created to match a MiSTer system, but this generally won't work great beyond specific testing. The usual development cycle is to build a MiSTer ARM binary, copy it to your own MiSTer and run on there to test.

### Dependencies

- [Go](https://go.dev/)
  
  The whole meat of the project. Version 1.27.1 or newer.

- [Mage](https://magefile.org/)

  Used for builds and automation. Install Mage globally, or replace `mage` in commands below with `go run github.com/magefile/mage`.

- [Node.js](https://nodejs.org/) 24 or newer and npm 12.0.2

  Required to build Remote's embedded web UI. npm version is pinned by `web/remote/package.json` and CI.

### Optional Dependencies

- [Python](https://www.python.org/)

  Used by `scripts/generate_repo.py` in the repository publishing workflow.

## Building

To start, you can run `go mod download` from the root of the project folder. This will download all dependencies used by the project. Builds automatically do this, but running it now will stop your editor from complaining about missing modules.

All build steps are done with the `mage` command run from the root of the project folder. Run `mage` by itself to see a list of available commands. Build, test, lint, coverage, and systems-documentation targets first generate an ignored metadata asset from a pinned Zaparoo Core Git revision. The first run requires network access and Git; later runs reuse the generated asset and cached checkout. Set `ZAPAROO_CORE_SOURCE` to a local Core checkout when developing metadata changes.

Built binaries will be created in the `_bin` directory under the appropriate architecture subdirectory.

Check the `apps` variable for a list of application target names near the top of the `magefile.go` file. These are the targets used for the commands below. Usually they should match the application folder name in the `cmd` folder.

These are the important commands:

- `mage build <target>`

  Builds a binary of the target application for the current system. Building `remote` or `all` first installs and builds the embedded web UI.

- `mage mister <target>`

  Cross-compiles a static Linux ARMv7 binary for MiSTer with `CGO_ENABLED=0`. Building `remote` or `all` also rebuilds its web UI.

- `mage deploy <address>`

  Builds every application for MiSTer and copies the binaries to `/media/fat/Scripts` on the device at `<address>` over SSH as `root`. The address is a plain hostname or IP. Because it builds everything, it also rebuilds Remote's web UI. Stop any running mrext service on the device first if you are replacing a binary it is executing.

- `mage remoteWeb`

  Runs deterministic npm installation and builds `web/remote` into the ignored `cmd/remote/_client/build` directory.

- `mage release <target>`

  Builds a binary of the target application for MiSTer, copies it to the appropriate folder in `releases`, generates an updated `<target>.json` repo file for use with `update` and `update_all` on MiSTer and updates the combined `all.json` repo file.

Binary releases all go in the `releases` folder.

`mage build` and `mage mister` stamp the binary with `git describe` and the short commit, so every app answers `-version`. A plain `go build` leaves the stamp empty and falls back to the VCS information Go records, so the flag still reports something useful. Apps with a full-screen interface also print the version in the bottom border of the page frame.

## Project Layout

This is an overview of all the major files and folders in the project.

### cmd

Each folder in here represents a separate application and is the entry point for each binary. The complexity depends on the application, but as much as possible they should be using the shared library. They cannot depend on or reference each other.

### docs

All application and project documentation and notes. Markdown format is preferred.

### pkg

The shared library for the whole project.

#### config

All global configuration settings, MiSTer environment paths and the module for parsing per-app .ini configuration files. If you're hardcoding a path or a special value, it should go here instead.

#### tui

Reusable terminal UI components built with tview and tcell, including list pickers, an on-screen keyboard, progress views, and MiSTer's framebuffer-console retry flow.

#### games

All functions related to indexing, searching and interacting with game files on a system.

The `systems.go` file combines the dependency-free MiSTer catalog from `github.com/ZaparooProject/zaparoo-core/mister` with display metadata generated directly from pinned Zaparoo Core source files. Zaparoo owns names, categories, release dates, manufacturers, aliases, scan folders, extensions, RBF paths, setnames, groups, MGL slots, and pure MGL generation. mrext retains stock-filesystem scanning, hooks, legacy JSON structures, and menu output. Run `mage generateSystemMetadata` to refresh the ignored build asset and `mage genSystemsDoc` to regenerate `docs/systems.md`.

#### input

For interacting with and impersonating input devices.

#### mister

Functions for interacting with various parts of the MiSTer system. Somewhat of a catch-all for modules that aren't big enough for their own folder. Does things like generating and running MGL files, managing the startup services file and reading the main MiSTer .ini file.

#### utils

Simple generic functions used throughout the project. This is mostly used for common functions that are not present in the Go stdlib for some reason.

### releases

Final binary releases and repo files go here. Automatically generated from build script.

### web

Remote's React/TypeScript web UI lives in `web/remote`. Its original `mrext-client` Git history is retained in this repository. Run `npm ci` and `npm run dev` there for frontend development. Production builds are generated directly into `cmd/remote/_client/build` and embedded in `remote.sh`.

Retired Android and iOS wrappers are available through repository history but are not part of the maintained source tree.

### scripts

Various support scripts for project.
