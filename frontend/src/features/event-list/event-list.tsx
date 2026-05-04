"use client";

import { friendlyApiError } from "@/shared/api/errors";
import { Alert } from "@/shared/ui/alert";
import { EmptyState } from "@/shared/ui/empty-state";
import { Spinner } from "@/shared/ui/spinner";
import { EventCard } from "@/widgets/event-card/event-card";
import { useEvents } from "@/features/event-list/event-list.hooks";

export function EventList() {
  const events = useEvents();

  if (events.isLoading) {
    return <Spinner />;
  }
  if (events.error) {
    return <Alert variant="error">{friendlyApiError(events.error)}</Alert>;
  }
  if (!events.data?.length) {
    return <EmptyState title="Мероприятий пока нет" description="Попробуйте обновить страницу позже." />;
  }
  return (
    <div className="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
      {events.data.map((event) => (
        <EventCard key={event.id} event={event} />
      ))}
    </div>
  );
}
