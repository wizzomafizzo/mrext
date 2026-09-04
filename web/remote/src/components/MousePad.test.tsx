// @vitest-environment jsdom

import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
} from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import MousePad from "./MousePad";

describe("MousePad", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    HTMLElement.prototype.setPointerCapture = vi.fn();
  });

  afterEach(() => {
    cleanup();
    vi.runOnlyPendingTimers();
    vi.useRealTimers();
  });

  it("does not turn a canceled pointer gesture into a click", () => {
    const sendMessage = vi.fn();
    render(<MousePad open onClose={vi.fn()} sendMessage={sendMessage} />);

    const pad = screen.getByTestId("mouse-pad");
    fireEvent.pointerDown(pad, { clientX: 20, clientY: 20 });
    fireEvent.pointerCancel(pad, { clientX: 20, clientY: 20 });
    act(() => vi.advanceTimersByTime(500));

    expect(sendMessage).not.toHaveBeenCalled();
  });

  it("releases a long-press drag when its pointer is canceled", () => {
    const sendMessage = vi.fn();
    render(<MousePad open onClose={vi.fn()} sendMessage={sendMessage} />);

    const pad = screen.getByTestId("mouse-pad");
    fireEvent.pointerDown(pad, { clientX: 20, clientY: 20 });
    act(() => vi.advanceTimersByTime(400));
    fireEvent.pointerCancel(pad, { clientX: 20, clientY: 20 });

    expect(sendMessage.mock.calls).toEqual([
      ["mouseBtn:left_down"],
      ["mouseBtn:left_up"],
    ]);
  });
});
