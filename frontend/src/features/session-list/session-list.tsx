"use client";

import type { Session } from "@/entities/session/types";
import { useMe } from "@/features/auth/auth.hooks";
import { EmptyState } from "@/shared/ui/empty-state";
import { SessionCard } from "@/widgets/session-card/session-card";

export function SessionList({ sessions }: { sessions: Session[] }) {
  const me = useMe();

  if (!sessions.length) {
    return (
      <EmptyState
        title="Сеансов пока нет"
        description="Для этого мероприятия расписание еще не опубликовано."
      />
    );
  }
  return (
    <div className="space-y-4">
      {sessions.map((session) => (
        <SessionCard key={session.id} session={session} isAdmin={me.data?.role === "admin"} />
      ))}
    </div>
  );
}
