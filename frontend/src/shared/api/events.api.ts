import type { Event } from "@/entities/event/types";
import { clientFetch } from "@/shared/api/client";
import { normalizeEvent } from "@/shared/api/mappers";
import type { ApiList } from "@/shared/types/api";

export type EventsQuery = {
  city?: string;
  source?: string;
  category?: string;
  from?: string;
  to?: string;
  bookingMode?: string;
  onlyActual?: boolean;
  limit?: number;
  offset?: number;
};

export type EventsListResult = {
  items: Event[];
  total: number;
};

export async function getEvents(query: EventsQuery = {}): Promise<EventsListResult> {
  const data = await clientFetch<ApiList<unknown>>(`events${toQueryString(query)}`);
  return {
    items: (data.items ?? []).map(normalizeEvent),
    total: typeof data.total === "number" ? data.total : (data.items ?? []).length
  };
}

export async function getEventById(eventId: string): Promise<Event> {
  return normalizeEvent(await clientFetch(`events/${eventId}`));
}

export async function getEventsMap(query: EventsQuery = {}): Promise<Event[]> {
  const data = await clientFetch<{ events?: unknown[] }>(`events/map${toQueryString(query)}`);
  return (data.events ?? []).map(normalizeEvent);
}

function toQueryString(query: EventsQuery) {
  const params = new URLSearchParams();
  if (query.city) params.set("city", query.city);
  if (query.source) params.set("source", query.source);
  if (query.category) params.set("category", query.category);
  if (query.from) params.set("from", query.from);
  if (query.to) params.set("to", query.to);
  if (query.bookingMode) params.set("bookingMode", query.bookingMode);
  if (query.onlyActual !== undefined) params.set("onlyActual", String(query.onlyActual));
  if (query.limit !== undefined) params.set("limit", String(query.limit));
  if (query.offset !== undefined) params.set("offset", String(query.offset));
  const suffix = params.toString();
  return suffix ? `?${suffix}` : "";
}
