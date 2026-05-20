"use client";

import { useQuery } from "@tanstack/react-query";

import { getEventById, getEvents, getEventsMap, type EventsQuery } from "@/shared/api/events.api";

export function useEvents(query: EventsQuery) {
  return useQuery({
    queryKey: ["events", query],
    queryFn: () => getEvents(query)
  });
}

export function useEvent(eventId: string) {
  return useQuery({
    queryKey: ["events", eventId],
    queryFn: () => getEventById(eventId),
    enabled: Boolean(eventId)
  });
}

export function useEventsMap(query: EventsQuery) {
  return useQuery({
    queryKey: ["events", "map", query],
    queryFn: () => getEventsMap(query)
  });
}
