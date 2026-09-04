export interface System {
  id: string;
  name: string;
  category: string;
}

export interface Game {
  name: string;
  system: System;
  path: string;
}

export interface IndexedSystems {
  systems: {
    id: string;
    name: string;
  }[];
}

export interface SearchResults {
  data: Game[];
  total: number;
  pageSize: number;
  page: number;
}

export interface Screenshot {
  filename: string;
  path: string;
  core: string;
  modified: string;
  game: string;
}

export interface Wallpaper {
  filename: string;
  name: string;
  width: number;
  height: number;
  active: boolean;
}

export interface AllWallpapers {
  active: string;
  backgroundMode: number;
  wallpapers: Wallpaper[];
}

export interface MusicServiceStatus {
  running: boolean;
  playing: boolean;
  playback: string;
  playlist: string;
  track: string;
}

export interface SearchServiceStatus {
  ready: boolean;
  indexing: boolean;
  totalSteps: number;
  currentStep: number;
  currentDesc: string;
}

export interface MenuItem {
  name: string;
  namesTxt?: string;
  path: string;
  parent: string;
  filename: string;
  extension: string;
  next?: string;
  type: "folder" | "rbf" | "mra" | "mgl" | "zip";
  modified: string;
  version?: string;
  size: number;
  inZip: boolean;
  system?: System;
}

export interface ViewMenu {
  up?: string;
  items: MenuItem[];
}

export type KeyboardCodes =
  | "up"
  | "down"
  | "left"
  | "right"
  | "volume_up"
  | "volume_down"
  | "volume_mute"
  | "menu"
  | "back"
  | "confirm"
  | "cancel"
  | "osd"
  | "screenshot"
  | "raw_screenshot"
  | "pair_bluetooth"
  | "change_background"
  | "core_select"
  | "user"
  | "reset"
  | "toggle_core_dates"
  | "console"
  | "exit_console"
  | "computer_osd";

export type MouseButtons =
  | "click"
  | "double_click"
  | "right"
  | "middle"
  | "left_down"
  | "left_up"
  | "right_down"
  | "right_up"
  | "middle_down"
  | "middle_up";

export interface CreateLauncherRequest {
  gamePath: string;
  folder: string;
  name: string;
}

interface IniResponse {
  displayName: string;
  filename: string;
  path: string;
}

export interface ListInisPayload {
  active: number;
  inis: IniResponse[];
}

export interface SysInfoResponse {
  ips: string[];
  hostname: string;
  dns: string;
  version: string;
  updated: string;
  disks: {
    path: string;
    total: number;
    used: number;
    free: number;
    displayName: string;
  }[];
}

export interface PeersResponse {
  peers: {
    hostname: string;
    version: string;
    ip: string;
  }[];
}

export interface Script {
  name: string;
  path: string;
  filename: string;
}

export interface ScriptsResponse {
  canLaunch: boolean;
  scripts: Script[];
}
