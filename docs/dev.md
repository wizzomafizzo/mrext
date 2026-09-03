# Developer Guide

> [!IMPORTANT]
> This guide documents legacy mrext development. Before using or extending mrext for new MiSTer work, read [MIGRATE.md](../MIGRATE.md) to see whether current Zaparoo features provide a maintained replacement. mrext can still be appropriate where the migration guide records a gap, especially for stock-MiSTer workflows that do not use Zaparoo Core.

MiSTer Extensions is a single Go project that outputs multiple individual binary applications. The goal of the project is to create a unified library to manage all aspects of a MiSTer system, and offer a set of modular applications that create a rich user experience for the MiSTer userspace.

Applications should:
- Be usable with only a controller for at least the core functionality
- Be installable by copying a single binary and running it from the Scripts menu
- Not require any external dependencies, including applications shipped with the MiSTer Linux image
- Minimise writes to the SD card and disk space usage
- Minimise polluting the filesystem so they're easy to uninstall
- Minimise CPU usage when running as a daemon

## Development Environment

The project is written in pure Go and uses Mage for build scripts. Development and MiSTer ARM32 cross-compilation can run on any platform supported by Go; no C compiler, native libraries, ARM container, or Docker installation is required.

Most applications use a lot of MiSTer-specific paths and files to function. They will mostly work on a desktop with a `/media/fat` directory created to match a MiSTer system, but this generally won't work great beyond specific testing. The usual development cycle is to build a MiSTer ARM binary, copy it to your own MiSTer and run on there to test.

### Dependencies

- [Go](https://go.dev/)
  
  The whole meat of the project. Version 1.27.1 or newer.

- [Mage](https://magefile.org/)

  Used for all builds and automations in the project. Easiest way to get it running is install the binary somewhere globally, rather than installing via the Go package manager as it recommends.

### Optional Dependencies

- [Python](https://www.python.org/)

  Used for some scripts and older projects. Remember that MiSTer currently ships with version 3.9, so don't use any newer Python features.

## Building

To start, you can run `go mod download` from the root of the project folder. This will download all dependencies used by the project. Builds automatically do this, but running it now will stop your editor from complaining about missing modules.

All build steps are done with the `mage` command run from the root of the project folder. Run `mage` by itself to see a list of available commands. Build, test, and systems-documentation targets first generate an ignored metadata asset from a pinned Zaparoo Core Git revision. The first run requires network access and Git; later runs reuse the generated asset and cached checkout. Set `ZAPAROO_CORE_SOURCE` to a local Core checkout when developing metadata changes.

Built binaries will be created in the `_bin` directory under the appropriate architecture subdirectory.

Check the `apps` variable for a list of application target names near the top of the `magefile.go` file. These are the targets used for the commands below. Usually they should match the application folder name in the `cmd` folder.

These are the important commands:

- `mage build <target>`

  Builds a binary of the target application for the current system.

- `mage mister <target>`

  Cross-compiles a static Linux ARMv7 binary for MiSTer with `CGO_ENABLED=0`.

- `mage release <target>`

  Builds a binary of the target application for MiSTer, copies it to the appropriate folder in `releases`, generates an updated `<target>.json` repo file for use with `update` and `update_all` on MiSTer and updates the combined `all.json` repo file.

Binary releases all go in the `releases` folder.

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

#### curses

For showing a GUI/TUI using curses. These should be modular as much as possible and shareable between applications.

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

### scripts

Various support scripts for project.
