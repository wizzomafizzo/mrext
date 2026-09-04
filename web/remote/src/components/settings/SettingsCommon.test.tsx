// @vitest-environment jsdom

import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { useState } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { NumberOption } from "./SettingsCommon";

afterEach(cleanup);

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
