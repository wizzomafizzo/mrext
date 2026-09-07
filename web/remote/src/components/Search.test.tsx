// @vitest-environment jsdom

import { cleanup, render, screen } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import Search from "./Search";

const mocks = vi.hoisted(() => ({
  search: {
    ready: true,
    indexing: false,
    totalSteps: 0,
    currentStep: 0,
    currentDesc: "",
  },
  reset: vi.fn(),
  refetch: vi.fn(),
  sendMessage: vi.fn(),
}));

vi.mock("../lib/store", () => ({
  useServerStateStore: () => ({ search: mocks.search }),
}));
vi.mock("../lib/queries", () => ({
  useIndexedSystems: () => ({ data: { systems: [] }, refetch: mocks.refetch }),
}));
vi.mock("./WebSocket", () => ({
  default: () => ({ readyState: 1, sendMessage: mocks.sendMessage }),
}));
vi.mock("@tanstack/react-query", () => ({
  useMutation: () => ({
    mutate: vi.fn(),
    reset: mocks.reset,
    isLoading: false,
  }),
}));
vi.mock("./ScrollToTop", () => ({ default: () => null }));
vi.mock("./Shortcuts", () => ({ SingleShortcut: () => null }));

afterEach(cleanup);
beforeEach(() => {
  vi.clearAllMocks();
  Object.assign(mocks.search, {
    ready: true,
    indexing: false,
    totalSteps: 0,
    currentStep: 0,
    currentDesc: "",
  });
});

describe("index regeneration", () => {
  it("shows failed regeneration while keeping the previous index searchable", () => {
    mocks.search.currentDesc = "Index failed: disk full";
    render(<Search />);
    expect(screen.getByRole("alert").textContent).toContain(
      "Index failed: disk full",
    );
    expect(screen.getByPlaceholderText("Search")).toBeTruthy();
  });

  it("shows first-build failure with a retry button", () => {
    mocks.search.ready = false;
    mocks.search.currentDesc = "Index failed: unreadable game folder";
    render(<Search />);
    expect(screen.getByRole("alert").textContent).toContain(
      "unreadable game folder",
    );
    expect(screen.getByRole("button", { name: "Generate Index" })).toBeTruthy();
  });

  it("invalidates old results and system choices only after success", () => {
    mocks.search.indexing = true;
    const view = render(<Search />);
    mocks.search.indexing = false;
    view.rerender(<Search />);
    expect(mocks.reset).toHaveBeenCalledOnce();
    expect(mocks.refetch).toHaveBeenCalledOnce();
    expect(screen.queryByRole("alert")).toBeNull();
  });

  it("keeps cached results when regeneration fails", () => {
    mocks.search.indexing = true;
    const view = render(<Search />);
    mocks.search.indexing = false;
    mocks.search.currentDesc = "Index failed: scan interrupted";
    view.rerender(<Search />);
    expect(mocks.reset).not.toHaveBeenCalled();
    expect(mocks.refetch).not.toHaveBeenCalled();
  });
});
