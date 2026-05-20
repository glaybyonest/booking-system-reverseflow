"use client";

import dynamic from "next/dynamic";

import { EventCatalogBar } from "@/features/event-list/event-list";
import { useEventsMap } from "@/features/event-list/event-list.hooks";
import { friendlyApiError } from "@/shared/api/errors";
import { defaultMapCenter } from "@/shared/config/map";
import { routes } from "@/shared/lib/routes";
import { Alert } from "@/shared/ui/alert";
import { EmptyState } from "@/shared/ui/empty-state";
import { Spinner } from "@/shared/ui/spinner";

const DynamicMapCanvas = dynamic(
  () => import("@/widgets/event-map/map-canvas").then((module) => module.MapCanvas),
  {
    ssr: false,
    loading: () => (
      <div className="flex h-[70vh] items-center justify-center rounded-2xl border border-border bg-white shadow-card">
        <Spinner />
      </div>
    )
  }
);

export function EventsMapPage() {
  const defaults = defaultMapCenter();
  const events = useEventsMap({ onlyActual: true, limit: 200, offset: 0 });

  if (events.isLoading) {
    return (
      <div className="flex justify-center py-20">
        <Spinner />
      </div>
    );
  }

  if (events.error) {
    return <Alert variant="error">{friendlyApiError(events.error)}</Alert>;
  }

  const items = events.data ?? [];

  return (
    <div className="space-y-6">
      <EventCatalogBar total={items.length} actionHref={routes.events} actionLabel="К каталогу" />
      {items.length ? (
        <DynamicMapCanvas
          events={items}
          center={{ lat: defaults.lat, lon: defaults.lon }}
          zoom={defaults.zoom}
        />
      ) : (
        <EmptyState
          title="На карте пока нет событий"
          description="Импорт событий выполняется автоматически. Если карта ещё пустая, обновите страницу через минуту."
        />
      )}
    </div>
  );
}
