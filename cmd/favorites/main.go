// mrext
// Copyright (c) 2026 mrext contributors.
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This file is part of mrext.
//
// mrext is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// mrext is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with mrext. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/favorites"
	"github.com/wizzomafizzo/mrext/pkg/version"
)

type cliOptions struct {
	root        string
	command     string
	showVersion bool
}

func parseCLI(args []string) (cliOptions, error) {
	flags := flag.NewFlagSet("favorites", flag.ContinueOnError)
	root := flags.String("root", "", "use a local MiSTer filesystem root")
	showVersion := flags.Bool("version", false, "print the version and exit")
	if err := flags.Parse(args); err != nil {
		return cliOptions{}, fmt.Errorf("parse arguments: %w", err)
	}
	positionals := flags.Args()
	if len(positionals) > 1 || len(positionals) == 1 && positionals[0] != "refresh" {
		return cliOptions{}, errors.New("usage: favorites.sh [--root PATH] [refresh]")
	}
	options := cliOptions{root: *root, showVersion: *showVersion}
	if len(positionals) == 1 {
		options.command = positionals[0]
	}
	return options, nil
}

func loadRuntime(root string) (*config.UserConfig, *favorites.Manager, error) {
	if root == "" {
		cfg, err := favorites.LoadConfig()
		if err != nil {
			return nil, nil, fmt.Errorf("load MiSTer configuration: %w", err)
		}
		return cfg, favorites.NewManager(cfg), nil
	}

	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, nil, fmt.Errorf("resolve local MiSTer root: %w", err)
	}
	scriptsFolder := filepath.Join(absoluteRoot, "Scripts")
	// #nosec G301 -- local fixture mirrors MiSTer's shared filesystem permissions.
	if mkdirErr := os.MkdirAll(scriptsFolder, 0o755); mkdirErr != nil {
		return nil, nil, fmt.Errorf("create local MiSTer root: %w", mkdirErr)
	}
	appPath := filepath.Join(scriptsFolder, "favorites.sh")
	cfg, err := favorites.LoadConfigAt(filepath.Join(scriptsFolder, "favorites.ini"), appPath)
	if err != nil {
		return nil, nil, fmt.Errorf("load local configuration: %w", err)
	}
	cfg.Systems.GamesFolder = append([]string{absoluteRoot}, cfg.Favorites.GamesFolder...)
	if cfg.Favorites.ExternalFolder == "/media/usb0" {
		cfg.Favorites.ExternalFolder = filepath.Join(absoluteRoot, "external")
	}
	manager := favorites.NewManagerWithPaths(cfg, favorites.RuntimePaths{
		SDRoot:            absoluteRoot,
		StartupScript:     filepath.Join(absoluteRoot, "linux", "user-startup.sh"),
		ArcadeCoresFolder: filepath.Join(absoluteRoot, "_Arcade", "cores"),
	})
	return cfg, manager, nil
}

func run(args []string) error {
	options, err := parseCLI(args)
	if err != nil {
		return err
	}
	if options.showVersion {
		_, _ = fmt.Printf("favorites %s\n", version.String())
		return nil
	}
	cfg, manager, err := loadRuntime(options.root)
	if err != nil {
		return fmt.Errorf("load Favorites configuration: %w", err)
	}
	if startupErr := manager.TryAddToStartup(); startupErr != nil {
		return fmt.Errorf("configure startup refresh: %w", startupErr)
	}

	if options.command == "refresh" {
		if refreshErr := manager.Refresh(); refreshErr != nil {
			return fmt.Errorf("refresh Favorites: %w", refreshErr)
		}
		return nil
	}
	if configErr := favorites.EnsureConfigFile(cfg); configErr != nil {
		return fmt.Errorf("create Favorites configuration: %w", configErr)
	}

	createdDefault := false
	if cfg.Favorites.CreateDefaultFolder {
		createdDefault, err = manager.CreateDefaultFolder()
		if err != nil {
			return fmt.Errorf("prepare default Favorites folder: %w", err)
		}
	}
	defer func() {
		if cleanupErr := manager.CleanupCreatedDefault(createdDefault); cleanupErr != nil {
			_, _ = fmt.Fprintln(os.Stderr, cleanupErr)
		}
	}()

	// SetupArcadeLinks walks every favourites folder and Refresh stats every
	// favourite, so both run inside the app behind a progress modal rather
	// than as dead air before the first frame.
	view := newUI(cfg, manager)
	view.SetStartupWork(func(report func(string)) error {
		report("Checking arcade core links...")
		if setupErr := manager.SetupArcadeLinks(); setupErr != nil {
			return fmt.Errorf("prepare arcade links: %w", setupErr)
		}
		report("Refreshing favorites...")
		if refreshErr := manager.Refresh(); refreshErr != nil {
			return fmt.Errorf("refresh Favorites: %w", refreshErr)
		}
		return nil
	})
	return view.Run()
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
