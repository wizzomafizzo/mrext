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
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	s "strings"
	"time"

	mglgen "github.com/ZaparooProject/zaparoo-core/mister/mgl"
	"github.com/wizzomafizzo/mrext/pkg/config"
	"github.com/wizzomafizzo/mrext/pkg/games"
	"github.com/wizzomafizzo/mrext/pkg/input"
	"github.com/wizzomafizzo/mrext/pkg/utils"
)

func GenerateMgl(cfg *config.UserConfig, system *games.System, path, override string) (string, error) {
	if system == nil {
		return "", errors.New("no system supplied for MGL generation")
	}

	core := games.CatalogCore(system)

	// Preserve legacy per-system RBF overrides.
	for _, setCore := range cfg.Systems.SetCore {
		parts := s.SplitN(setCore, ":", 2)
		if len(parts) == 2 && s.EqualFold(parts[0], system.Id) {
			core.RBF = parts[1]
			break
		}
	}

	mgl, err := mglgen.Generate(&core, core.RBF, path, override)
	if err != nil {
		return "", fmt.Errorf("generate MGL: %w", err)
	}
	return mgl, nil
}

func writeTempFile(content string) (string, error) {
	// #nosec G306 -- MiSTer main must read this cross-process launcher file.
	if err := os.WriteFile(config.LastLaunchFile, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("write temporary launcher: %w", err)
	}
	return config.LastLaunchFile, nil
}

func launchFile(path string) error {
	_, err := os.Stat(config.CmdInterface)
	if err != nil {
		return fmt.Errorf("command interface not accessible: %w", err)
	}

	lowerPath := s.ToLower(path)
	if !s.HasSuffix(lowerPath, ".mgl") &&
		!s.HasSuffix(lowerPath, ".mra") &&
		!s.HasSuffix(lowerPath, ".rbf") {
		return fmt.Errorf("not a valid launch file: %s", path)
	}

	cmd, err := os.OpenFile(config.CmdInterface, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("open command interface: %w", err)
	}
	defer func() { _ = cmd.Close() }()

	if _, err := fmt.Fprintf(cmd, "load_core %s\n", path); err != nil {
		return fmt.Errorf("write launch command: %w", err)
	}
	return nil
}

func launchTempMgl(cfg *config.UserConfig, system *games.System, path string) error {
	override, err := games.RunSystemHook(cfg, system, path)
	if err != nil {
		return fmt.Errorf("run system hook: %w", err)
	}

	mgl, err := GenerateMgl(cfg, system, path, override)
	if err != nil {
		return err
	}

	tmpFile, err := writeTempFile(mgl)
	if err != nil {
		return err
	}

	return launchFile(tmpFile)
}

// LaunchShortCore attempts to launch a core with a short path, as per what's
// allowed in an MGL file.
func LaunchShortCore(path string) error {
	mgl := fmt.Sprintf(
		"<mistergamedescription>\n\t<rbf>%s</rbf>\n</mistergamedescription>\n",
		path,
	)

	tmpFile, err := writeTempFile(mgl)
	if err != nil {
		return err
	}

	return launchFile(tmpFile)
}

func LaunchGame(cfg *config.UserConfig, system *games.System, path string) error {
	switch s.ToLower(filepath.Ext(path)) {
	case ".mra":
		err := launchFile(path)
		if err != nil {
			return err
		}
	case ".mgl":
		err := launchFile(path)
		if err != nil {
			return err
		}

		if ActiveGameEnabled() {
			if err := SetActiveGame(path); err != nil {
				return err
			}
		}
	default:
		err := launchTempMgl(cfg, system, path)
		if err != nil {
			return err
		}

		if ActiveGameEnabled() {
			if err := SetActiveGame(path); err != nil {
				return err
			}
		}
	}

	return nil
}

func GetLauncherFilename(system *games.System, folder, name string) string {
	if system.Id == "Arcade" {
		return filepath.Join(folder, name+".mra")
	}
	return filepath.Join(folder, name+".mgl")
}

func TrySetupArcadeCoresLink(path string) error {
	folder, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat launcher directory: %w", err)
	}
	if !folder.IsDir() {
		return fmt.Errorf("parent is not a directory: %s", path)
	}

	coresLinkPath := filepath.Join(path, filepath.Base(config.ArcadeCoresFolder))
	coresLink, err := os.Lstat(coresLinkPath)

	coresLinkExists := false
	switch {
	case err == nil:
		if coresLink.Mode()&os.ModeSymlink == 0 {
			// cores exists but it's not a symlink. not touching this!
			return nil
		}
		coresLinkExists = true
	case os.IsNotExist(err):
		coresLinkExists = false
	default:
		return fmt.Errorf("inspect arcade cores link: %w", err)
	}

	files, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("read launcher directory: %w", err)
	}

	mraCount := 0
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if s.HasSuffix(s.ToLower(file.Name()), ".mra") {
			mraCount++
		}
	}

	if mraCount > 0 && !coresLinkExists {
		err = os.Symlink(config.ArcadeCoresFolder, coresLinkPath)
		if err != nil {
			return fmt.Errorf("create arcade cores link: %w", err)
		}
	} else if mraCount == 0 && coresLinkExists {
		err = os.Remove(coresLinkPath)
		if err != nil {
			return fmt.Errorf("remove arcade cores link: %w", err)
		}
	}

	return nil
}

func CreateLauncher(
	cfg *config.UserConfig,
	system *games.System,
	gameFile, folder, name string,
) (string, error) {
	if system == nil {
		return "", errors.New("no system specified")
	}

	if system.Id == "Arcade" {
		mraPath := GetLauncherFilename(system, folder, name)
		if _, err := os.Lstat(mraPath); err == nil {
			if err := os.Remove(mraPath); err != nil {
				return "", fmt.Errorf("remove existing game link: %w", err)
			}
		}

		if err := os.Symlink(gameFile, mraPath); err != nil {
			return "", fmt.Errorf("create game link: %w", err)
		}

		if err := TrySetupArcadeCoresLink(filepath.Dir(mraPath)); err != nil {
			return "", fmt.Errorf("set up arcade cores link: %w", err)
		}

		return mraPath, nil
	}

	mglPath := GetLauncherFilename(system, folder, name)
	override, err := games.RunSystemHook(cfg, system, gameFile)
	if err != nil {
		return "", fmt.Errorf("run system hook: %w", err)
	}

	mgl, err := GenerateMgl(cfg, system, gameFile, override)
	if err != nil {
		return "", err
	}

	// #nosec G306,G703 -- generated MGL launchers must remain world-readable.
	if err := os.WriteFile(mglPath, []byte(mgl), 0o644); err != nil {
		return "", fmt.Errorf("write MGL file: %w", err)
	}

	return mglPath, nil
}

// LaunchCore Launch a core given a possibly partial path, as per MGL files.
func LaunchCore(cfg *config.UserConfig, system *games.System) error {
	if _, err := os.Stat(config.CmdInterface); err != nil {
		return fmt.Errorf("command interface not accessible: %w", err)
	}

	if system.SetName != "" {
		return LaunchGame(cfg, system, "")
	}

	rbf, ok := games.SystemsWithRBF()[system.Id]
	if !ok {
		return fmt.Errorf("no core found for system %s", system.Id)
	}

	cmd, err := os.OpenFile(config.CmdInterface, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("open command interface: %w", err)
	}
	defer func() { _ = cmd.Close() }()

	if _, err := fmt.Fprintf(cmd, "load_core %s\n", rbf.Path); err != nil {
		return fmt.Errorf("write launch command: %w", err)
	}
	return nil
}

func LaunchMenu() error {
	if _, err := os.Stat(config.CmdInterface); err != nil {
		return fmt.Errorf("command interface not accessible: %w", err)
	}

	cmd, err := os.OpenFile(config.CmdInterface, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("open command interface: %w", err)
	}
	defer func() { _ = cmd.Close() }()

	// TODO: don't hardcode here
	if _, err := fmt.Fprintf(cmd, "load_core %s\n", filepath.Join(config.SdFolder, "menu.rbf")); err != nil {
		return fmt.Errorf("write menu launch command: %w", err)
	}
	return nil
}

// LaunchGenericFile Given a generic file path, launch it using the correct method, if possible.
func LaunchGenericFile(cfg *config.UserConfig, path string) error {
	var err error
	isGame := false
	ext := s.ToLower(filepath.Ext(path))
	switch ext {
	case ".mra":
		err = launchFile(path)
		if err != nil {
			return err
		}
	case ".mgl":
		err = launchFile(path)
		if err != nil {
			return err
		}
		isGame = true
	case ".rbf":
		err = launchFile(path)
		if err != nil {
			return err
		}
	default:
		system, err := games.BestSystemMatch(cfg, path)
		if err != nil {
			return fmt.Errorf("unknown file type: %s", ext)
		}

		err = launchTempMgl(cfg, &system, path)
		if err != nil {
			return err
		}
		isGame = true
	}

	if ActiveGameEnabled() && isGame {
		err := SetActiveGame(path)
		if err != nil {
			return err
		}
	}

	return nil
}

// TryPickRandomGame recursively searches through given folder for a valid game
// file for that system.
func TryPickRandomGame(system *games.System, folder string) (string, error) {
	files, err := os.ReadDir(folder)
	if err != nil {
		return "", fmt.Errorf("read game directory: %w", err)
	}

	if len(files) == 0 {
		return "", fmt.Errorf("no files in %s", folder)
	}

	var validFiles []os.DirEntry
	for _, file := range files {
		if file.IsDir() || utils.IsZip(file.Name()) || games.MatchSystemFile(system, file.Name()) {
			validFiles = append(validFiles, file)
		}
	}

	if len(validFiles) == 0 {
		return "", fmt.Errorf("no valid files in %s", folder)
	}

	file, err := utils.RandomElem(validFiles)
	if err != nil {
		return "", fmt.Errorf("pick random game entry: %w", err)
	}

	path := filepath.Join(folder, file.Name())
	switch {
	case file.IsDir():
		return TryPickRandomGame(system, path)
	case utils.IsZip(path):
		zipFiles, err := utils.ListZip(path)
		if err != nil {
			return "", fmt.Errorf("list ZIP contents: %w", err)
		}
		if len(zipFiles) == 0 {
			return "", fmt.Errorf("no files in %s", path)
		}

		randomZip, err := utils.RandomElem(zipFiles)
		if err != nil {
			return "", fmt.Errorf("pick random ZIP entry: %w", err)
		}
		zipPath := filepath.Join(path, randomZip)
		if !games.MatchSystemFile(system, zipPath) {
			return "", fmt.Errorf("invalid file picked in %s", path)
		}
		return zipPath, nil
	default:
		return path, nil
	}
}

func LaunchRandomGame(cfg *config.UserConfig, systems []games.System) error {
	const maxTries = 100

	populated := games.GetPopulatedGamesFolders(cfg, systems)
	if len(populated) == 0 {
		return errors.New("no populated games folders found")
	}

	for range maxTries {
		systemID, err := utils.RandomElem(utils.MapKeys(populated))
		if err != nil {
			return fmt.Errorf("pick random system: %w", err)
		}

		folders := populated[systemID]
		var files []string
		for _, folder := range folders {
			results, fileErr := games.GetFiles(systemID, folder)
			if fileErr != nil {
				return fmt.Errorf("list games for %s: %w", systemID, fileErr)
			}
			files = append(files, results...)
		}

		if len(files) == 0 {
			continue
		}

		system, err := games.GetSystem(systemID)
		if err != nil {
			return fmt.Errorf("get system %s: %w", systemID, err)
		}

		game, err := utils.RandomElem(files)
		if err != nil {
			return fmt.Errorf("pick random game: %w", err)
		}

		return LaunchGame(cfg, system, game)
	}

	return errors.New("failed to find a random game")
}

func triggerHTTPGet(rawURL string) error {
	// #nosec G107 -- URL is an explicit launch-token action.
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, rawURL, http.NoBody)
	if err != nil {
		return fmt.Errorf("create HTTP request: %w", err)
	}

	go func() {
		client := &http.Client{Timeout: 15 * time.Second}
		resp, requestErr := client.Do(req)
		if requestErr == nil {
			_ = resp.Body.Close()
		}
	}()
	return nil
}

func LaunchToken(cfg *config.UserConfig, manual bool, kbd input.Keyboard, text string) error {
	// detection can never be perfect, but these characters are illegal in
	// windows filenames and heavily avoided in linux. use them to mark that
	// this is a command
	if s.HasPrefix(text, "**") {
		text = s.TrimPrefix(text, "**")
		parts := s.SplitN(text, ":", 2)
		if len(parts) < 2 {
			return fmt.Errorf("invalid command: %s", text)
		}

		cmd, args := s.TrimSpace(parts[0]), s.TrimSpace(parts[1])

		// TODO: search game file
		// TODO: game file by hash

		switch cmd {
		case "system":
			if s.EqualFold(args, "menu") {
				return LaunchMenu()
			}

			system, err := games.LookupSystem(args)
			if err != nil {
				return fmt.Errorf("look up system: %w", err)
			}

			return LaunchCore(cfg, system)
		case "command":
			if !manual {
				return errors.New("commands must be manually run")
			}

			// #nosec G204 -- shell command is permitted only for explicit manual tokens.
			command := exec.CommandContext(context.Background(), "bash", "-c", args)
			err := command.Start()
			if err != nil {
				return fmt.Errorf("start manual command: %w", err)
			}

			return nil
		case "random":
			if args == "" {
				return errors.New("no system specified")
			}

			if args == "all" {
				return LaunchRandomGame(cfg, games.AllSystems())
			}

			// TODO: allow multiple systems
			system, err := games.LookupSystem(args)
			if err != nil {
				return fmt.Errorf("look up random-game system: %w", err)
			}

			return LaunchRandomGame(cfg, []games.System{*system})
		case "ini":
			inis, err := GetAllMisterIni()
			if err != nil {
				return fmt.Errorf("list MiSTer INI files: %w", err)
			}

			if len(inis) == 0 {
				return errors.New("no ini files found")
			}

			id, err := strconv.Atoi(args)
			if err != nil {
				return fmt.Errorf("parse INI ID: %w", err)
			}

			if id < 1 || id > len(inis) {
				return fmt.Errorf("ini id out of range: %d", id)
			}

			return SetActiveIni(id, true)
		case "get":
			return triggerHTTPGet(args)
		case "key":
			code, err := strconv.Atoi(args)
			if err != nil {
				return fmt.Errorf("parse key code: %w", err)
			}

			if err := kbd.Press(code); err != nil {
				return fmt.Errorf("press key: %w", err)
			}
			return nil
		case "coinp1":
			amount, err := strconv.Atoi(args)
			if err != nil {
				return fmt.Errorf("parse player-one coin count: %w", err)
			}

			for range amount {
				if err := kbd.Press(6); err != nil {
					return fmt.Errorf("press player-one coin key: %w", err)
				}
				time.Sleep(100 * time.Millisecond)
			}

			return nil
		case "coinp2":
			// TODO: this is lazy, make a function
			amount, err := strconv.Atoi(args)
			if err != nil {
				return fmt.Errorf("parse player-two coin count: %w", err)
			}

			for range amount {
				if err := kbd.Press(7); err != nil {
					return fmt.Errorf("press player-two coin key: %w", err)
				}
				time.Sleep(100 * time.Millisecond)
			}

			return nil
		default:
			return fmt.Errorf("unknown command: %s", cmd)
		}
	}

	// if it's not a command, assume it's some kind of file path
	if filepath.IsAbs(text) {
		return LaunchGenericFile(cfg, text)
	}

	// if it's a relative path with no extension, assume it's a core
	if filepath.Ext(text) == "" {
		return LaunchShortCore(text)
	}

	// if the file is in a .zip, just check .zip exists in each games folder
	parts := s.Split(text, "/")
	for i, part := range parts {
		if s.HasSuffix(s.ToLower(part), ".zip") {
			zipPath := filepath.Join(parts[:i+1]...)
			for _, folder := range games.GetGamesFolders(cfg) {
				if _, err := os.Stat(filepath.Join(folder, zipPath)); err == nil {
					return LaunchGenericFile(cfg, filepath.Join(folder, text))
				}
			}
			break
		}
	}

	// then try check for the whole path in each game folder
	for _, folder := range games.GetGamesFolders(cfg) {
		path := filepath.Join(folder, text)
		if _, err := os.Stat(path); err == nil {
			return LaunchGenericFile(cfg, path)
		}
	}

	return fmt.Errorf("could not find file: %s", text)
}

func RelaunchIfInMenu() error {
	if _, err := os.Stat(config.CoreNameFile); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("stat active core name: %w", err)
	}

	name, err := os.ReadFile(config.CoreNameFile)
	if err != nil || string(name) == config.MenuCore {
		return LaunchMenu()
	}
	return nil
}
