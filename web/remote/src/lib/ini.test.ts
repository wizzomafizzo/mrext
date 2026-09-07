// @vitest-environment jsdom

import { beforeEach, describe, expect, it, vi } from "vitest";

import { ControlApi } from "./api";
import { loadMisterIni, saveMisterIni, useIniSettingsStore } from "./ini";

describe("INI settings state", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    useIniSettingsStore.getState().reset();
    vi.spyOn(ControlApi.prototype, "listMisterInis").mockResolvedValue({
      active: 1,
      inis: [
        {
          id: 1,
          filename: "MiSTer.ini",
          displayName: "Main",
          path: "/media/fat/MiSTer.ini",
        },
        {
          id: 2,
          filename: "MiSTer_auto.ini",
          displayName: "auto",
          path: "/media/fat/MiSTer_auto.ini",
        },
        {
          id: 4,
          filename: "MiSTer_ultrawide.ini",
          displayName: "ultrawide",
          path: "/media/fat/MiSTer_ultrawide.ini",
        },
      ],
    });
  });

  it("pins saves to the loaded file even when active slot changes", async () => {
    vi.spyOn(ControlApi.prototype, "loadMisterIni").mockResolvedValue({});
    const save = vi
      .spyOn(ControlApi.prototype, "saveMisterIni")
      .mockResolvedValue();
    await loadMisterIni(4, useIniSettingsStore.getState());
    useIniSettingsStore.getState().setHostname("wide");
    vi.mocked(ControlApi.prototype.listMisterInis).mockResolvedValue({
      active: 2,
      inis: [],
    });
    await saveMisterIni(4, useIniSettingsStore.getState());
    expect(save).toHaveBeenCalledWith(
      4,
      { __hostname: "wide" },
      "MiSTer_ultrawide.ini",
    );
    await expect(
      saveMisterIni(2, useIniSettingsStore.getState()),
    ).rejects.toThrow("Load a valid INI");
  });

  it("rejects unavailable slot gaps rather than editing the next file", async () => {
    const load = vi.spyOn(ControlApi.prototype, "loadMisterIni");
    await expect(
      loadMisterIni(3, useIniSettingsStore.getState()),
    ).rejects.toThrow("unavailable");
    expect(load).not.toHaveBeenCalled();
    expect(useIniSettingsStore.getState().loadedIni).toBeNull();
  });

  it("resets settings omitted by the target file on explicit switch", async () => {
    vi.spyOn(ControlApi.prototype, "loadMisterIni")
      .mockResolvedValueOnce({ font: "custom/font" })
      .mockResolvedValueOnce({});
    await loadMisterIni(1, useIniSettingsStore.getState());
    await loadMisterIni(2, useIniSettingsStore.getState(), true);
    expect(useIniSettingsStore.getState().font).toBe("");
    expect(useIniSettingsStore.getState().loadedIni?.filename).toBe(
      "MiSTer_auto.ini",
    );
  });

  it("ignores late initial active-slot discovery after an explicit switch", async () => {
    const listing = await ControlApi.prototype.listMisterInis();
    let complete!: (value: typeof listing) => void;
    vi.mocked(ControlApi.prototype.listMisterInis).mockReturnValueOnce(
      new Promise((resolve) => {
        complete = resolve;
      }),
    );
    const request = vi
      .spyOn(ControlApi.prototype, "loadMisterIni")
      .mockResolvedValue({});
    const initial = loadMisterIni(undefined, useIniSettingsStore.getState());
    await loadMisterIni(4, useIniSettingsStore.getState(), true);
    complete(listing);
    await initial;
    expect(useIniSettingsStore.getState().loadedIni?.id).toBe(4);
    expect(request).toHaveBeenCalledTimes(1);
    expect(request).toHaveBeenCalledWith(4, "MiSTer_ultrawide.ini");
  });

  it("ignores an old load that completes after switching", async () => {
    let complete!: (value: { [key: string]: string }) => void;
    const old = new Promise<{ [key: string]: string }>((resolve) => {
      complete = resolve;
    });
    vi.spyOn(ControlApi.prototype, "loadMisterIni")
      .mockReturnValueOnce(old)
      .mockResolvedValueOnce({ font: "new/font" });
    const first = loadMisterIni(1, useIniSettingsStore.getState());
    await Promise.resolve();
    await loadMisterIni(2, useIniSettingsStore.getState(), true);
    complete({ font: "old/font" });
    await first;
    expect(useIniSettingsStore.getState().font).toBe("new/font");
    expect(useIniSettingsStore.getState().loadedIni?.id).toBe(2);
  });

  it("ignores save completion after the editor switches files", async () => {
    vi.spyOn(ControlApi.prototype, "loadMisterIni").mockResolvedValue({});
    let complete!: () => void;
    vi.spyOn(ControlApi.prototype, "saveMisterIni").mockReturnValue(
      new Promise<void>((resolve) => {
        complete = resolve;
      }),
    );
    await loadMisterIni(1, useIniSettingsStore.getState());
    useIniSettingsStore.getState().setFont("old/font");
    const save = saveMisterIni(1, useIniSettingsStore.getState());
    await loadMisterIni(2, useIniSettingsStore.getState(), true);
    useIniSettingsStore.getState().setFont("new/font");
    complete();
    await save;
    expect(useIniSettingsStore.getState().original.font).toBe("");
    expect(useIniSettingsStore.getState().modified).toContain("font");
  });

  it("keeps failed loads unsaveable and reports the failure", async () => {
    vi.spyOn(ControlApi.prototype, "loadMisterIni").mockRejectedValue(
      new Error("layout changed"),
    );
    const save = vi.spyOn(ControlApi.prototype, "saveMisterIni");
    await expect(
      loadMisterIni(1, useIniSettingsStore.getState()),
    ).rejects.toThrow("layout changed");
    expect(useIniSettingsStore.getState().iniError).toBe("layout changed");
    expect(useIniSettingsStore.getState().loadingIni).toBe(false);
    await expect(
      saveMisterIni(1, useIniSettingsStore.getState()),
    ).rejects.toThrow();
    expect(save).not.toHaveBeenCalled();
  });

  it("blocks overlapping saves even with the same stale UI snapshot", async () => {
    vi.spyOn(ControlApi.prototype, "loadMisterIni").mockResolvedValue({});
    await loadMisterIni(1, useIniSettingsStore.getState());
    let complete!: () => void;
    const request = vi
      .spyOn(ControlApi.prototype, "saveMisterIni")
      .mockReturnValue(
        new Promise<void>((resolve) => {
          complete = resolve;
        }),
      );
    const snapshot = useIniSettingsStore.getState();
    const first = saveMisterIni(1, snapshot);
    await expect(saveMisterIni(1, snapshot)).rejects.toThrow();
    expect(request).toHaveBeenCalledTimes(1);
    complete();
    await first;
    expect(useIniSettingsStore.getState().savingIni).toBe(false);
  });

  it("clears modified keys when all changes are reverted", () => {
    const store = useIniSettingsStore.getState();
    store.setHostname("temporary");
    expect(useIniSettingsStore.getState().modified).toContain("hostname");

    useIniSettingsStore.getState().revertChanges();

    const reverted = useIniSettingsStore.getState();
    expect(reverted.hostname).toBe("MiSTer");
    expect(reverted.modified).toEqual([]);
  });

  it("preserves unsaved camel-case keys when settings reload", async () => {
    vi.spyOn(ControlApi.prototype, "loadMisterIni").mockResolvedValue({
      composite_sync: "0",
    });
    useIniSettingsStore.getState().setCompositeSync("1");

    await loadMisterIni(1, useIniSettingsStore.getState());

    const current = useIniSettingsStore.getState();
    expect(current.compositeSync).toBe("1");
    expect(current.modified).toContain("compositeSync");
  });

  it("applies target values when switching INI files", async () => {
    vi.spyOn(ControlApi.prototype, "loadMisterIni").mockResolvedValue({
      composite_sync: "0",
    });
    useIniSettingsStore.getState().setCompositeSync("1");

    await loadMisterIni(2, useIniSettingsStore.getState(), true);

    const current = useIniSettingsStore.getState();
    expect(current.compositeSync).toBe("0");
    expect(current.modified).toEqual([]);
  });

  it("preserves edits made while an INI load is pending", async () => {
    let completeLoad: ((value: { composite_sync: string }) => void) | undefined;
    const pendingLoad = new Promise<{ composite_sync: string }>((resolve) => {
      completeLoad = resolve;
    });
    vi.spyOn(ControlApi.prototype, "loadMisterIni").mockReturnValue(
      pendingLoad,
    );

    const load = loadMisterIni(1, useIniSettingsStore.getState());
    useIniSettingsStore.getState().setCompositeSync("1");
    completeLoad?.({ composite_sync: "0" });
    await load;

    const current = useIniSettingsStore.getState();
    expect(current.compositeSync).toBe("1");
    expect(current.modified).toContain("compositeSync");
  });

  it("preserves edits made while a save is pending", async () => {
    let completeSave: (() => void) | undefined;
    const pendingSave = new Promise<void>((resolve) => {
      completeSave = resolve;
    });
    vi.spyOn(ControlApi.prototype, "saveMisterIni").mockReturnValue(
      pendingSave,
    );

    vi.spyOn(ControlApi.prototype, "loadMisterIni").mockResolvedValue({});
    await loadMisterIni(1, useIniSettingsStore.getState());
    useIniSettingsStore.getState().setHostname("first");
    const save = saveMisterIni(1, useIniSettingsStore.getState());
    useIniSettingsStore.getState().setHostname("second");
    useIniSettingsStore.getState().setFont("font/path");

    completeSave?.();
    await save;

    const current = useIniSettingsStore.getState();
    expect(current.hostname).toBe("second");
    expect(current.original.hostname).toBe("first");
    expect(current.modified).toEqual(["hostname", "font"]);
  });
});
