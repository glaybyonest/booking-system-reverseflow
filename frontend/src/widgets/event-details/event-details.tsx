"use client";

import { friendlyApiError } from "@/shared/api/errors";
import { Badge } from "@/shared/ui/badge";
import { Card, CardContent } from "@/shared/ui/card";
import { Alert } from "@/shared/ui/alert";
import { Spinner } from "@/shared/ui/spinner";
import { useEvent } from "@/features/event-list/event-list.hooks";

export function EventDetails({ eventId }: { eventId: string }) {
  const event = useEvent(eventId);
  if (event.isLoading) return <Spinner />;
  if (event.error) return <Alert variant="error">{friendlyApiError(event.error)}</Alert>;
  if (!event.data) return null;
  return (
    <Card>
      <CardContent>
        <Badge>{event.data.category ?? "Мероприятие"}</Badge>
        <h1 className="mt-5 text-3xl font-bold tracking-tight md:text-4xl">{event.data.title}</h1>
        <p className="mt-4 text-base leading-relaxed text-gray-500">
          {event.data.description ?? "Описание появится позже."}
        </p>
      </CardContent>
    </Card>
  );
}
