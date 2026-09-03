import { ControlApi } from "./api";
import { create } from "zustand";

const iniKeyMap: { [key: string]: string } = {
  ypbpr: "ypbpr",
  compositeSync: "composite_sync",
  forcedScandoubler: "forced_scandoubler",
  vgaScaler: "vga_scaler",
  vgaSog: "vga_sog",
  keyrahMode: "keyrah_mode",
  resetCombo: "reset_combo",
  keyMenuAsRgui: "key_menu_as_rgui",
  videoMode: "video_mode",
  videoModePal: "video_mode_pal",
  videoModeNtsc: "video_mode_ntsc",
  videoInfo: "video_info",
  vsyncAdjust: "vsync_adjust",
  hdmiAudio96k: "hdmi_audio_96k",
  dviMode: "dvi_mode",
  hdmiLimited: "hdmi_limited",
  keyboardNoMouse: "kbd_nomouse",
  mouseThrottle: "mouse_throttle",
  bootScreen: "bootscreen",
  verticalScaleMode: "vscale_mode",
  vscaleBorder: "vscale_border",
  rbfHideDatecode: "rbf_hide_datecode",
  menuPal: "menu_pal",
  bootCore: "bootcore",
  bootCoreTimeout: "bootcore_timeout",
  font: "font",
  fbSize: "fb_size",
  fbTerminal: "fb_terminal",
  osdTimeout: "osd_timeout",
  directVideo: "direct_video",
  osdRotate: "osd_rotate",
  gamepadDefaults: "gamepad_defaults",
  recents: "recents",
  controllerInfo: "controller_info",
  refreshMin: "refresh_min",
  refreshMax: "refresh_max",
  jammaVid: "jamma_vid",
  jammaPid: "jamma_pid",
  sniperMode: "sniper_mode",
  browseExpand: "browse_expand",
  logo: "logo",
  sharedFolder: "shared_folder",
  noMergeVid: "no_merge_vid",
  noMergePid: "no_merge_pid",
  noMergeVidPid: "no_merge_vidpid",
  customAspectRatio1: "custom_aspect_ratio_1",
  customAspectRatio2: "custom_aspect_ratio_2",
  spinnerVid: "spinner_vid",
  spinnerPid: "spinner_pid",
  spinnerAxis: "spinner_axis",
  spinnerThrottle: "spinner_throttle",
  aFilterDefault: "afilter_default",
  vfilterDefault: "vfilter_default",
  vfilterVerticalDefault: "vfilter_vertical_default",
  vfilterScanlinesDefault: "vfilter_scanlines_default",
  shmaskDefault: "shmask_default",
  shmaskModeDefault: "shmask_mode_default",
  presetDefault: "preset_default",
  logFileEntry: "log_file_entry",
  btAutoDisconnect: "bt_auto_disconnect",
  btResetBeforePair: "bt_reset_before_pair",
  waitMount: "waitmount",
  rumble: "rumble",
  wheelForce: "wheel_force",
  wheelRange: "wheel_range",
  hdmiGameMode: "hdmi_game_mode",
  vrrMode: "vrr_mode",
  vrrMinFramerate: "vrr_min_framerate",
  vrrMaxFramerate: "vrr_max_framerate",
  vrrVesaFramerate: "vrr_vesa_framerate",
  videoOff: "video_off",
  player1Controller: "player_1_controller",
  player2Controller: "player_2_controller",
  player3Controller: "player_3_controller",
  player4Controller: "player_4_controller",
  player5Controller: "player_5_controller",
  player6Controller: "player_6_controller",
  disableAutoFire: "disable_autofire",
  videoBrightness: "video_brightness",
  videoContrast: "video_contrast",
  videoSaturation: "video_saturation",
  videoHue: "video_hue",
  videoGainOffset: "video_gain_offset",
  hdr: "hdr",
  hdrMaxNits: "hdr_max_nits",
  hdrAvgNits: "hdr_avg_nits",
  vgaMode: "vga_mode",
  ntscMode: "ntsc_mode",
  // : "controller_unique_mapping",
  hostname: "__hostname",
  ethernetMacAddress: "__ethernetMacAddress",
};

interface IniActions {
  setAttribute: (key: string, value: string) => void;
  reset: () => void;
  resetModified: () => void;
  setOriginal: (ini: IniState) => void;
  revertChanges: () => void;

  // video
  setVideoMode: (v: string) => void;
  setVideoModeNtsc: (v: string) => void;
  setVideoModePal: (v: string) => void;
  setVerticalScaleMode: (v: string) => void;
  setVsyncAdjust: (v: string) => void;
  setRefreshMin: (v: string) => void;
  setRefreshMax: (v: string) => void;
  setDviMode: (v: string) => void;
  setVscaleBorder: (v: string) => void;
  setVfilterDefault: (v: string) => void;
  setVfilterVerticalDefault: (v: string) => void;
  setVfilterScanlinesDefault: (v: string) => void;
  setShmaskDefault: (v: string) => void;
  setShmaskModeDefault: (v: string) => void;
  setHdmiGameMode: (v: string) => void;
  setHdmiLimited: (v: string) => void;
  setVrrMode: (v: string) => void;
  setVrrMinFramerate: (v: string) => void;
  setVrrMaxFramerate: (v: string) => void;
  setVrrVesaFramerate: (v: string) => void;
  setPresetDefault: (v: string) => void;
  setDirectVideo: (v: string) => void;
  setForcedScandoubler: (v: string) => void;
  setYpbpr: (v: string) => void;
  setCompositeSync: (v: string) => void;
  setVgaScaler: (v: string) => void;
  setVgaSog: (v: string) => void;
  setCustomAspectRatio1: (v: string) => void;
  setCustomAspectRatio2: (v: string) => void;
  setHdr: (v: string) => void;
  setVideoBrightness: (v: string) => void;
  setVideoContrast: (v: string) => void;
  setVideoSaturation: (v: string) => void;
  setVideoHue: (v: string) => void;
  setVideoGainOffset: (v: string) => void;
  setHdrMaxNits: (v: string) => void;
  setHdrAvgNits: (v: string) => void;
  setVgaMode: (v: string) => void;
  setNtscMode: (v: string) => void;

  // cores
  setBootScreen: (v: string) => void;
  setRecents: (v: string) => void;
  setVideoInfo: (v: string) => void;
  setControllerInfo: (v: string) => void;
  setSharedFolder: (v: string) => void;
  setLogFileEntry: (v: string) => void;
  setKeyMenuAsRgui: (v: string) => void;

  // input devices
  setBtAutoDisconnect: (v: string) => void;
  setBtResetBeforePair: (v: string) => void;
  setMouseThrottle: (v: string) => void;
  setResetCombo: (v: string) => void;
  setPlayer1Controller: (v: string) => void;
  setPlayer2Controller: (v: string) => void;
  setPlayer3Controller: (v: string) => void;
  setPlayer4Controller: (v: string) => void;
  setPlayer5Controller: (v: string) => void;
  setPlayer6Controller: (v: string) => void;
  setSniperMode: (v: string) => void;
  setSpinnerThrottle: (v: string) => void;
  setSpinnerAxis: (v: string) => void;
  setGamepadDefaults: (v: string) => void;
  setDisableAutoFire: (v: string) => void;
  setWheelForce: (v: string) => void;
  setWheelRange: (v: string) => void;
  setSpinnerVid: (v: string) => void;
  setSpinnerPid: (v: string) => void;
  setKeyrahMode: (v: string) => void;
  setJammaVid: (v: string) => void;
  setJammaPid: (v: string) => void;
  setNoMergeVid: (v: string) => void;
  setNoMergePid: (v: string) => void;
  setNoMergeVidPid: (v: string) => void;
  setRumble: (v: string) => void;
  setKeyboardNoMouse: (v: string) => void;

  // audio
  setHdmiAudio96k: (v: string) => void;
  setAFilterDefault: (v: string) => void;

  // system
  setFbSize: (v: string) => void;
  setFbTerminal: (v: string) => void;
  setBootCore: (v: string) => void;
  setBootCoreTimeout: (v: string) => void;
  setWaitMount: (v: string) => void;
  setHostname: (v: string) => void;
  setEthernetMacAddress: (v: string) => void;

  // osd/menu
  setRbfHideDatecode: (v: string) => void;
  setOsdRotate: (v: string) => void;
  setBrowseExpand: (v: string) => void;
  setOsdTimeout: (v: string) => void;
  setVideoOff: (v: string) => void;
  setMenuPal: (v: string) => void;
  setFont: (v: string) => void;
  setLogo: (v: string) => void;
}

interface IniState {
  // video
  videoMode: string;
  videoModeNtsc: string;
  videoModePal: string;
  verticalScaleMode: string;
  vsyncAdjust: string;
  refreshMin: string;
  refreshMax: string;
  dviMode: string;
  vscaleBorder: string;
  vfilterDefault: string;
  vfilterVerticalDefault: string;
  vfilterScanlinesDefault: string;
  shmaskDefault: string;
  shmaskModeDefault: string;
  hdmiGameMode: string;
  hdmiLimited: string;
  vrrMode: string;
  vrrMinFramerate: string;
  vrrMaxFramerate: string;
  vrrVesaFramerate: string;
  presetDefault: string;
  directVideo: string;
  forcedScandoubler: string;
  ypbpr: string;
  compositeSync: string;
  vgaScaler: string;
  vgaSog: string;
  customAspectRatio1: string;
  customAspectRatio2: string;
  hdr: string;
  videoBrightness: string;
  videoContrast: string;
  videoSaturation: string;
  videoHue: string;
  videoGainOffset: string;
  hdrMaxNits: string;
  hdrAvgNits: string;
  vgaMode: string;
  ntscMode: string;

  // cores
  bootScreen: string;
  recents: string;
  videoInfo: string;
  controllerInfo: string;
  sharedFolder: string;
  logFileEntry: string;
  keyMenuAsRgui: string;

  // input devices
  btAutoDisconnect: string;
  btResetBeforePair: string;
  mouseThrottle: string;
  resetCombo: string;
  player1Controller: string;
  player2Controller: string;
  player3Controller: string;
  player4Controller: string;
  player5Controller: string;
  player6Controller: string;
  sniperMode: string;
  spinnerThrottle: string;
  spinnerAxis: string;
  gamepadDefaults: string;
  disableAutoFire: string;
  wheelForce: string;
  wheelRange: string;
  spinnerVid: string;
  spinnerPid: string;
  keyrahMode: string;
  jammaVid: string;
  jammaPid: string;
  noMergeVid: string;
  noMergePid: string;
  noMergeVidPid: string;
  rumble: string;
  keyboardNoMouse: string;

  // audio
  hdmiAudio96k: string;
  aFilterDefault: string;

  // system
  fbSize: string;
  fbTerminal: string;
  bootCore: string;
  bootCoreTimeout: string;
  waitMount: string;
  hostname: string;
  ethernetMacAddress: string;

  // osd/menu
  rbfHideDatecode: string;
  osdRotate: string;
  browseExpand: string;
  osdTimeout: string;
  videoOff: string;
  menuPal: string;
  font: string;
  logo: string;
}

type IniStore = IniState &
  IniActions & {
    original: IniState;
    modified: string[];
  };

const initialState: IniState = {
  // video
  videoMode: "",
  videoModeNtsc: "",
  videoModePal: "",
  verticalScaleMode: "0",
  vsyncAdjust: "0",
  refreshMin: "0",
  refreshMax: "0",
  dviMode: "2",
  vscaleBorder: "0",
  vfilterDefault: "",
  vfilterVerticalDefault: "",
  vfilterScanlinesDefault: "",
  shmaskDefault: "",
  shmaskModeDefault: "0",
  hdmiGameMode: "0",
  hdmiLimited: "0",
  vrrMode: "0",
  vrrMinFramerate: "0",
  vrrMaxFramerate: "0",
  vrrVesaFramerate: "0",
  directVideo: "0",
  forcedScandoubler: "0",
  ypbpr: "0",
  compositeSync: "0",
  vgaScaler: "0",
  vgaSog: "0",
  customAspectRatio1: "",
  customAspectRatio2: "",
  hdr: "0",
  videoBrightness: "50",
  videoContrast: "50",
  videoSaturation: "100",
  videoHue: "0",
  videoGainOffset: "1, 0, 1, 0, 1, 0",
  presetDefault: "",
  hdrMaxNits: "1000",
  hdrAvgNits: "250",
  vgaMode: "",
  ntscMode: "0",

  // cores
  bootScreen: "1",
  recents: "0",
  videoInfo: "0",
  controllerInfo: "6",
  sharedFolder: "",
  logFileEntry: "0",
  keyMenuAsRgui: "0",

  // input devices
  btAutoDisconnect: "0",
  btResetBeforePair: "0",
  mouseThrottle: "0",
  resetCombo: "0",
  player1Controller: "",
  player2Controller: "",
  player3Controller: "",
  player4Controller: "",
  player5Controller: "",
  player6Controller: "",
  sniperMode: "0",
  spinnerThrottle: "",
  spinnerAxis: "0",
  gamepadDefaults: "0",
  disableAutoFire: "0",
  wheelForce: "50",
  wheelRange: "",
  spinnerVid: "",
  spinnerPid: "",
  keyrahMode: "",
  jammaVid: "",
  jammaPid: "",
  noMergeVid: "",
  noMergePid: "",
  noMergeVidPid: "",
  rumble: "1",
  keyboardNoMouse: "0",

  // audio
  hdmiAudio96k: "0",
  aFilterDefault: "",

  // system
  fbSize: "0",
  fbTerminal: "1",
  bootCore: "",
  bootCoreTimeout: "0",
  waitMount: "",
  hostname: "MiSTer",
  ethernetMacAddress: "",

  // osd/menu
  rbfHideDatecode: "0",
  osdRotate: "0",
  browseExpand: "1",
  osdTimeout: "0",
  videoOff: "0",
  menuPal: "0",
  font: "",
  logo: "1",
};

const distinct = (arr: string[]) => [...new Set(arr)];

function m(
  state: IniStore,
  key: string,
  value: string
): IniStore | Partial<IniStore> {
  console.log("update", key, "=", value);

  if ((state as Indexable)[key] === value) {
    console.log("no change");
    return {};
  }

  if ((state.original as Indexable)[key] === value) {
    console.log("reverted");
    return {
      [key]: value,
      modified: state.modified.filter((m) => m !== key),
    };
  }

  const partial = {
    [key]: value,
    modified: distinct([...state.modified, key]),
  };

  console.log("partial", partial);

  return partial;
}

export const useIniSettingsStore = create<IniStore>()((set) => ({
  ...initialState,

  original: initialState,
  modified: [],

  setAttribute: (key: string, value: string) =>
    set(() => ({
      [key]: value,
    })),

  reset: () => set(initialState),
  resetModified: () => set({ modified: [] }),

  setOriginal: (ini: IniState) => set({ original: ini }),
  revertChanges: () => set((s) => ({ ...s, ...s.original })),

  setVideoMode: (v: string) => set((s) => m(s, "videoMode", v)),
  setVideoModeNtsc: (v: string) => set((s) => m(s, "videoModeNtsc", v)),
  setVideoModePal: (v: string) => set((s) => m(s, "videoModePal", v)),
  setVerticalScaleMode: (v: string) => set((s) => m(s, "verticalScaleMode", v)),
  setVsyncAdjust: (v: string) => set((s) => m(s, "vsyncAdjust", v)),
  setRefreshMin: (v: string) => set((s) => m(s, "refreshMin", v)),
  setRefreshMax: (v: string) => set((s) => m(s, "refreshMax", v)),
  setDviMode: (v: string) => set((s) => m(s, "dviMode", v)),
  setVscaleBorder: (v: string) => set((s) => m(s, "vscaleBorder", v)),
  setVfilterDefault: (v: string) => set((s) => m(s, "vfilterDefault", v)),
  setVfilterVerticalDefault: (v: string) =>
    set((s) => m(s, "vfilterVerticalDefault", v)),
  setVfilterScanlinesDefault: (v: string) =>
    set((s) => m(s, "vfilterScanlinesDefault", v)),
  setShmaskDefault: (v: string) => set((s) => m(s, "shmaskDefault", v)),
  setShmaskModeDefault: (v: string) => set((s) => m(s, "shmaskModeDefault", v)),
  setHdmiGameMode: (v: string) => set((s) => m(s, "hdmiGameMode", v)),
  setHdmiLimited: (v: string) => set((s) => m(s, "hdmiLimited", v)),
  setVrrMode: (v: string) => set((s) => m(s, "vrrMode", v)),
  setVrrMinFramerate: (v: string) => set((s) => m(s, "vrrMinFramerate", v)),
  setVrrMaxFramerate: (v: string) => set((s) => m(s, "vrrMaxFramerate", v)),
  setVrrVesaFramerate: (v: string) => set((s) => m(s, "vrrVesaFramerate", v)),
  setDirectVideo: (v: string) => set((s) => m(s, "directVideo", v)),
  setForcedScandoubler: (v: string) => set((s) => m(s, "forcedScandoubler", v)),
  setYpbpr: (v: string) => set((s) => m(s, "ypbpr", v)),
  setCompositeSync: (v: string) => set((s) => m(s, "compositeSync", v)),
  setVgaScaler: (v: string) => set((s) => m(s, "vgaScaler", v)),
  setVgaSog: (v: string) => set((s) => m(s, "vgaSog", v)),
  setCustomAspectRatio1: (v: string) =>
    set((s) => m(s, "customAspectRatio1", v)),
  setCustomAspectRatio2: (v: string) =>
    set((s) => m(s, "customAspectRatio2", v)),
  setHdr: (v: string) => set((s) => m(s, "hdr", v)),
  setVideoBrightness: (v: string) => set((s) => m(s, "videoBrightness", v)),
  setVideoContrast: (v: string) => set((s) => m(s, "videoContrast", v)),
  setVideoSaturation: (v: string) => set((s) => m(s, "videoSaturation", v)),
  setVideoHue: (v: string) => set((s) => m(s, "videoHue", v)),
  setVideoGainOffset: (v: string) => set((s) => m(s, "videoGainOffset", v)),
  setPresetDefault: (v: string) => set((s) => m(s, "presetDefault", v)),
  setHdrMaxNits: (v: string) => set((s) => m(s, "hdrMaxNits", v)),
  setHdrAvgNits: (v: string) => set((s) => m(s, "hdrAvgNits", v)),
  setVgaMode: (v: string) => set((s) => m(s, "vgaMode", v)),
  setNtscMode: (v: string) => set((s) => m(s, "ntscMode", v)),

  // cores
  setBootScreen: (v: string) => set((s) => m(s, "bootScreen", v)),
  setRecents: (v: string) => set((s) => m(s, "recents", v)),
  setVideoInfo: (v: string) => set((s) => m(s, "videoInfo", v)),
  setControllerInfo: (v: string) => set((s) => m(s, "controllerInfo", v)),
  setSharedFolder: (v: string) => set((s) => m(s, "sharedFolder", v)),
  setLogFileEntry: (v: string) => set((s) => m(s, "logFileEntry", v)),
  setKeyMenuAsRgui: (v: string) => set((s) => m(s, "keyMenuAsRgui", v)),

  // input devices
  setBtAutoDisconnect: (v: string) => set((s) => m(s, "btAutoDisconnect", v)),
  setBtResetBeforePair: (v: string) => set((s) => m(s, "btResetBeforePair", v)),
  setMouseThrottle: (v: string) => set((s) => m(s, "mouseThrottle", v)),
  setResetCombo: (v: string) => set((s) => m(s, "resetCombo", v)),
  setPlayer1Controller: (v: string) => set((s) => m(s, "player1Controller", v)),
  setPlayer2Controller: (v: string) => set((s) => m(s, "player2Controller", v)),
  setPlayer3Controller: (v: string) => set((s) => m(s, "player3Controller", v)),
  setPlayer4Controller: (v: string) => set((s) => m(s, "player4Controller", v)),
  setPlayer5Controller: (v: string) => set((s) => m(s, "player5Controller", v)),
  setPlayer6Controller: (v: string) => set((s) => m(s, "player6Controller", v)),
  setSniperMode: (v: string) => set((s) => m(s, "sniperMode", v)),
  setSpinnerThrottle: (v: string) => set((s) => m(s, "spinnerThrottle", v)),
  setSpinnerAxis: (v: string) => set((s) => m(s, "spinnerAxis", v)),
  setGamepadDefaults: (v: string) => set((s) => m(s, "gamepadDefaults", v)),
  setDisableAutoFire: (v: string) => set((s) => m(s, "disableAutoFire", v)),
  setWheelForce: (v: string) => set((s) => m(s, "wheelForce", v)),
  setWheelRange: (v: string) => set((s) => m(s, "wheelRange", v)),
  setSpinnerVid: (v: string) => set((s) => m(s, "spinnerVid", v)),
  setSpinnerPid: (v: string) => set((s) => m(s, "spinnerPid", v)),
  setKeyrahMode: (v: string) => set((s) => m(s, "keyrahMode", v)),
  setJammaVid: (v: string) => set((s) => m(s, "jammaVid", v)),
  setJammaPid: (v: string) => set((s) => m(s, "jammaPid", v)),
  setNoMergeVid: (v: string) => set((s) => m(s, "noMergeVid", v)),
  setNoMergePid: (v: string) => set((s) => m(s, "noMergePid", v)),
  setNoMergeVidPid: (v: string) => set((s) => m(s, "noMergeVidPid", v)),
  setRumble: (v: string) => set((s) => m(s, "rumble", v)),
  setKeyboardNoMouse: (v: string) => set((s) => m(s, "keyboardNoMouse", v)),

  // audio
  setHdmiAudio96k: (v: string) => set((s) => m(s, "hdmiAudio96k", v)),
  setAFilterDefault: (v: string) => set((s) => m(s, "aFilterDefault", v)),

  // system
  setFbSize: (v: string) => set((s) => m(s, "fbSize", v)),
  setFbTerminal: (v: string) => set((s) => m(s, "fbTerminal", v)),
  setBootCore: (v: string) => set((s) => m(s, "bootCore", v)),
  setBootCoreTimeout: (v: string) => set((s) => m(s, "bootCoreTimeout", v)),
  setWaitMount: (v: string) => set((s) => m(s, "waitMount", v)),
  setHostname: (v: string) => set((s) => m(s, "hostname", v)),
  setEthernetMacAddress: (v: string) =>
    set((s) => m(s, "ethernetMacAddress", v)),

  // osd/menu
  setRbfHideDatecode: (v: string) => set((s) => m(s, "rbfHideDatecode", v)),
  setOsdRotate: (v: string) => set((s) => m(s, "osdRotate", v)),
  setBrowseExpand: (v: string) => set((s) => m(s, "browseExpand", v)),
  setOsdTimeout: (v: string) => set((s) => m(s, "osdTimeout", v)),
  setVideoOff: (v: string) => set((s) => m(s, "videoOff", v)),
  setMenuPal: (v: string) => set((s) => m(s, "menuPal", v)),
  setFont: (v: string) => set((s) => m(s, "font", v)),
  setLogo: (v: string) => set((s) => m(s, "logo", v)),
}));

const flip = (data: { [key: string]: string }) =>
  Object.fromEntries(Object.entries(data).map(([key, value]) => [value, key]));

const iniKeyMapReverse = flip(iniKeyMap);

interface Indexable {
  [key: string]: any;
}

function newIniRequest(state: IniStore): {
  [key: string]: string;
} {
  const changes: { [key: string]: string } = {};

  for (const name of state.modified) {
    if (!(name in iniKeyMap)) {
      console.error(`Unknown ini key: ${name}`);
      continue;
    }
    const key = iniKeyMap[name];
    changes[key] = (state as Indexable)[name];
  }

  return changes;
}

export function saveMisterIni(id: number, state: IniStore) {
  const changes = newIniRequest(state);
  console.log(changes);
  const api = new ControlApi();
  return api.saveMisterIni(id, changes).then(() => {
    state.resetModified();
    const newState = { ...state.original };
    for (const key in changes) {
      if (key in iniKeyMapReverse) {
        const mKey = iniKeyMapReverse[key];
        (newState as Indexable)[mKey] = changes[key];
      } else {
        console.warn(`Unknown ini key ${key}`);
      }
    }
    state.setOriginal(newState);
  });
}

export function loadMisterIni(id: number, state: IniStore, reset = false) {
  const api = new ControlApi();
  return api
    .loadMisterIni(id)
    .then((data) => {
      console.log("Loading MiSTer.ini data...");
      const newState = { ...initialState };

      for (const key in data) {
        if (key in iniKeyMapReverse) {
          const mKey = iniKeyMapReverse[key];
          (newState as Indexable)[mKey] = data[key];
        } else {
          console.warn(`Unknown ini key ${key}`);
        }
      }

      state.setOriginal(newState);

      if (reset) {
        state.resetModified();
      }

      for (const key in data) {
        if (key in iniKeyMapReverse) {
          const mKey = iniKeyMapReverse[key];

          if (state.modified.includes(key)) {
            console.log(`Skipping ini key ${mKey} as ${key} is modified`);
            continue;
          }

          console.log(`Setting ini key ${mKey} to ${data[key]}`);
          state.setAttribute(mKey, data[key]);
        } else {
          console.warn(`Unknown ini key ${key}`);
        }
      }
    })
    .catch((e) => {
      console.error(e);
    });
}
