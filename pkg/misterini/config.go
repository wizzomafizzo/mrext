package misterini

const (
	KeyYpbpr                   = "ypbpr"
	KeyCompositeSync           = "composite_sync"
	KeyForcedScandoubler       = "forced_scandoubler"
	KeyVgaScaler               = "vga_scaler"
	KeyVgaSog                  = "vga_sog"
	KeyKeyrahMode              = "keyrah_mode"
	KeyResetCombo              = "reset_combo"
	KeyKeyMenuAsRgui           = "key_menu_as_rgui"
	KeyVideoMode               = "video_mode"
	KeyVideoModePal            = "video_mode_pal"
	KeyVideoModeNtsc           = "video_mode_ntsc"
	KeyVideoInfo               = "video_info"
	KeyVsyncAdjust             = "vsync_adjust"
	KeyHdmiAudio96k            = "hdmi_audio_96k"
	KeyDviMode                 = "dvi_mode"
	KeyHdmiLimited             = "hdmi_limited"
	KeyKbdNomouse              = "kbd_nomouse"
	KeyMouseThrottle           = "mouse_throttle"
	KeyBootscreen              = "bootscreen"
	KeyVscaleMode              = "vscale_mode"
	KeyVscaleBorder            = "vscale_border"
	KeyRbfHideDatecode         = "rbf_hide_datecode"
	KeyMenuPal                 = "menu_pal"
	KeyBootcore                = "bootcore"
	KeyBootcoreTimeout         = "bootcore_timeout"
	KeyFont                    = "font"
	KeyFbSize                  = "fb_size"
	KeyFbTerminal              = "fb_terminal"
	KeyOsdTimeout              = "osd_timeout"
	KeyDirectVideo             = "direct_video"
	KeyOsdRotate               = "osd_rotate"
	KeyGamepadDefaults         = "gamepad_defaults"
	KeyRecents                 = "recents"
	KeyControllerInfo          = "controller_info"
	KeyRefreshMin              = "refresh_min"
	KeyRefreshMax              = "refresh_max"
	KeyJammaVid                = "jamma_vid"
	KeyJammaPid                = "jamma_pid"
	KeySniperMode              = "sniper_mode"
	KeyBrowseExpand            = "browse_expand"
	KeyLogo                    = "logo"
	KeySharedFolder            = "shared_folder"
	KeyNoMergeVid              = "no_merge_vid"
	KeyNoMergePid              = "no_merge_pid"
	KeyNoMergeVidpid           = "no_merge_vidpid"
	KeyCustomAspectRatio1      = "custom_aspect_ratio_1"
	KeyCustomAspectRatio2      = "custom_aspect_ratio_2"
	KeySpinnerVid              = "spinner_vid"
	KeySpinnerPid              = "spinner_pid"
	KeySpinnerAxis             = "spinner_axis"
	KeySpinnerThrottle         = "spinner_throttle"
	KeyAfilterDefault          = "afilter_default"
	KeyVfilterDefault          = "vfilter_default"
	KeyVfilterVerticalDefault  = "vfilter_vertical_default"
	KeyVfilterScanlinesDefault = "vfilter_scanlines_default"
	KeyShmaskDefault           = "shmask_default"
	KeyShmaskModeDefault       = "shmask_mode_default"
	KeyPresetDefault           = "preset_default"
	KeyLogFileEntry            = "log_file_entry"
	KeyBtAutoDisconnect        = "bt_auto_disconnect"
	KeyBtResetBeforePair       = "bt_reset_before_pair"
	KeyWaitmount               = "waitmount"
	KeyRumble                  = "rumble"
	KeyWheelForce              = "wheel_force"
	KeyWheelRange              = "wheel_range"
	KeyHdmiGameMode            = "hdmi_game_mode"
	KeyVrrMode                 = "vrr_mode"
	KeyVrrMinFramerate         = "vrr_min_framerate"
	KeyVrrMaxFramerate         = "vrr_max_framerate"
	KeyVrrVesaFramerate        = "vrr_vesa_framerate"
	KeyVideoOff                = "video_off"
	KeyPlayer1Controller       = "player_1_controller"
	KeyPlayer2Controller       = "player_2_controller"
	KeyPlayer3Controller       = "player_3_controller"
	KeyPlayer4Controller       = "player_4_controller"
	KeyPlayer5Controller       = "player_5_controller"
	KeyPlayer6Controller       = "player_6_controller"
	KeyDisableAutofire         = "disable_autofire"
	KeyVideoBrightness         = "video_brightness"
	KeyVideoContrast           = "video_contrast"
	KeyVideoSaturation         = "video_saturation"
	KeyVideoHue                = "video_hue"
	KeyVideoGainOffset         = "video_gain_offset"
	KeyHdr                     = "hdr"
	KeyHdrMaxNits              = "hdr_max_nits"
	KeyHdrAvgNits              = "hdr_avg_nits"
	KeyVgaMode                 = "vga_mode"
	KeyNtscMode                = "ntsc_mode"
	KeyControllerUniqueMapping = "controller_unique_mapping"
)

var ValidIniKeys = []string{
	KeyYpbpr,
	KeyCompositeSync,
	KeyForcedScandoubler,
	KeyVgaScaler,
	KeyVgaSog,
	KeyKeyrahMode,
	KeyResetCombo,
	KeyKeyMenuAsRgui,
	KeyVideoMode,
	KeyVideoModePal,
	KeyVideoModeNtsc,
	KeyVideoInfo,
	KeyVsyncAdjust,
	KeyHdmiAudio96k,
	KeyDviMode,
	KeyHdmiLimited,
	KeyKbdNomouse,
	KeyMouseThrottle,
	KeyBootscreen,
	KeyVscaleMode,
	KeyVscaleBorder,
	KeyRbfHideDatecode,
	KeyMenuPal,
	KeyBootcore,
	KeyBootcoreTimeout,
	KeyFont,
	KeyFbSize,
	KeyFbTerminal,
	KeyOsdTimeout,
	KeyDirectVideo,
	KeyOsdRotate,
	KeyGamepadDefaults,
	KeyRecents,
	KeyControllerInfo,
	KeyRefreshMin,
	KeyRefreshMax,
	KeyJammaVid,
	KeyJammaPid,
	KeySniperMode,
	KeyBrowseExpand,
	KeyLogo,
	KeySharedFolder,
	KeyNoMergeVid,
	KeyNoMergePid,
	KeyNoMergeVidpid,
	KeyCustomAspectRatio1,
	KeyCustomAspectRatio2,
	KeySpinnerVid,
	KeySpinnerPid,
	KeySpinnerAxis,
	KeySpinnerThrottle,
	KeyAfilterDefault,
	KeyVfilterDefault,
	KeyVfilterVerticalDefault,
	KeyVfilterScanlinesDefault,
	KeyShmaskDefault,
	KeyShmaskModeDefault,
	KeyPresetDefault,
	KeyLogFileEntry,
	KeyBtAutoDisconnect,
	KeyBtResetBeforePair,
	KeyWaitmount,
	KeyRumble,
	KeyWheelForce,
	KeyWheelRange,
	KeyHdmiGameMode,
	KeyVrrMode,
	KeyVrrMinFramerate,
	KeyVrrMaxFramerate,
	KeyVrrVesaFramerate,
	KeyVideoOff,
	KeyPlayer1Controller,
	KeyPlayer2Controller,
	KeyPlayer3Controller,
	KeyPlayer4Controller,
	KeyPlayer5Controller,
	KeyPlayer6Controller,
	KeyDisableAutofire,
	KeyVideoBrightness,
	KeyVideoContrast,
	KeyVideoSaturation,
	KeyVideoHue,
	KeyVideoGainOffset,
	KeyHdr,
	KeyHdrMaxNits,
	KeyHdrAvgNits,
	KeyVgaMode,
	KeyNtscMode,
	KeyControllerUniqueMapping,
}

// ShadowedIniKeys are keys which can be defined multiple times in an .ini file
var ShadowedIniKeys = []string{
	KeyNoMergeVidpid,
	KeyControllerUniqueMapping,
}

const (
	DefaultIniFilename = "MiSTer.ini"
	MainIniSection     = "MiSTer"
)
