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

type SystemsConfig struct {
	GamesFolder []string `ini:"games_folder,omitempty,allowshadow"`
	SetCore     []string `ini:"set_core,omitempty,allowshadow"`
}

type UserConfig struct {
	AppPath    string
	IniPath    string
	LaunchSync LaunchSyncConfig `ini:"launchsync,omitempty"`
	PlayLog    PlayLogConfig    `ini:"playlog,omitempty"`
	Random     RandomConfig     `ini:"random,omitempty"`
	Search     SearchConfig     `ini:"search,omitempty"`
	LastPlayed LastPlayedConfig `ini:"lastplayed,omitempty"`
	Remote     RemoteConfig     `ini:"remote,omitempty"`
	Systems    SystemsConfig    `ini:"systems,omitempty"`
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

	defaultConfig.AppPath = exePath
	defaultConfig.IniPath = iniPath

	// #nosec G703 -- configuration path is explicitly selected by caller or environment.
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
