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

package mister

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/input"
)

type Script struct {
	Name     string `json:"name"`
	Filename string `json:"filename"`
	Path     string `json:"path"`
}

func IsMenuRunning() bool {
	activeCore, err := GetActiveCoreName()
	if err != nil {
		return false
	}

	return activeCore == config.MenuCore
}

func IsScriptRunning() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := "ps ax | grep /tmp/script | grep -v grep"
	out, err := exec.CommandContext(ctx, "sh", "-c", cmd).Output()
	return err == nil && len(out) > 0
}

func KillActiveScript() error {
	if !IsScriptRunning() {
		return nil
	}

	// TODO: this doesn't actually work right now. it just orphans the launched script process
	// one good idea is to launch scripts with and env variable that contains the pid of the menu
	// so it will get picked up in the grep. it's not urgent though

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := "ps ax | grep /tmp/script | grep -v grep | awk '{print $1}' | xargs kill"
	if err := exec.CommandContext(ctx, "sh", "-c", cmd).Run(); err != nil {
		return fmt.Errorf("kill active script: %w", err)
	}
	return nil
}

func ScriptCanLaunch() bool {
	scriptRunning := IsScriptRunning()

	return IsMenuRunning() && !scriptRunning
}

func OpenConsole(kbd input.Keyboard) error {
	if !IsMenuRunning() {
		return errors.New("cannot open console, active core is not menu")
	}

	getTTY := func() (string, error) {
		sys := "/sys/devices/virtual/tty/tty0/active"
		if _, err := os.Stat(sys); err != nil {
			return "", fmt.Errorf("stat active TTY: %w", err)
		}

		tty, err := os.ReadFile(sys)
		if err != nil {
			return "", fmt.Errorf("read active TTY: %w", err)
		}

		return strings.TrimSpace(string(tty)), nil
	}

	// we use the F9 key as a means to disable main's usage of the framebuffer and allow scripts to run
	// unfortunately when the menu "sleeps", any key press will be eaten by main and not trigger the console switch
	// there's also no simple way to tell if mister has switched to the console
	// so what we do is switch to tty3, which is unused by mister, then attempt to switch to console,
	// which sets tty to 1 on success, then check in a loop if it actually did change to 1 and keep pressing F9
	// until it's switched

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := exec.CommandContext(ctx, "chvt", "3").Run(); err != nil {
		return fmt.Errorf("switch to staging console: %w", err)
	}

	tries := 0
	for {
		if tries > 20 {
			return errors.New("could not switch to tty1")
		}
		if err := kbd.Console(); err != nil {
			return fmt.Errorf("send console shortcut: %w", err)
		}
		time.Sleep(50 * time.Millisecond)
		tty, err := getTTY()
		if err != nil {
			return err
		}
		if tty == "tty1" {
			break
		}
		tries++
	}

	return nil
}

func GetAllScripts() ([]Script, error) {
	scripts := make([]Script, 0)

	files, err := os.ReadDir(config.ScriptsFolder)
	if err != nil {
		return scripts, fmt.Errorf("read scripts directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		fn := file.Name()
		if strings.HasSuffix(strings.ToLower(fn), ".sh") {
			scripts = append(scripts, Script{
				Name:     strings.TrimSuffix(fn, filepath.Ext(fn)),
				Filename: fn,
				Path:     filepath.Join(config.ScriptsFolder, fn),
			})
		}
	}

	return scripts, nil
}

func RunScript(kbd input.Keyboard, path string) error {
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("stat script: %w", err)
	}

	canLaunch := ScriptCanLaunch()
	if !canLaunch {
		return errors.New("script cannot be launched, active core is not menu or script is already running")
	}

	err := OpenConsole(kbd)
	if err != nil {
		return err
	}

	// this is just to follow mister's convention, which reserves tty2 for scripts
	ctx := context.Background()
	err = exec.CommandContext(ctx, "chvt", "2").Run()
	if err != nil {
		return fmt.Errorf("switch to script console: %w", err)
	}

	// this is how mister launches scripts itself
	// TODO: press any key should be configurable
	launcher := fmt.Sprintf(`#!/bin/bash
export LC_ALL=en_US.UTF-8
export HOME=/root
export LESSKEY=/media/fat/linux/lesskey
cd $(dirname %q)
%q
`, path, path)

	// TODO: this is no longer functional, if we still even want it, need to find a way to make it wait for
	//       input but not block in the background like for the random script
	// echo "Press any key to continue"
	// read -n 1 -s -r -p ""

	// #nosec G303,G306 -- fixed executable path is required by MiSTer's script-launch convention.
	err = os.WriteFile("/tmp/script", []byte(launcher), 0o700)
	if err != nil {
		return fmt.Errorf("write script launcher: %w", err)
	}

	err = exec.CommandContext(
		ctx,
		"/sbin/agetty",
		"-a",
		"root",
		"-l",
		"/tmp/script",
		"--nohostname",
		"-L",
		"tty2",
		"linux",
	).Run()
	if err != nil {
		return fmt.Errorf("run script console: %w", err)
	}

	if err := kbd.ExitConsole(); err != nil {
		return fmt.Errorf("exit script console: %w", err)
	}
	return nil
}
