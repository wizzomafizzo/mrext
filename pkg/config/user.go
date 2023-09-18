package config

import (
	"os"
	"path/filepath"

	"gopkg.in/ini.v1"
)

type LaunchSyncConfig struct{}

type PlayLogConfig struct {
	SaveEvery   int    `ini:"save_every,omitempty"`
	OnCoreStart string `ini:"on_core_start,omitempty"`
	OnCoreStop  string `ini:"on_core_stop,omitempty"`
	OnGameStart string `ini:"on_game_start,omitempty"`
	OnGameStop  string `ini:"on_game_stop,omitempty"`
}

type RandomConfig struct{}

type SearchConfig struct {
	Filter []string `ini:"filter,omitempty" delim:","`
	Sort   string   `ini:"sort,omitempty"`
}

type LastPlayedConfig struct {
	Name                string `ini:"name,omitempty"`
	LastPlayedName      string `ini:"last_played_name,omitempty"`
	DisableLastPlayed   bool   `ini:"disable_last_played,omitempty"`
	RecentFolderName    string `ini:"recent_folder_name,omitempty"`
	DisableRecentFolder bool   `ini:"disable_recent_folder,omitempty"`
}

type RemoteConfig struct {
	MdnsService     bool   `ini:"mdns_service,omitempty"`
	SyncSSHKeys     bool   `ini:"sync_ssh_keys,omitempty"`
	CustomLogo      string `ini:"custom_logo,omitempty"`
	AnnounceGameUrl string `ini:"announce_game_url,omitempty"`
}

type NfcConfig struct {
	ConnectionString string `ini:"connection_string,omitempty"`
	AllowCommands    bool   `ini:"allow_commands,omitempty"`
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
	Nfc        NfcConfig        `ini:"nfc,omitempty"`
	Systems    SystemsConfig    `ini:"systems,omitempty"`
}

func LoadUserConfig(name string, defaultConfig *UserConfig) (*UserConfig, error) {
	iniPath := os.Getenv(UserConfigEnv)

	exePath, err := os.Executable()
	if err != nil {
		return defaultConfig, err
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

	if _, err := os.Stat(iniPath); os.IsNotExist(err) {
		return defaultConfig, nil
	}

	cfg, err := ini.ShadowLoad(iniPath)
	if err != nil {
		return defaultConfig, err
	}

	err = cfg.StrictMapTo(defaultConfig)
	if err != nil {
		return defaultConfig, err
	}

	return defaultConfig, nil
}
