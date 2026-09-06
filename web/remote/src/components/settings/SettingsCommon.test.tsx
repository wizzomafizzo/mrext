// @vitest-environment jsdom

import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { useState } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";

import { NumberOption } from "./SettingsCommon";

afterEach(cleanup);

function NumberOptionHarness(props: {
  onChange: (value: string) => void;
  min?: number;
}) {
  const [value, setValue] = useState("10");
  return (
    <NumberOption
      value={value}
      setValue={(next) => {
        props.onChange(next);
        setValue(next);
      }}
      label="Offset"
      min={props.min ?? -100}
      max={100}
      defaultValue="0"
      disabledValue=""
    />
  );
}

describe("NumberOption", () => {
  it("allows clearing and replacing a value above a positive minimum", () => {
    const onChange = vi.fn();
    render(<NumberOptionHarness onChange={onChange} min={5} />);
    const input = screen.getByLabelText("Offset value") as HTMLInputElement;

    fireEvent.change(input, { target: { value: "" } });
    expect(input.value).toBe("");
    expect(onChange).not.toHaveBeenCalled();
    expect((screen.getByRole("checkbox") as HTMLInputElement).checked).toBe(
      true,
    );

    fireEvent.change(input, { target: { value: "2" } });
    expect(input.value).toBe("2");
    expect(onChange).not.toHaveBeenCalled();
    fireEvent.change(input, { target: { value: "25" } });
    fireEvent.blur(input);
    expect(input.value).toBe("25");
    expect(onChange).toHaveBeenLastCalledWith("25");
    expect(input.inputMode).toBe("decimal");
  });

  it("clamps out-of-range drafts only on blur and restores an empty draft", () => {
    const onChange = vi.fn();
    render(<NumberOptionHarness onChange={onChange} />);
    const input = screen.getByLabelText("Offset value") as HTMLInputElement;

    fireEvent.change(input, { target: { value: "200" } });
    expect(input.value).toBe("200");
    expect(onChange).not.toHaveBeenCalled();
    fireEvent.blur(input);
    expect(input.value).toBe("100");
    expect(onChange).toHaveBeenLastCalledWith("100");

    onChange.mockClear();
    fireEvent.change(input, { target: { value: "" } });
    fireEvent.blur(input);
    expect(input.value).toBe("100");
    expect(onChange).not.toHaveBeenCalled();
  });

  it("retains disabled and default values when toggled", () => {
    const onChange = vi.fn();
    render(<NumberOptionHarness onChange={onChange} />);
    const checkbox = screen.getByRole("checkbox");

    fireEvent.click(checkbox);
    expect(onChange).toHaveBeenLastCalledWith("");
    expect(screen.queryByLabelText("Offset value")).toBeNull();
    fireEvent.click(checkbox);
    expect(onChange).toHaveBeenLastCalledWith("0");
    expect(
      (screen.getByLabelText("Offset value") as HTMLInputElement).value,
    ).toBe("0");
  });

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
