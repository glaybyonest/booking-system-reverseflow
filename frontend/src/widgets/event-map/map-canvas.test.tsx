import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import React from "react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import type { Event } from "@/entities/event/types";
import { EventMiniMap } from "@/widgets/event-map/event-mini-map";
import { MapCanvas } from "@/widgets/event-map/map-canvas";
import { loadYandexMapsApi } from "@/widgets/event-map/yandex-maps";

vi.mock("@/widgets/event-map/yandex-maps", () => ({
  YandexMapsConfigError: class YandexMapsConfigError extends Error {},
  loadYandexMapsApi: vi.fn()
}));

class FakeYMapMarker {
  constructor(
    public props: { coordinates: [number, number] },
    public element: HTMLElement
  ) {}
}

class FakeYMap {
  private readonly children: unknown[] = [];

  constructor(private readonly element: HTMLElement) {}

  addChild(child: unknown) {
    this.children.push(child);
    if (child instanceof FakeYMapMarker) {
      this.element.appendChild(child.element);
    }
    return this;
  }

  removeChild(child: unknown) {
    const index = this.children.indexOf(child);
    if (index >= 0) {
      this.children.splice(index, 1);
    }
    if (child instanceof FakeYMapMarker) {
      child.element.remove();
    }
    return this;
  }

  update() {}

  setBehaviors() {}

  destroy() {
    this.element.innerHTML = "";
  }
}

class FakeLayer {}

describe("MapCanvas", () => {
  beforeEach(() => {
    vi.mocked(loadYandexMapsApi).mockResolvedValue({
      ready: Promise.resolve(),
      YMap: FakeYMap as unknown as typeof ymaps3.YMap,
      YMapMarker: FakeYMapMarker as unknown as typeof ymaps3.YMapMarker,
      YMapDefaultSchemeLayer: FakeLayer as unknown as typeof ymaps3.YMapDefaultSchemeLayer,
      YMapDefaultFeaturesLayer: FakeLayer as unknown as typeof ymaps3.YMapDefaultFeaturesLayer
    } as typeof ymaps3);
  });

  it("renders markers only for events with coordinates and opens an overlay on click", async () => {
    const events: Event[] = [
      {
        id: "event-with-coords",
        title: "Концерт на карте",
        status: "published",
        source: "kudago",
        bookingMode: "external_link_only",
        startsAt: "2030-01-01T18:00:00+03:00",
        venue: {
          id: "venue-1",
          name: "Дом музыки",
          address: "Москва, Космодамианская набережная, 52",
          city: "Москва",
          latitude: 55.7343,
          longitude: 37.6467
        }
      },
      {
        id: "event-without-coords",
        title: "Событие без координат",
        status: "published",
        source: "kudago",
        bookingMode: "external_link_only",
        venue: {
          id: "venue-2",
          name: "Секретная площадка",
          address: "Москва",
          city: "Москва",
          latitude: null,
          longitude: null
        }
      }
    ];

    const { container } = render(
      <MapCanvas events={events} center={{ lat: 55.75, lon: 37.61 }} zoom={11} />
    );

    await waitFor(() => {
      expect(container.querySelectorAll(".event-map-marker")).toHaveLength(1);
    });

    fireEvent.click(screen.getByRole("button", { name: "Концерт на карте" }));

    expect(await screen.findByText("Концерт на карте")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Открыть" })).toBeInTheDocument();
  });

  it("renders a single static marker for the mini map", async () => {
    const { container } = render(
      <EventMiniMap latitude={55.7343} longitude={37.6467} title="Дом музыки" />
    );

    await waitFor(() => {
      expect(container.querySelectorAll(".event-map-marker")).toHaveLength(1);
    });

    expect(screen.queryByText("Закрыть")).not.toBeInTheDocument();
  });
});
