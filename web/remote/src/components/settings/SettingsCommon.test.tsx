// @vitest-environment jsdom

import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { useState } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { NumberOption, SaveButton } from "./SettingsCommon";
import { useIniSettingsStore } from "../../lib/ini";
import { ControlApi } from "../../lib/api";

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
  useIniSettingsStore.getState().reset();
});

function NumberOptionHarness(props: { onChange: (value: string) => void }) {
  const [value, setValue] = useState("10");
  return (
    <NumberOption
      value={value}
      setValue={(next) => {
        props.onChange(next);
        setValue(next);
      }}
      label="Offset"
      min={-100}
      max={100}
      defaultValue="0"
      disabledValue=""
    />
  );
}

describe("SaveButton", () => {
  it("shows and saves the loaded filename without re-fetching active slot", async () => {
    useIniSettingsStore.setState({
      loadedIni: {
        id: 4,
        filename: "MiSTer_ultrawide.ini",
        displayName: "ultrawide",
        path: "/media/fat/MiSTer_ultrawide.ini",
      },
    });
    useIniSettingsStore.getState().setFont("test/font");
    const list = vi.spyOn(ControlApi.prototype, "listMisterInis");
    const save = vi
      .spyOn(ControlApi.prototype, "saveMisterIni")
      .mockResolvedValue();
    render(<SaveButton />);
    expect(
      screen.getByText("Editing: ultrawide (MiSTer_ultrawide.ini)"),
    ).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    await waitFor(() =>
      expect(save).toHaveBeenCalledWith(
        4,
        { font: "test/font" },
        "MiSTer_ultrawide.ini",
      ),
    );
    expect(list).not.toHaveBeenCalled();
  });

  it("explains that Save creates missing Main without editing the example", () => {
    render(<SaveButton />);
    expect(
      screen.getByText(/Saving Main creates MiSTer.ini if it is missing/),
    ).toBeTruthy();
    expect(screen.getByText(/MiSTer_example.ini is never edited/)).toBeTruthy();
  });
});

describe("NumberOption", () => {
  it("preserves an incomplete negative value without committing zero", () => {
    const onChange = vi.fn();
    render(<NumberOptionHarness onChange={onChange} />);
    const input = screen.getByLabelText("Offset value");

    fireEvent.change(input, { target: { value: "-" } });

    expect((input as HTMLInputElement).value).toBe("-");
    expect(onChange).not.toHaveBeenCalled();
  });

  it("commits complete in-range negative values", () => {
    const onChange = vi.fn();
    render(<NumberOptionHarness onChange={onChange} />);
    const input = screen.getByLabelText("Offset value");

    fireEvent.change(input, { target: { value: "-25" } });

    expect(onChange).toHaveBeenLastCalledWith("-25");
    expect((input as HTMLInputElement).value).toBe("-25");
  });
});
