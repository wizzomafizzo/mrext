import axios, { AxiosError } from "axios";

import {
  AllWallpapers,
  IndexedSystems,
  MusicServiceStatus,
  Screenshot,
  SearchResults,
  System,
  ViewMenu,
  CreateLauncherRequest,
  KeyboardCodes,
  ListInisPayload,
  SysInfoResponse,
  PeersResponse,
  ScriptsResponse,
  MenuItem,
} from "./models";

const API_ENDPOINT_KEY = "apiEndpoint";
const WS_ENDPOINT_KEY = "wsEndpoint";

export function getStoredApiEndpoint(): string | null {
  return localStorage.getItem(API_ENDPOINT_KEY);
}

export function getApiEndpoint(): string {
  const stored = getStoredApiEndpoint();

  if (stored) {
    return stored;
  } else {
    return window.location.protocol + "//" + window.location.host + "/api";
  }
}

export function setApiEndpoint(endpoint: string): void {
  localStorage.setItem(API_ENDPOINT_KEY, endpoint);
}

export function getStoredWsEndpoint(): string | null {
  return localStorage.getItem(WS_ENDPOINT_KEY);
}

export function getWsEndpoint(): string {
  const stored = getStoredWsEndpoint();
  const apiEndpoint = getStoredApiEndpoint();

  if (stored) {
    return stored;
  } else if (apiEndpoint) {
    return apiEndpoint.replace(/^http/, "ws") + "/ws";
  } else {
    const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
    return protocol + "//" + window.location.host + "/api/ws";
  }
}

export function setWsEndpoint(endpoint: string): void {
  localStorage.setItem(WS_ENDPOINT_KEY, endpoint);
}

export class ControlApi {
  apiUrl: string;

  constructor() {
    let endpoint = getApiEndpoint();
    axios.defaults.baseURL = endpoint;
    this.apiUrl = endpoint;
  }

  // screenshots

  async getScreenshots(): Promise<Screenshot[]> {
    return (await axios.get<Screenshot[]>(`/screenshots`)).data;
  }

  getScreenshotUrl(path: string): string {
    return `${this.apiUrl}/screenshots/${path}`;
  }

  async takeScreenshot(): Promise<Screenshot> {
    return (await axios.post<Screenshot>(`/screenshots`)).data;
  }

  async deleteScreenshot(path: string): Promise<void> {
    await axios.delete(`/screenshots/${path}`);
  }

  // systems

  async getSystems(): Promise<System[]> {
    return (await axios.get<System[]>(`/systems`)).data;
  }

  async launchSystem(id: string): Promise<void> {
    await axios.post(`/systems/${id}`);
  }

  // wallpaper

  async getWallpapers(): Promise<AllWallpapers> {
    return (await axios.get<AllWallpapers>(`/wallpapers`)).data;
  }

  getWallpaperUrl(filename: string): string {
    return `${this.apiUrl}/wallpapers/${filename}`;
  }

  async setWallpaper(filename: string): Promise<void> {
    await axios.post(`/wallpapers/${filename}`);
  }

  async unsetWallpaper(): Promise<void> {
    await axios.delete(`/wallpapers`);
  }

  async deleteWallpaper(filename: string): Promise<void> {
    await axios.delete(`/wallpapers/${filename}`);
  }

  // music

  async musicStatus(): Promise<MusicServiceStatus> {
    return (await axios.get<MusicServiceStatus>(`/music/status`)).data;
  }

  async playMusic(): Promise<void> {
    await axios.post(`/music/play`);
  }

  async stopMusic(): Promise<void> {
    await axios.post(`/music/stop`);
  }

  async nextMusic(): Promise<void> {
    await axios.post(`/music/next`);
  }

  async setMusicPlayback(playback: string): Promise<void> {
    await axios.post(`/music/playback/${playback}`);
  }

  async setMusicPlaylist(playlist: string): Promise<void> {
    await axios.post(`/music/playlist/${playlist}`);
  }

  async getMusicPlaylists(): Promise<string[]> {
    return (await axios.get<string[]>(`/music/playlist`)).data;
  }

  // games

  async searchGames(data: {
    query: string;
    system: string;
  }): Promise<SearchResults> {
    return (await axios.post<SearchResults>(`/games/search`, data)).data;
  }

  async launchGame(path: string): Promise<void> {
    await axios.post(`/games/launch`, { path });
  }

  async startSearchIndex(): Promise<void> {
    await axios.post(`/games/index`);
  }

  async indexedSystems(): Promise<IndexedSystems> {
    return (await axios.get<IndexedSystems>(`/games/search/systems`)).data;
  }

  // control

  async sendKeyboard(key: KeyboardCodes) {
    await axios.post(`/controls/keyboard/${key}`);
  }

  async sendRawKeyboard(key: number) {
    await axios.post(`/controls/keyboard-raw/${key}`);
  }

  // menu

  async listMenuFolder(path: string): Promise<ViewMenu> {
    return (
      await axios.post<ViewMenu>(`/menu/view`, {
        path,
      })
    ).data;
  }

  // launch
  async launchFile(path: string) {
    await axios.post(`/launch`, { path });
  }

  async launchMenu(): Promise<void> {
    return (await axios.post(`/launch/menu`)).data;
  }

  async createLauncher(data: CreateLauncherRequest): Promise<{ path: string }> {
    return (await axios.post(`/launch/new`, data)).data;
  }

  // settings
  async saveMisterIni(
    id: number,
    data: { [key: string]: string },
  ): Promise<void> {
    await axios.put(`/settings/inis/${id}`, data);
  }

  async listMisterInis(): Promise<ListInisPayload> {
    return (await axios.get<ListInisPayload>(`/settings/inis`)).data;
  }

  async setMisterIni(data: { ini: number }): Promise<void> {
    await axios.put(`/settings/inis`, data);
  }

  async setMenuBackgroundMode(data: { mode: number }): Promise<void> {
    await axios.put(`/settings/cores/menu`, data);
  }

  async restartRemoteService(): Promise<void> {
    await axios.post(`/settings/remote/restart`);
  }

  async loadMisterIni(id: number): Promise<{ [key: string]: string }> {
    return (await axios.get<{ [key: string]: string }>(`/settings/inis/${id}`))
      .data;
  }

  async sysInfo(): Promise<SysInfoResponse> {
    return (await axios.get<SysInfoResponse>(`/sysinfo`)).data;
  }

  async getPeers(): Promise<PeersResponse> {
    return (await axios.get<PeersResponse>(`/settings/remote/peers`)).data;
  }

  async reboot(): Promise<void> {
    await axios.post(`/settings/system/reboot`);
  }

  async generateMac(): Promise<{ mac: string }> {
    return (await axios.get<{ mac: string }>(`/settings/system/generate-mac`))
      .data;
  }

  async createMenuFile(data: { type: string; folder: string; name: string }) {
    await axios.post(`/menu/files/create`, data);
  }

  async renameMenuFile(data: { fromPath: string; toPath: string }) {
    await axios.post(`/menu/files/rename`, data);
  }

  async deleteMenuFile(data: { path: string }) {
    await axios.post(`/menu/files/delete`, data);
  }

  async getAllScripts(): Promise<ScriptsResponse> {
    return (await axios.get<ScriptsResponse>(`/scripts/list`)).data;
  }

  async runScript(filename: string) {
    await axios.post(`/scripts/launch/${filename}`);
  }

  async openConsole() {
    await axios.post(`/scripts/console`);
  }

  async closeConsole() {
    return this.sendKeyboard("exit_console");
  }

  async killScript() {
    await axios.post(`/scripts/kill`);
  }

  async nfcStatus(): Promise<{
    available: boolean;
    running: boolean;
  }> {
    return (await axios.get(`/nfc/status`)).data;
  }

  async nfcWrite(data: { path: string }) {
    await axios.post(`/nfc/write`, data);
  }

  async nfcCancel() {
    await axios.post(`/nfc/cancel`);
  }

  async listGamesFolder(path: string): Promise<ViewMenu> {
    return (
      await axios.post<ViewMenu>(`/games/view`, {
        path,
      })
    ).data;
  }
}
