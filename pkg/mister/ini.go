package mister

import (
	"fmt"
	"github.com/wizzomafizzo/mrext/pkg/utils"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/ini.v1"

	"github.com/wizzomafizzo/mrext/pkg/config"
)

var ValidIniKeys = []string{
	"ypbpr",
	"composite_sync",
	"forced_scandoubler",
	"vga_scaler",
	"vga_sog",
	"keyrah_mode",
	"reset_combo",
	"key_menu_as_rgui",
	"video_mode",
	"video_mode_pal",
	"video_mode_ntsc",
	"video_info",
	"vsync_adjust",
	"hdmi_audio_96k",
	"dvi_mode",
	"hdmi_limited",
	"kbd_nomouse",
	"mouse_throttle",
	"bootscreen",
	"vscale_mode",
	"vscale_border",
	"rbf_hide_datecode",
	"menu_pal",
	"bootcore",
	"bootcore_timeout",
	"font",
	"fb_size",
	"fb_terminal",
	"osd_timeout",
	"direct_video",
	"osd_rotate",
	"gamepad_defaults",
	"recents",
	"controller_info",
	"refresh_min",
	"refresh_max",
	"jamma_vid",
	"jamma_pid",
	"sniper_mode",
	"browse_expand",
	"logo",
	"shared_folder",
	"no_merge_vid",
	"no_merge_pid",
	"no_merge_vidpid",
	"custom_aspect_ratio_1",
	"custom_aspect_ratio_2",
	"spinner_vid",
	"spinner_pid",
	"spinner_axis",
	"spinner_throttle",
	"afilter_default",
	"vfilter_default",
	"vfilter_vertical_default",
	"vfilter_scanlines_default",
	"shmask_default",
	"shmask_mode_default",
	"preset_default",
	"log_file_entry",
	"bt_auto_disconnect",
	"bt_reset_before_pair",
	"waitmount",
	"rumble",
	"wheel_force",
	"wheel_range",
	"hdmi_game_mode",
	"vrr_mode",
	"vrr_min_framerate",
	"vrr_max_framerate",
	"vrr_vesa_framerate",
	"video_off",
	"player_1_controller",
	"player_2_controller",
	"player_3_controller",
	"player_4_controller",
	"player_5_controller",
	"player_6_controller",
	"disable_autofire",
	"video_brightness",
	"video_contrast",
	"video_saturation",
	"video_hue",
	"video_gain_offset",
	"hdr",
	"hdr_max_nits",
	"hdr_avg_nits",
	"vga_mode",
	"ntsc_mode",
	"controller_unique_mapping",
}

const MisterIniSection = "MiSTer"

func LoadMisterIni(id int) (int, *ini.File, error) {
	inis, err := ListMisterInis()
	if err != nil {
		return id, nil, err
	}

	if id < 0 || id > len(inis) {
		return id, nil, fmt.Errorf("ini id is out of range")
	}

	if id == 0 {
		activeId, err := GetActiveIni()
		if err != nil {
			return id, nil, err
		}

		if activeId == 0 {
			id = 1
		} else {
			id = activeId
		}
	}

	iniFileInfo := inis[id-1]
	iniPath := iniFileInfo.Path

	if _, err := os.Stat(iniPath); os.IsNotExist(err) {
		return id, nil, err
	}

	iniFile, err := ini.Load(iniPath)
	if err != nil {
		return id, nil, err
	}

	if !iniFile.HasSection(MisterIniSection) {
		return id, nil, fmt.Errorf("mister.ini does not have a [MiSTer] section")
	}

	ini.PrettyFormat = false
	ini.PrettyEqual = false

	return id, iniFile, nil
}

func LoadActiveMisterIni() (int, *ini.File, error) {
	return LoadMisterIni(0)
}

func UpdateMisterIni(iniFile *ini.File, key string, value string) error {
	// TODO: support updating specific sections
	if iniFile == nil {
		return fmt.Errorf("iniFile is nil")
	}

	section := iniFile.Section(MisterIniSection)
	if section == nil {
		return fmt.Errorf("mister.ini does not have a [MiSTer] section")
	}

	if !utils.Contains(ValidIniKeys, key) {
		return fmt.Errorf("invalid ini key: %s", key)
	}

	if section.HasKey(key) {
		section.Key(key).SetValue(value)
	} else {
		_, err := section.NewKey(key, value)
		if err != nil {
			return err
		}
	}

	return nil
}

func SaveMisterIni(id int, iniFile *ini.File) error {
	inis, err := ListMisterInis()
	if err != nil {
		return err
	}

	if id < 1 || id > len(inis) {
		return fmt.Errorf("ini id is out of range")
	}

	iniFileInfo := inis[id-1]
	iniPath := iniFileInfo.Path

	backupPath := fmt.Sprintf("%s.backup", iniPath)

	backupData, err := os.ReadFile(iniPath)
	if err != nil {
		return err
	}

	err = os.WriteFile(backupPath, backupData, 0644)
	if err != nil {
		return err
	}

	return iniFile.SaveTo(iniPath)
}

func GetMisterIniOption(file *ini.File, name string) string {
	if file == nil {
		return ""
	}

	section := file.Section(MisterIniSection)
	if section == nil {
		return ""
	}

	key := section.Key(name)
	if key == nil {
		return ""
	}

	return key.String()
}

func RecentsOptionEnabled() bool {
	_, file, err := LoadActiveMisterIni()
	if err != nil {
		return false
	}

	option := GetMisterIniOption(file, "recents")
	if option == "" {
		return false
	}

	return option == "1"
}

type IniFile struct {
	DisplayName string `json:"displayName"`
	Filename    string `json:"filename"`
	Path        string `json:"path"`
}

func ListMisterInis() ([]IniFile, error) {
	var inis []IniFile

	files, err := os.ReadDir(config.SdFolder)
	if err != nil {
		return nil, err
	}

	var iniFilenames []string

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if filepath.Ext(strings.ToLower(file.Name())) == ".ini" {
			iniFilenames = append(iniFilenames, file.Name())
		}
	}

	for _, filename := range iniFilenames {
		lower := strings.ToLower(filename)

		if lower == "mister.ini" {
			inis = append(inis, IniFile{
				DisplayName: "Main",
				Filename:    filename,
				Path:        filepath.Join(config.SdFolder, filename),
			})
		} else if strings.HasPrefix(lower, "mister_") {
			iniFile := IniFile{
				DisplayName: "",
				Filename:    filename,
				Path:        filepath.Join(config.SdFolder, filename),
			}

			iniFile.DisplayName = filename[7:]
			iniFile.DisplayName = strings.TrimSuffix(iniFile.DisplayName, filepath.Ext(iniFile.DisplayName))

			if iniFile.DisplayName == "" {
				iniFile.DisplayName = " -- "
			} else if iniFile.DisplayName == "alt_1" {
				iniFile.DisplayName = "Alt1"
			} else if iniFile.DisplayName == "alt_2" {
				iniFile.DisplayName = "Alt2"
			} else if iniFile.DisplayName == "alt_3" {
				iniFile.DisplayName = "Alt3"
			}

			if len(iniFile.DisplayName) > 4 {
				iniFile.DisplayName = iniFile.DisplayName[0:4]
			}

			if len(inis) < 4 {
				inis = append(inis, iniFile)
			}
		}
	}

	return inis, nil
}
