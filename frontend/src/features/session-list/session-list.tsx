"use client";

import { friendlyApiError } from "@/shared/api/errors";
import { Alert } from "@/shared/ui/alert";
import { EmptyState } from "@/shared/ui/empty-state";
import { Spinner } from "@/shared/ui/spinner";
import { SessionCard } from "@/widgets/session-card/session-card";
import { useEventSessions } from "@/features/event-list/event-list.hooks";

export function SessionList({ eventId }: { eventId: string }) {
  const sessions = useEventSessions(eventId);
  if (sessions.isLoading) return <Spinner />;
  if (sessions.error) return <Alert variant="error">{friendlyApiError(sessions.error)}</Alert>;
  if (!sessions.data?.length) {
    return <EmptyState title="Сеансов пока нет" description="Для этого мероприятия еще нет расписания." />;
  }
  return (
    <div className="space-y-4">
      {sessions.data.map((session) => (
        <SessionCard key={session.id} session={session} />
      ))}
    </div>
  );
}
