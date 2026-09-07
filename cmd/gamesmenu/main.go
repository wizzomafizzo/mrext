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
	"github.com/wizzomafizzo/mrext/pkg/gamesmenu"
	"github.com/wizzomafizzo/mrext/pkg/version"
)

type cliOptions struct {
	root        string
	showVersion bool
}

func parseCLI(args []string) (cliOptions, error) {
	flags := flag.NewFlagSet("gamesmenu", flag.ContinueOnError)
	root := flags.String("root", "", "use a local MiSTer filesystem root")
	showVersion := flags.Bool("version", false, "print the version and exit")
	if err := flags.Parse(args); err != nil {
		return cliOptions{}, fmt.Errorf("parse arguments: %w", err)
	}
	if flags.NArg() != 0 {
		return cliOptions{}, errors.New("usage: gamesmenu.sh [--root PATH]")
	}
	return cliOptions{root: *root, showVersion: *showVersion}, nil
}

func loadRuntime(root string) (*config.UserConfig, *gamesmenu.Manager, error) {
	var cfg *config.UserConfig
	var err error
	paths := gamesmenu.DefaultPaths()
	if root == "" {
		cfg, err = gamesmenu.LoadConfig()
	} else {
		absolute, absErr := filepath.Abs(root)
		if absErr != nil {
			return nil, nil, fmt.Errorf("resolve local MiSTer root: %w", absErr)
		}
		scripts := filepath.Join(absolute, "Scripts")
		// #nosec G301 -- local fixture mirrors MiSTer's shared filesystem permissions.
		if mkdirErr := os.MkdirAll(scripts, 0o755); mkdirErr != nil {
			return nil, nil, fmt.Errorf("create local MiSTer root: %w", mkdirErr)
		}
		cfg, err = gamesmenu.LoadConfigAt(
			filepath.Join(scripts, gamesmenu.IniFilename), filepath.Join(scripts, gamesmenu.AppFilename),
		)
		if err == nil {
			cfg.Systems.GamesFolder = append([]string{absolute}, cfg.Systems.GamesFolder...)
		}
		paths = gamesmenu.RootedPaths(absolute)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("load GamesMenu configuration: %w", err)
	}
	manager, err := gamesmenu.NewManagerWithPaths(cfg, paths)
	if err != nil {
		return nil, nil, fmt.Errorf("create GamesMenu manager: %w", err)
	}
	return cfg, manager, nil
}

func run(args []string) error {
	options, err := parseCLI(args)
	if err != nil {
		return err
	}
	if options.showVersion {
		_, _ = fmt.Printf("gamesmenu %s\n", version.String())
		return nil
	}
	cfg, manager, err := loadRuntime(options.root)
	if err != nil {
		return err
	}
	if err := gamesmenu.EnsureConfigFile(cfg); err != nil {
		return fmt.Errorf("create GamesMenu configuration: %w", err)
	}
	return newUI(cfg, manager).Run()
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
