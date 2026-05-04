import type { Event } from "@/entities/event/types";
import type { Session } from "@/entities/session/types";
import { clientFetch } from "@/shared/api/client";
import { normalizeEvent, normalizeSession } from "@/shared/api/mappers";
import type { ApiList } from "@/shared/types/api";

export async function getEvents() {
  const data = await clientFetch<ApiList<unknown>>("events");
  return (data.items ?? []).map(normalizeEvent);
}

export async function getEventById(eventId: string): Promise<Event> {
  return normalizeEvent(await clientFetch(`events/${eventId}`));
}

export async function getEventSessions(eventId: string): Promise<Session[]> {
  const data = await clientFetch<ApiList<unknown>>(`events/${eventId}/sessions`);
  return (data.items ?? []).map(normalizeSession);
}
