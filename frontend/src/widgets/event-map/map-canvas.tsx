"use client";

import Link from "next/link";
import { useEffect, useRef, useState } from "react";

import type { Event } from "@/entities/event/types";
import { formatDateTime } from "@/shared/lib/date";
import { displayLabel, sourceLabels } from "@/shared/lib/labels";
import { routes } from "@/shared/lib/routes";
import { Alert } from "@/shared/ui/alert";
import { Badge } from "@/shared/ui/badge";
import { Spinner } from "@/shared/ui/spinner";
import { loadYandexMapsApi, YandexMapsConfigError } from "@/widgets/event-map/yandex-maps";

const interactiveBehaviors = ["drag", "scrollZoom", "dblClick", "pinchZoom"] as const;

type MarkerEvent = Event & {
  venue: NonNullable<Event["venue"]> & {
    latitude: number;
    longitude: number;
  };
};

export function MapCanvas({
  events,
  center,
  zoom,
  heightClassName = "h-[70vh]",
  interactive = true
}: {
  events: Event[];
  center: { lat: number; lon: number };
  zoom: number;
  heightClassName?: string;
  interactive?: boolean;
}) {
  const containerRef = useRef<HTMLDivElement | null>(null);
  const mapRef = useRef<InstanceType<typeof ymaps3.YMap> | null>(null);
  const markerRefs = useRef<Array<InstanceType<typeof ymaps3.YMapMarker>>>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [mapError, setMapError] = useState<string | null>(null);
  const [selectedEventId, setSelectedEventId] = useState<string | null>(null);

  const markers = events.filter(hasCoordinates);
  const selectedEvent = markers.find((event) => event.id === selectedEventId) ?? null;

  useEffect(() => {
    let disposed = false;

    async function initMap() {
      if (!containerRef.current) {
        return;
      }

      setIsLoading(true);
      setMapError(null);

      try {
        const ymaps = await loadYandexMapsApi();
        if (disposed || !containerRef.current) {
          return;
        }

        const map = new ymaps.YMap(containerRef.current, {
          location: {
            center: [center.lon, center.lat],
            zoom
          },
          behaviors: interactive ? [...interactiveBehaviors] : []
        });

        map.addChild(new ymaps.YMapDefaultSchemeLayer({}));
        map.addChild(new ymaps.YMapDefaultFeaturesLayer({}));

        mapRef.current = map;
        setMapError(null);
      } catch (error) {
        if (!disposed) {
          setMapError(formatMapError(error));
        }
      } finally {
        if (!disposed) {
          setIsLoading(false);
        }
      }
    }

    void initMap();

    return () => {
      disposed = true;
      clearMarkers(mapRef.current, markerRefs.current);
      markerRefs.current = [];
      mapRef.current?.destroy();
      mapRef.current = null;
      if (containerRef.current) {
        containerRef.current.innerHTML = "";
      }
    };
  }, [center.lat, center.lon, interactive, zoom]);

  useEffect(() => {
    if (!selectedEventId) {
      return;
    }

    if (!markers.some((event) => event.id === selectedEventId)) {
      setSelectedEventId(null);
    }
  }, [markers, selectedEventId]);

  useEffect(() => {
    let disposed = false;

    async function syncMarkers() {
      if (isLoading || !mapRef.current) {
        return;
      }

      try {
        const ymaps = await loadYandexMapsApi();
        if (disposed || !mapRef.current) {
          return;
        }

        clearMarkers(mapRef.current, markerRefs.current);

        const nextMarkers = events.filter(hasCoordinates).map((event) => {
          const markerElement = document.createElement("button");
          markerElement.type = "button";
          markerElement.className = "event-map-marker";
          markerElement.setAttribute("aria-label", event.title);
          markerElement.innerHTML = '<span class="event-map-marker__dot"></span>';

          if (interactive) {
            markerElement.addEventListener("click", () => {
              setSelectedEventId(event.id);
            });
          } else {
            markerElement.disabled = true;
            markerElement.tabIndex = -1;
            markerElement.style.pointerEvents = "none";
          }

          const marker = new ymaps.YMapMarker(
            {
              coordinates: [event.venue.longitude, event.venue.latitude]
            },
            markerElement
          );

          mapRef.current?.addChild(marker);
          return marker;
        });

        markerRefs.current = nextMarkers;
        setMapError(null);
      } catch (error) {
        if (!disposed) {
          setMapError(formatMapError(error));
        }
      }
    }

    void syncMarkers();

    return () => {
      disposed = true;
    };
  }, [events, interactive, isLoading]);

  return (
    <div
      className={`relative rounded-2xl border border-border bg-white shadow-card ${heightClassName}`}
    >
      {/* Separate overflow-hidden wrapper so map tiles are clipped to rounded corners
          without affecting the stacking context of overlays and popups above it */}
      <div className="absolute inset-0 z-0 overflow-hidden rounded-2xl">
        <div ref={containerRef} className="h-full w-full" data-testid="yandex-map-canvas" />
      </div>

      {isLoading ? (
        <div className="absolute inset-0 z-10 flex items-center justify-center bg-white/80">
          <Spinner />
        </div>
      ) : null}

      {mapError ? (
        <div className="absolute inset-0 z-10 flex items-center justify-center bg-white/95 p-4">
          <div className="max-w-md">
            <Alert variant="error">{mapError}</Alert>
          </div>
        </div>
      ) : null}

      {interactive && selectedEvent ? (
        <div className="absolute bottom-4 left-4 z-10 max-w-sm rounded-2xl border border-border bg-white p-4 shadow-card">
          <div className="space-y-3 text-sm">
            <div className="flex items-start justify-between gap-3">
              <span className="rounded-full bg-bg px-2.5 py-1 text-[11px] font-semibold text-mute-2">
                {displayLabel(selectedEvent.source, sourceLabels)}
              </span>
              <button
                type="button"
                onClick={() => setSelectedEventId(null)}
                className="rounded-full border border-border px-2 py-1 text-xs text-mute hover:text-ink"
              >
                ✕
              </button>
            </div>
            <p className="font-bold text-ink line-clamp-2">{selectedEvent.title}</p>
            <p className="text-xs text-mute">
              {selectedEvent.startsAt ? formatDateTime(selectedEvent.startsAt) : "Дата уточняется"}
            </p>
            <p className="text-xs text-mute-2">
              {selectedEvent.venue.name || selectedEvent.venue.address || "Площадка уточняется"}
            </p>
            <Link
              href={routes.event(selectedEvent.id, selectedEvent.title)}
              className="inline-flex items-center gap-1 rounded-full bg-ink px-4 py-2 text-xs font-semibold text-white hover:bg-ink-2"
            >
              Подробнее →
            </Link>
          </div>
        </div>
      ) : null}
    </div>
  );
}

function hasCoordinates(event: Event): event is MarkerEvent {
  return event.venue?.latitude != null && event.venue?.longitude != null;
}

function clearMarkers(
  map: InstanceType<typeof ymaps3.YMap> | null,
  markers: Array<InstanceType<typeof ymaps3.YMapMarker>>
) {
  if (!map) {
    return;
  }

  for (const marker of markers) {
    map.removeChild(marker);
  }
}

function formatMapError(error: unknown) {
  if (error instanceof YandexMapsConfigError) {
    return error.message;
  }

  return "Не удалось загрузить карту Яндекса. Проверьте ключ и перезапустите frontend.";
}
