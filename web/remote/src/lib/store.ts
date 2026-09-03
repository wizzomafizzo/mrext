import { create } from "zustand";
import { persist } from "zustand/middleware";
import { SearchServiceStatus } from "./models";

export enum SettingsPageId {
  Main,
  Cores,
  GeneralVideo,
  OSDMenu,
  InputDevices,
  Remote,
  Audio,
  System,
  VideoFilters,
  AnalogVideo,
}

export interface UIState {
  activeTheme: string;
  setActiveTheme: (theme: string) => void;
  fontSize: number;
  setFontSize: (size: number) => void;
  activeSettingsPage: SettingsPageId;
  setActiveSettingsPage: (page: SettingsPageId) => void;
  lastFavoriteFolder: string;
  setLastFavoriteFolder: (folder: string) => void;
  favoriteScripts: string[];
  setFavoriteScripts: (scripts: string[]) => void;
  autoControlKeys: number[];
  setAutoControlKeys: (keys: number[]) => void;
}

export const useUIStateStore = create<UIState>()(
  persist(
    (set, get) => ({
      activeTheme: "mister",
      setActiveTheme: (id: string) => set({ activeTheme: id }),
      fontSize: 14,
      setFontSize: (size: number) => set({ fontSize: size }),
      activeSettingsPage: SettingsPageId.Main,
      setActiveSettingsPage: (page: SettingsPageId) =>
        set({ activeSettingsPage: page }),
      lastFavoriteFolder: "",
      setLastFavoriteFolder: (folder: string) =>
        set({ lastFavoriteFolder: folder }),
      favoriteScripts: [],
      setFavoriteScripts: (scripts: string[]) =>
        set({ favoriteScripts: scripts }),
      autoControlKeys: [
        24, // l2
        25, // l1
        37, // r1
        38, // r2
        103, // up
        108, // down
        105, // left
        106, // right
        23, // osd
        36, // select
        22, // start
        21, // x
        20, // y
        19, // b
        18, // a
        2, // 1
        3, // 2
        4, // 3
        5, // 4
        6, // 5
        7, // 6
        8, // 7
        9, // 8
        10, // 9
        11, // 10
      ],
      setAutoControlKeys: (keys: number[]) => set({ autoControlKeys: keys }),
    }),
    {
      name: "uiState",
    }
  )
);

export interface ServerState {
  search: SearchServiceStatus;
  setSearch: (search: SearchServiceStatus) => void;
  activeGame: string;
  setActiveGame: (game: string) => void;
  activeCore: string;
  setActiveCore: (core: string) => void;
}

export const useServerStateStore = create<ServerState>()((set, get) => ({
  search: {
    ready: true,
    indexing: false,
    totalSteps: 0,
    currentStep: 0,
    currentDesc: "",
  },
  setSearch: (search: SearchServiceStatus) => set({ search }),
  activeGame: "",
  setActiveGame: (game: string) => set({ activeGame: game }),
  activeCore: "",
  setActiveCore: (core: string) => set({ activeCore: core }),
}));

// TODO: KBD_NOMOUSE
// TODO: RUMBLE
// TODO: DVI_MODE is presented as a bool, but defaults to 2, don't know what 0 and 1 actually do
// TODO: add max length to all string inputs
// TODO: preset_default
// TODO: ntsc_mode
// TODO: hdr_max_nits
// TODO: hdr_avg_nits
// TODO: spinner_axis has new option
