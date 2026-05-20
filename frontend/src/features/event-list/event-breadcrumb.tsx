"use client";

import { useEvent } from "@/features/event-list/event-list.hooks";

export function EventBreadcrumb({ eventId }: { eventId: string }) {
  const event = useEvent(eventId);
  // Show skeleton while loading, then the real category (or a fallback)
  return (
    <span className="text-mute-2">
      {event.isLoading ? "…" : (event.data?.category ?? "Мероприятие")}
    </span>
  );
}
