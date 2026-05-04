"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { getBooking } from "@/shared/api/bookings.api";
import { createPayment } from "@/shared/api/payments.api";

export function useBooking(bookingId: string) {
  return useQuery({
    queryKey: ["bookings", bookingId],
    queryFn: () => getBooking(bookingId),
    enabled: Boolean(bookingId),
    refetchInterval: (query) => (query.state.data?.status === "pending" ? 5000 : false)
  });
}

export function useCreatePayment() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: createPayment,
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ["bookings", variables.bookingId] });
      queryClient.invalidateQueries({ queryKey: ["bookings", "me"] });
      queryClient.invalidateQueries({ queryKey: ["sessions"] });
      queryClient.invalidateQueries({ queryKey: ["notifications"] });
    }
  });
}
