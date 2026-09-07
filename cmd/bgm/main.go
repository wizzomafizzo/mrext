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
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/wizzomafizzo/mrext/pkg/bgm"
	"github.com/wizzomafizzo/mrext/pkg/version"
)

const (
	commandExec    = "exec"
	commandStart   = "start"
	commandStop    = "stop"
	commandRestart = "restart"
	usage          = "usage: bgm.sh [--root PATH] [exec|start|stop|restart]"
)

// serviceTimeout bounds how long the CLI waits for the service socket to
// appear or disappear; tests shorten it.
var serviceTimeout = 5 * time.Second

type cliOptions struct {
	root        string
	command     string
	showVersion bool
}

func parseCLI(args []string) (cliOptions, error) {
	flags := flag.NewFlagSet("bgm", flag.ContinueOnError)
	root := flags.String("root", "", "use a local MiSTer filesystem root")
	showVersion := flags.Bool("version", false, "print the version and exit")
	if err := flags.Parse(args); err != nil {
		return cliOptions{}, fmt.Errorf("parse arguments: %w", err)
	}
	positionals := flags.Args()
	if len(positionals) > 1 {
		return cliOptions{}, errors.New(usage)
	}
	options := cliOptions{root: *root, showVersion: *showVersion}
	if len(positionals) == 1 {
		switch positionals[0] {
		case commandExec, commandStart, commandStop, commandRestart:
			options.command = positionals[0]
		default:
			return cliOptions{}, errors.New(usage)
		}
	}
	return options, nil
}

func loadRuntime(root string) (bgm.Paths, error) {
	if root == "" {
		return bgm.DefaultPaths(), nil
	}
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return bgm.Paths{}, fmt.Errorf("resolve local MiSTer root: %w", err)
	}
	paths := bgm.RootedPaths(absoluteRoot)
	folders := []string{
		absoluteRoot, paths.TempFolder, filepath.Dir(paths.CmdInterface), filepath.Dir(paths.StartupScript),
	}
	for _, folder := range folders {
		// #nosec G301 -- local fixture mirrors MiSTer's shared filesystem permissions.
		if mkdirErr := os.MkdirAll(folder, 0o755); mkdirErr != nil {
			return bgm.Paths{}, fmt.Errorf("create local MiSTer root: %w", mkdirErr)
		}
	}
	return paths, nil
}

// app wires the CLI commands together. spawn and runTUI are replaced by tests.
type app struct {
	logger *bgm.Logger
	spawn  func() error
	runTUI func(cfg *bgm.Config) error
	paths  bgm.Paths
	root   string
}

func newApp(options cliOptions, stdout io.Writer) (*app, error) {
	paths, err := loadRuntime(options.root)
	if err != nil {
		return nil, err
	}
	application := &app{
		logger: bgm.NewLoggerTo(&paths, stdout),
		paths:  paths,
		root:   options.root,
	}
	application.spawn = func() error { return bgm.SpawnService(&application.paths, application.root) }
	application.runTUI = func(cfg *bgm.Config) error {
		return newUI(&application.paths, cfg, application.logger).Run()
	}
	return application, nil
}

func run(args []string, stdout io.Writer) error {
	options, err := parseCLI(args)
	if err != nil {
		return err
	}
	if options.showVersion {
		_, _ = fmt.Fprintf(stdout, "bgm %s\n", version.String())
		return nil
	}
	application, err := newApp(options, stdout)
	if err != nil {
		return err
	}
	return application.dispatch(options.command)
}

// dispatch mirrors the Python __main__ block.
func (a *app) dispatch(command string) error {
	cfg, _ := bgm.LoadConfig(&a.paths)
	switch command {
	case commandExec:
		return a.runExec()
	case commandStart:
		return a.start(&cfg)
	case commandStop:
		return stopService(&a.paths, a.logger)
	case commandRestart:
		if err := stopService(&a.paths, a.logger); err != nil {
			return err
		}
		if bgm.SocketExists(&a.paths) && !bgm.WaitForSocketGone(&a.paths, serviceTimeout) {
			return errors.New("BGM service did not stop")
		}
		restarted, _ := bgm.LoadConfig(&a.paths)
		return a.start(&restarted)
	default:
		return a.interactive()
	}
}

func (a *app) start(cfg *bgm.Config) error {
	if !cfg.Startup {
		a.logger.Print("Auto-start is disabled in configuration")
		return nil
	}
	return a.spawn()
}

func (a *app) interactive() error {
	created, err := bgm.EnsureMusicFolder(&a.paths)
	if err != nil {
		return fmt.Errorf("prepare music folder: %w", err)
	}
	if created {
		a.logger.Print("Created music folder.")
	}
	if err := bgm.TryAddToStartup(&a.paths, a.logger); err != nil {
		return fmt.Errorf("configure startup: %w", err)
	}
	cfg, _ := bgm.LoadConfig(&a.paths)
	player := bgm.NewPlayer(&a.paths, a.logger, &cfg)
	if player.TotalTracks(player.Playlist(), true) == 0 {
		a.logger.Print(fmt.Sprintf("Add music files to %s and re-run this script to start.", a.paths.MusicFolder))
		return nil
	}
	if bgm.SocketStale(&a.paths) {
		// A crashed service left its socket behind; the Python script could
		// only be recovered from this with a reboot.
		_ = os.Remove(a.paths.SocketFile)
	}
	if !bgm.SocketExists(&a.paths) {
		a.logger.Print("Starting BGM service...")
		if err := a.spawn(); err != nil {
			return err
		}
		if !bgm.WaitForSocket(&a.paths, serviceTimeout) {
			return nil
		}
	}
	return a.runTUI(&cfg)
}

func (a *app) runExec() error {
	if bgm.SocketExists(&a.paths) {
		if !bgm.SocketStale(&a.paths) {
			a.logger.Print("BGM service is already running, exiting...")
			return nil
		}
		_ = os.Remove(a.paths.SocketFile)
	}
	service := bgm.NewService(&a.paths, a.logger)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		received := <-signals
		a.logger.Log(fmt.Sprintf("Stopping service (%d)", signalNumber(received)))
		a.shutdown(service)
		os.Exit(0)
	}()
	err := service.Run(context.Background())
	a.logger.Log("Stopping service (0)")
	a.shutdown(service)
	if err != nil {
		return fmt.Errorf("run BGM service: %w", err)
	}
	return nil
}

func stopService(paths *bgm.Paths, logger *bgm.Logger) error {
	if err := bgm.StopService(paths, logger); err != nil {
		return fmt.Errorf("stop BGM service: %w", err)
	}
	return nil
}

func (a *app) shutdown(service *bgm.Service) {
	service.Cleanup()
	bgm.RemoveServiceBinary(&a.paths)
}

func signalNumber(received os.Signal) int {
	if number, ok := received.(syscall.Signal); ok {
		return int(number)
	}
	return 0
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
