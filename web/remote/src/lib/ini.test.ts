// @vitest-environment jsdom

import { beforeEach, describe, expect, it, vi } from "vitest";

import { ControlApi } from "./api";
import { loadMisterIni, saveMisterIni, useIniSettingsStore } from "./ini";

describe("INI settings state", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    useIniSettingsStore.getState().reset();
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

  it("preserves edits made while a save is pending", async () => {
    let completeSave: (() => void) | undefined;
    const pendingSave = new Promise<void>((resolve) => {
      completeSave = resolve;
    });
    vi.spyOn(ControlApi.prototype, "saveMisterIni").mockReturnValue(
      pendingSave,
    );

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
