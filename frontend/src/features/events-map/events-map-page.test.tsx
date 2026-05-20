import { render } from "@testing-library/react";
import React from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { EventsMapPage } from "@/features/events-map/events-map-page";

const replaceMock = vi.fn();
const useEventsMapMock = vi.fn();

vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace: replaceMock }),
  usePathname: () => "/events/map",
  useSearchParams: () => new URLSearchParams()
}));

vi.mock("@/features/event-list/event-list.hooks", () => ({
  useEventsMap: (...args: unknown[]) => useEventsMapMock(...args)
}));

describe("EventsMapPage", () => {
  beforeEach(() => {
    replaceMock.mockReset();
    useEventsMapMock.mockReset();
  });

  it("renders loading fallback while map data is loading", () => {
    useEventsMapMock.mockReturnValue({
      isLoading: true,
      error: null,
      data: undefined
    });

    const { container } = render(<EventsMapPage />);

    expect(container.querySelector(".animate-spin")).toBeTruthy();
  });
});
