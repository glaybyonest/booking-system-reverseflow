import { fireEvent, render, screen } from "@testing-library/react";
import React from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { HomeEventsShowcase } from "@/features/home/home-events-showcase";

const useEventsMock = vi.fn();

vi.mock("@/features/event-list/event-list.hooks", () => ({
  useEvents: (...args: unknown[]) => useEventsMock(...args)
}));

describe("HomeEventsShowcase", () => {
  beforeEach(() => {
    useEventsMock.mockReset();
  });

  it("renders loading state while events are being fetched", () => {
    useEventsMock.mockReturnValue({
      data: undefined,
      error: null,
      isLoading: true
    });

    const { container } = render(<HomeEventsShowcase />);

    expect(screen.getByText("Популярное")).toBeInTheDocument();
    expect(container.querySelector(".animate-pulse")).toBeTruthy();
  });

  it("filters events by search and category without touching the catalog page", () => {
    useEventsMock.mockReturnValue({
      data: {
        items: [
          {
            id: "1",
            title: "Большой концерт в парке",
            category: "concert",
            source: "kudago",
            bookingMode: "reserveflow_managed",
            status: "published",
            startsAt: "2026-05-15T19:00:00+03:00",
            venue: { id: "v1", name: "Зелёный театр", address: "", city: "Москва" }
          },
          {
            id: "2",
            title: "Ночная выставка",
            category: "exhibition",
            source: "kudago",
            bookingMode: "reserveflow_managed",
            status: "published",
            startsAt: "2026-05-16T18:00:00+03:00",
            venue: { id: "v2", name: "Галерея", address: "", city: "Москва" }
          }
        ],
        total: 2
      },
      error: null,
      isLoading: false
    });

    render(<HomeEventsShowcase />);

    fireEvent.click(screen.getByRole("button", { name: "Концерты" }));
    expect(screen.getByText("Большой концерт в парке")).toBeInTheDocument();
    expect(screen.queryByText("Ночная выставка")).not.toBeInTheDocument();

    fireEvent.change(screen.getByPlaceholderText("Концерт, спектакль, выставка или площадка"), {
      target: { value: "парк" }
    });
    expect(screen.getByText("Большой концерт в парке")).toBeInTheDocument();
  });
});
