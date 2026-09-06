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

package bgm

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// DefaultINI is written only when bgm.ini is missing.
const DefaultINI = "[bgm]\nplayback = random\nplaylist = none\nstartup = yes\nplayincore = no\n" +
	"corebootdelay = 0\nbootdelay = 0\nmenuvolume = -1\ndefaultvolume = -1\ndebug = no\n"

const (
	iniSectionBGM = "bgm"
	iniSectionTUI = "tui"
)

const (
	PlaybackRandom   = "random"
	PlaybackLoop     = "loop"
	PlaybackDisabled = "disabled"
	playlistNoneName = "none"
	playlistAllName  = "all"
)

// Playlist is a configured playlist name. The zero value is Python's None,
// written to the file as "none", and is distinct from an empty name.
type Playlist struct {
	name string
	set  bool
}

// NamedPlaylist returns a playlist for a folder name.
func NamedPlaylist(name string) Playlist {
	return Playlist{name: name, set: true}
}

// ParsePlaylist converts a configuration value, mapping "none" to no playlist.
func ParsePlaylist(value string) Playlist {
	if value == playlistNoneName {
		return Playlist{}
	}
	return NamedPlaylist(value)
}

// IsNone reports whether no playlist is configured.
func (p Playlist) IsNone() bool { return !p.set }

// IsAll reports whether the combined "all" playlist is configured.
func (p Playlist) IsAll() bool { return p.set && p.name == playlistAllName }

// Name returns the folder name; empty for no playlist.
func (p Playlist) Name() string { return p.name }

// String renders the playlist as stored in bgm.ini and reported over the
// socket.
func (p Playlist) String() string {
	if !p.set {
		return playlistNoneName
	}
	return p.name
}

// TUIOptions are the optional [tui] settings shared with other mrext apps.
type TUIOptions struct {
	Theme            string
	Mouse            bool
	CRTMode          bool
	OnScreenKeyboard bool
}

// Config is the typed view of bgm.ini with Python's defaults applied.
type Config struct {
	Playback      string
	Playlist      Playlist
	TUI           TUIOptions
	CoreBootDelay float64
	BootDelay     float64
	MenuVolume    int
	DefaultVolume int
	Startup       bool
	PlayInCore    bool
	Debug         bool
}

// DefaultConfig returns the values used when keys are missing.
func DefaultConfig() Config {
	return Config{
		Playback:      PlaybackRandom,
		MenuVolume:    -1,
		DefaultVolume: -1,
		Startup:       true,
		TUI: TUIOptions{
			Theme:            "default",
			Mouse:            true,
			CRTMode:          true,
			OnScreenKeyboard: true,
		},
	}
}

// ShouldChangeVolume mirrors should_change_volume(): both volumes enabled.
func (c *Config) ShouldChangeVolume() bool {
	return c.MenuVolume >= 0 && c.DefaultVolume >= 0
}

// WriteDefaultINI creates the default file when the music folder exists.
func WriteDefaultINI(paths *Paths) error {
	if !fileExists(paths.MusicFolder) {
		return nil
	}
	// #nosec G306 -- MiSTer configuration must remain editable through shared storage.
	if err := os.WriteFile(paths.IniFile, []byte(DefaultINI), 0o644); err != nil {
		return fmt.Errorf("write default BGM configuration: %w", err)
	}
	return nil
}

// LoadConfig mirrors get_ini(): recreate a missing file, then read it with
// fallbacks. Defaults are returned alongside any error.
func LoadConfig(paths *Paths) (Config, error) {
	if _, err := os.Stat(paths.IniFile); err != nil {
		if writeErr := WriteDefaultINI(paths); writeErr != nil {
			return DefaultConfig(), writeErr
		}
	}
	doc, err := ReadINI(paths.IniFile)
	if err != nil {
		return DefaultConfig(), err
	}
	return ConfigFromDocument(doc), nil
}

// ConfigFromDocument applies Python's fallbacks and parsing rules. Values
// Python would have rejected fall back to the default instead of failing.
func ConfigFromDocument(doc *Document) Config {
	cfg := DefaultConfig()
	if value, ok := doc.Get(iniSectionBGM, "playback"); ok {
		cfg.Playback = value
	}
	if value, ok := doc.Get(iniSectionBGM, "playlist"); ok {
		cfg.Playlist = ParsePlaylist(value)
	}
	cfg.Startup = docBool(doc, iniSectionBGM, "startup", cfg.Startup)
	cfg.PlayInCore = docBool(doc, iniSectionBGM, "playincore", cfg.PlayInCore)
	cfg.Debug = docBool(doc, iniSectionBGM, "debug", cfg.Debug)
	cfg.MenuVolume = docInt(doc, iniSectionBGM, "menuvolume", cfg.MenuVolume)
	cfg.DefaultVolume = docInt(doc, iniSectionBGM, "defaultvolume", cfg.DefaultVolume)
	if value, ok := doc.Get(iniSectionBGM, "corebootdelay"); ok {
		if parsed, valid := parsePythonFloat(value); valid {
			cfg.CoreBootDelay = parsed
		}
	}
	if value, ok := doc.Get(iniSectionBGM, "bootdelay"); ok {
		if parsed, err := ParseBootDelay(value); err == nil {
			cfg.BootDelay = parsed
		}
	}
	if value, ok := doc.Get(iniSectionTUI, "theme"); ok && strings.TrimSpace(value) != "" {
		cfg.TUI.Theme = strings.TrimSpace(value)
	}
	cfg.TUI.Mouse = docBool(doc, iniSectionTUI, "mouse", cfg.TUI.Mouse)
	cfg.TUI.CRTMode = docBool(doc, iniSectionTUI, "crt_mode", cfg.TUI.CRTMode)
	cfg.TUI.OnScreenKeyboard = docBool(doc, iniSectionTUI, "on_screen_keyboard", cfg.TUI.OnScreenKeyboard)
	return cfg
}

func docBool(doc *Document, section, key string, fallback bool) bool {
	value, ok := doc.Get(section, key)
	if !ok {
		return fallback
	}
	parsed, valid := parsePythonBool(value)
	if !valid {
		return fallback
	}
	return parsed
}

func docInt(doc *Document, section, key string, fallback int) int {
	value, ok := doc.Get(section, key)
	if !ok {
		return fallback
	}
	parsed, valid := parsePythonInt(value)
	if !valid {
		return fallback
	}
	return parsed
}

// parsePythonBool implements ConfigParser.getboolean's accepted states.
func parsePythonBool(value string) (result, valid bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "yes", "true", "on":
		return true, true
	case "0", "no", "false", "off":
		return false, true
	default:
		return false, false
	}
}

var (
	pythonIntPattern   = regexp.MustCompile(`^[+-]?\d+(?:_\d+)*$`)
	pythonFloatPattern = regexp.MustCompile(
		`^[+-]?(?:\d+(?:_\d+)*)?(?:\.(?:\d+(?:_\d+)*)?)?(?:e[+-]?\d+(?:_\d+)*)?$`,
	)
)

// parsePythonInt accepts what Python's int() accepts for decimal text.
func parsePythonInt(value string) (int, bool) {
	trimmed := strings.TrimSpace(value)
	if !pythonIntPattern.MatchString(trimmed) {
		return 0, false
	}
	parsed, err := strconv.Atoi(strings.ReplaceAll(trimmed, "_", ""))
	if err != nil {
		return 0, false
	}
	return parsed, true
}

// parsePythonFloat accepts finite values Python's float() accepts.
func parsePythonFloat(value string) (float64, bool) {
	trimmed := strings.ToLower(strings.TrimSpace(value))
	if !pythonFloatPattern.MatchString(trimmed) || !strings.ContainsAny(trimmed, "0123456789") {
		return 0, false
	}
	parsed, err := strconv.ParseFloat(strings.ReplaceAll(trimmed, "_", ""), 64)
	if err != nil || math.IsInf(parsed, 0) || math.IsNaN(parsed) {
		return 0, false
	}
	return parsed, true
}

// UpdateINI reads bgm.ini, lets apply change it, and writes it back
// atomically while retaining unknown sections and keys.
func UpdateINI(path string, apply func(*Document)) error {
	doc, err := ReadINI(path)
	if err != nil {
		return err
	}
	apply(doc)
	return doc.WriteFile(path)
}

// SavePlayback persists the playback type chosen on the main screen.
func SavePlayback(path, playback string) error {
	return UpdateINI(path, func(doc *Document) { doc.Set(iniSectionBGM, "playback", playback) })
}

// SavePlaylist persists the playlist chosen on the main screen.
func SavePlaylist(path string, playlist Playlist) error {
	return UpdateINI(path, func(doc *Document) { doc.Set(iniSectionBGM, "playlist", playlist.String()) })
}

// Settings is the staged view edited on the Settings screen.
type Settings struct {
	Theme            string
	CoreBootDelay    float64
	BootDelay        float64
	MenuVolume       int
	DefaultVolume    int
	Startup          bool
	PlayInCore       bool
	Debug            bool
	Mouse            bool
	CRTMode          bool
	OnScreenKeyboard bool
}

// SettingsFromConfig copies the editable values out of a configuration.
func SettingsFromConfig(cfg *Config) Settings {
	return Settings{
		Theme:            cfg.TUI.Theme,
		CoreBootDelay:    cfg.CoreBootDelay,
		BootDelay:        cfg.BootDelay,
		MenuVolume:       cfg.MenuVolume,
		DefaultVolume:    cfg.DefaultVolume,
		Startup:          cfg.Startup,
		PlayInCore:       cfg.PlayInCore,
		Debug:            cfg.Debug,
		Mouse:            cfg.TUI.Mouse,
		CRTMode:          cfg.TUI.CRTMode,
		OnScreenKeyboard: cfg.TUI.OnScreenKeyboard,
	}
}

// ApplyTo copies the staged values into a configuration.
func (s *Settings) ApplyTo(cfg *Config) {
	cfg.TUI.Theme = s.Theme
	cfg.CoreBootDelay = s.CoreBootDelay
	cfg.BootDelay = s.BootDelay
	cfg.MenuVolume = s.MenuVolume
	cfg.DefaultVolume = s.DefaultVolume
	cfg.Startup = s.Startup
	cfg.PlayInCore = s.PlayInCore
	cfg.Debug = s.Debug
	cfg.TUI.Mouse = s.Mouse
	cfg.TUI.CRTMode = s.CRTMode
	cfg.TUI.OnScreenKeyboard = s.OnScreenKeyboard
}

// Equal reports whether two staged settings match.
func (s *Settings) Equal(other *Settings) bool {
	return *s == *other
}

// Validate rejects values the service could not use.
func (s *Settings) Validate() error {
	if strings.TrimSpace(s.Theme) == "" {
		return errors.New("tui.theme must not be empty")
	}
	if s.CoreBootDelay < 0 || math.IsInf(s.CoreBootDelay, 0) || math.IsNaN(s.CoreBootDelay) {
		return errors.New("bgm.corebootdelay must be zero or a positive number of seconds")
	}
	if _, err := bootDelayDuration(s.BootDelay); err != nil {
		return err
	}
	for name, volume := range map[string]int{"menuvolume": s.MenuVolume, "defaultvolume": s.DefaultVolume} {
		if volume < -1 || volume > 7 {
			return fmt.Errorf("bgm.%s must be between -1 and 7", name)
		}
	}
	return nil
}

// SaveSettings writes the staged values into bgm.ini, keeping every other
// key and section untouched.
func SaveSettings(path string, settings *Settings) error {
	if err := settings.Validate(); err != nil {
		return err
	}
	return UpdateINI(path, func(doc *Document) {
		doc.Set(iniSectionBGM, "startup", formatYesNo(settings.Startup))
		doc.Set(iniSectionBGM, "playincore", formatYesNo(settings.PlayInCore))
		doc.Set(iniSectionBGM, "corebootdelay", FormatDelay(settings.CoreBootDelay))
		doc.Set(iniSectionBGM, "bootdelay", FormatDelay(settings.BootDelay))
		doc.Set(iniSectionBGM, "menuvolume", strconv.Itoa(settings.MenuVolume))
		doc.Set(iniSectionBGM, "defaultvolume", strconv.Itoa(settings.DefaultVolume))
		doc.Set(iniSectionBGM, "debug", formatYesNo(settings.Debug))
		doc.Set(iniSectionTUI, "theme", settings.Theme)
		doc.Set(iniSectionTUI, "mouse", formatYesNo(settings.Mouse))
		doc.Set(iniSectionTUI, "crt_mode", formatYesNo(settings.CRTMode))
		doc.Set(iniSectionTUI, "on_screen_keyboard", formatYesNo(settings.OnScreenKeyboard))
	})
}

// FormatDelay renders a delay without trailing zeros.
func FormatDelay(seconds float64) string {
	return strconv.FormatFloat(seconds, 'f', -1, 64)
}

// ParseDelay validates text entered for the core boot delay.
func ParseDelay(value string) (float64, error) {
	parsed, valid := parsePythonFloat(value)
	if !valid || parsed < 0 {
		return 0, errors.New("enter zero or a positive number of seconds, such as 0 or 1.5")
	}
	return parsed, nil
}

// ParseBootDelay validates startup seconds, including the timer's duration limit.
func ParseBootDelay(value string) (float64, error) {
	parsed, err := ParseDelay(value)
	if err != nil {
		return 0, err
	}
	if _, err = bootDelayDuration(parsed); err != nil {
		return 0, err
	}
	return parsed, nil
}

func bootDelayDuration(seconds float64) (time.Duration, error) {
	if seconds < 0 || math.IsNaN(seconds) || seconds >= float64(math.MaxInt64)/float64(time.Second) {
		return 0, errors.New("bgm.bootdelay must be finite, nonnegative, and within the supported timer range")
	}
	return time.Duration(seconds * float64(time.Second)), nil
}

func formatYesNo(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}

// EnsureMusicFolder creates the music folder when missing and reports
// whether it did so.
func EnsureMusicFolder(paths *Paths) (bool, error) {
	if fileExists(paths.MusicFolder) {
		return false, nil
	}
	// #nosec G301 -- MiSTer's shared storage keeps folders world accessible.
	if err := os.MkdirAll(filepath.Clean(paths.MusicFolder), 0o755); err != nil {
		return false, fmt.Errorf("create music folder: %w", err)
	}
	return true, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
