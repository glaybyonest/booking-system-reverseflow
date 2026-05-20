"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { cancelBooking, getMyBookings } from "@/shared/api/bookings.api";

export function useMyBookings() {
  return useQuery({
    queryKey: ["bookings", "me"],
    queryFn: getMyBookings,
    retry: false
  });
}

export function useCancelBooking() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: cancelBooking,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["bookings"] });
      queryClient.invalidateQueries({ queryKey: ["sessions"] });
    }
  });
}
