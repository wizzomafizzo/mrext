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

package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/ini.v1"
)

type LaunchSyncConfig struct{}

type PlayLogConfig struct {
	OnCoreStart string `ini:"on_core_start,omitempty"`
	OnCoreStop  string `ini:"on_core_stop,omitempty"`
	OnGameStart string `ini:"on_game_start,omitempty"`
	OnGameStop  string `ini:"on_game_stop,omitempty"`
	SaveEvery   int    `ini:"save_every,omitempty"`
}

type RandomConfig struct{}

type SearchConfig struct {
	Sort   string   `ini:"sort,omitempty"`
	Filter []string `ini:"filter,omitempty" delim:","`
}

type LastPlayedConfig struct {
	Name                string `ini:"name,omitempty"`
	LastPlayedName      string `ini:"last_played_name,omitempty"`
	RecentFolderName    string `ini:"recent_folder_name,omitempty"`
	DisableLastPlayed   bool   `ini:"disable_last_played,omitempty"`
	DisableRecentFolder bool   `ini:"disable_recent_folder,omitempty"`
}

type RemoteConfig struct {
	CustomLogo      string `ini:"custom_logo,omitempty"`
	AnnounceGameURL string `ini:"announce_game_url,omitempty"`
	MDNSService     bool   `ini:"mdns_service,omitempty"`
	SyncSSHKeys     bool   `ini:"sync_ssh_keys,omitempty"`
}

type FavoritesConfig struct {
	DefaultFolder         string   `ini:"default_folder,omitempty"`
	ExternalFolder        string   `ini:"external_folder,omitempty"`
	CorePrefix            string   `ini:"core_prefix,omitempty"`
	FolderNameContains    []string `ini:"folder_name_contains,omitempty" delim:","`
	GamesFolder           []string `ini:"games_folder,omitempty,allowshadow"`
	CreateDefaultFolder   bool     `ini:"create_default_folder,omitempty"`
	ManageArcadeCoreLinks bool     `ini:"manage_arcade_core_links,omitempty"`
	HideRootFiles         bool     `ini:"hide_root_files,omitempty"`
}

type FavoritesCoresConfig struct {
	All string `ini:"all,omitempty"`
}

type TUIConfig struct {
	Theme            string `ini:"theme,omitempty"`
	Mouse            bool   `ini:"mouse,omitempty"`
	CRTMode          bool   `ini:"crt_mode,omitempty"`
	OnScreenKeyboard bool   `ini:"on_screen_keyboard,omitempty"`
}

type SystemsConfig struct {
	GamesFolder []string `ini:"games_folder,omitempty,allowshadow"`
	SetCore     []string `ini:"set_core,omitempty,allowshadow"`
}

//nolint:govet // Field order keeps application sections grouped for configuration mapping.
type UserConfig struct {
	AppPath        string
	IniPath        string
	LaunchSync     LaunchSyncConfig     `ini:"launchsync,omitempty"`
	PlayLog        PlayLogConfig        `ini:"playlog,omitempty"`
	Random         RandomConfig         `ini:"random,omitempty"`
	Search         SearchConfig         `ini:"search,omitempty"`
	LastPlayed     LastPlayedConfig     `ini:"lastplayed,omitempty"`
	Remote         RemoteConfig         `ini:"remote,omitempty"`
	Favorites      FavoritesConfig      `ini:"favorites,omitempty"`
	FavoritesCores FavoritesCoresConfig `ini:"cores,omitempty"`
	TUI            TUIConfig            `ini:"tui,omitempty"`
	Systems        SystemsConfig        `ini:"systems,omitempty"`
}

func EnsureUserConfig(path, content string) error {
	// #nosec G302 -- MiSTer configuration must remain editable through shared storage.
	file, err := os.OpenFile(filepath.Clean(path), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if os.IsExist(err) {
			return nil
		}
		return fmt.Errorf("create user configuration: %w", err)
	}

	removeIncomplete := true
	defer func() {
		_ = file.Close()
		if removeIncomplete {
			_ = os.Remove(path)
		}
	}()

	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("write user configuration: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close user configuration: %w", err)
	}
	removeIncomplete = false
	return nil
}

func LoadUserConfig(name string, defaultConfig *UserConfig) (*UserConfig, error) {
	iniPath := os.Getenv(UserConfigEnv)

	exePath, err := os.Executable()
	if err != nil {
		return defaultConfig, fmt.Errorf("resolve executable path: %w", err)
	}

	appPath := os.Getenv(UserAppPathEnv)
	if appPath != "" {
		exePath = appPath
	}

	if iniPath == "" {
		iniPath = filepath.Join(filepath.Dir(exePath), name+".ini")
	}
	return LoadUserConfigAt(iniPath, exePath, defaultConfig)
}

func LoadUserConfigAt(iniPath, appPath string, defaultConfig *UserConfig) (*UserConfig, error) {
	defaultConfig.AppPath = appPath
	defaultConfig.IniPath = iniPath

	// Layer the shared [tui] file under the app's own INI. go-ini only assigns
	// fields the file actually declares, so seeding here and mapping the app
	// INI over it gives: app INI, then shared file, then built-in defaults.
	shared, sharedErr := LoadSharedTUI(defaultConfig.TUI)
	if sharedErr != nil {
		return defaultConfig, sharedErr
	}
	defaultConfig.TUI = shared

	// #nosec G703 -- configuration path is explicitly selected by caller.
	if _, statErr := os.Stat(iniPath); statErr != nil {
		if os.IsNotExist(statErr) {
			return defaultConfig, nil
		}
		return defaultConfig, fmt.Errorf("stat user configuration: %w", statErr)
	}

	cfg, err := ini.ShadowLoad(iniPath)
	if err != nil {
		return defaultConfig, fmt.Errorf("load user configuration: %w", err)
	}

	if err := cfg.StrictMapTo(defaultConfig); err != nil {
		return defaultConfig, fmt.Errorf("map user configuration: %w", err)
	}

	return defaultConfig, nil
}
