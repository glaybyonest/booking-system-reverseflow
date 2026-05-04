"use client";

import { useQuery } from "@tanstack/react-query";

import { getEventById, getEventSessions, getEvents } from "@/shared/api/events.api";

export function useEvents() {
  return useQuery({
    queryKey: ["events"],
    queryFn: getEvents
  });
}

export function useEvent(eventId: string) {
  return useQuery({
    queryKey: ["events", eventId],
    queryFn: () => getEventById(eventId),
    enabled: Boolean(eventId)
  });
}

export function useEventSessions(eventId: string) {
  return useQuery({
    queryKey: ["events", eventId, "sessions"],
    queryFn: () => getEventSessions(eventId),
    enabled: Boolean(eventId)
  });
}
