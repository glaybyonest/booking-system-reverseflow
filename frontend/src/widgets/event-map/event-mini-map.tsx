"use client";

import dynamic from "next/dynamic";

import type { Event } from "@/entities/event/types";
import { defaultMapCenter } from "@/shared/config/map";
import { Spinner } from "@/shared/ui/spinner";

const DynamicMapCanvas = dynamic(
  () => import("@/widgets/event-map/map-canvas").then((m) => ({ default: m.MapCanvas })),
  {
    ssr: false,
    loading: () => (
      <div className="flex h-56 w-full items-center justify-center rounded-2xl border border-border bg-white">
        <Spinner />
      </div>
    )
  }
);

export function EventMiniMap({
  latitude,
  longitude,
  title
}: {
  latitude: number;
  longitude: number;
  title: string;
}) {
  const defaults = defaultMapCenter();
  const miniMapEvent: Event = {
    id: "mini-map-event",
    title,
    status: "published",
    source: "manual",
    bookingMode: "reserveflow_managed",
    venue: {
      id: "mini-map-venue",
      name: title,
      address: "",
      city: "Москва",
      latitude,
      longitude
    }
  };

  return (
    <DynamicMapCanvas
      events={[miniMapEvent]}
      center={{ lat: latitude, lon: longitude }}
      zoom={Math.max(defaults.zoom, 13)}
      heightClassName="h-56"
      interactive={false}
    />
  );
}
