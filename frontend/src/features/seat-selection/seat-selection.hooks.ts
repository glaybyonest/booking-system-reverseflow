"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { holdSeat } from "@/shared/api/bookings.api";
import { getSession, getSessionSeats } from "@/shared/api/sessions.api";

export function useSession(sessionId: string) {
  return useQuery({
    queryKey: ["sessions", sessionId],
    queryFn: () => getSession(sessionId),
    enabled: Boolean(sessionId)
  });
}

export function useSeatMap(sessionId: string) {
  return useQuery({
    queryKey: ["sessions", sessionId, "seats"],
    queryFn: () => getSessionSeats(sessionId),
    enabled: Boolean(sessionId),
    refetchInterval: 5000,
    refetchOnWindowFocus: true
  });
}

export function useHoldSeat() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: holdSeat,
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ["sessions", variables.sessionId, "seats"] });
      queryClient.invalidateQueries({ queryKey: ["bookings"] });
    }
  });
}
