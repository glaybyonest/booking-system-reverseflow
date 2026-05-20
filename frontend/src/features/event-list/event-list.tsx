"use client";

import Link from "next/link";

import { useEvents } from "@/features/event-list/event-list.hooks";
import { friendlyApiError } from "@/shared/api/errors";
import { routes } from "@/shared/lib/routes";
import { Alert } from "@/shared/ui/alert";
import { EmptyState } from "@/shared/ui/empty-state";
import { Spinner } from "@/shared/ui/spinner";
import { EventCard } from "@/widgets/event-card/event-card";

export function EventList() {
  const events = useEvents({ onlyActual: false, limit: 60, offset: 0 });

  if (events.isLoading) {
    return (
      <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 xl:grid-cols-3">
        {Array.from({ length: 6 }, (_, i) => (
          <div key={i} className="h-[280px] animate-pulse rounded-2xl bg-white ring-1 ring-border" />
        ))}
      </div>
    );
  }
  if (events.error) {
    return <Alert variant="error">{friendlyApiError(events.error)}</Alert>;
  }

  const items = events.data?.items ?? [];
  const total = events.data?.total ?? items.length;

  return (
    <div className="space-y-6">
      <EventCatalogBar total={total} actionHref={routes.eventsMap} actionLabel="На карте" />
      {!items.length ? (
        <EmptyState
          title="Мероприятий пока нет"
          description="Каталог обновляется автоматически. Если события ещё не появились, обновите страницу чуть позже."
        />
      ) : (
        <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 xl:grid-cols-3">
          {items.map((event) => (
            <EventCard key={event.id} event={event} />
          ))}
        </div>
      )}
    </div>
  );
}

export function EventCatalogBar({
  total,
  actionHref,
  actionLabel
}: {
  total: number;
  actionHref: string;
  actionLabel: string;
}) {
  return (
    <div className="flex items-center justify-between rounded-2xl border border-border bg-white px-5 py-4 shadow-card">
      <p className="text-sm font-medium text-mute">
        {total > 0 ? `${total} событий в каталоге` : "События загружаются"}
      </p>
      <Link
        href={actionHref}
        className="rounded-full bg-ink px-5 py-2 text-[13px] font-medium text-white transition-colors hover:bg-ink-2"
      >
        {actionLabel}
      </Link>
    </div>
  );
}
