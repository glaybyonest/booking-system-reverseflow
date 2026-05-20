import type { Booking, HoldSeatResponse } from "@/entities/booking/types";
import { clientFetch } from "@/shared/api/client";
import { normalizeBooking, normalizeHoldSeatResponse } from "@/shared/api/mappers";
import type { ApiList } from "@/shared/types/api";

export async function holdSeat(input: { sessionId: string; seatIds: string[] }) {
  const response = await clientFetch<unknown>("bookings/hold", {
    method: "POST",
    json: input
  });
  return normalizeHoldSeatResponse(response);
}

export async function getBooking(bookingId: string): Promise<Booking> {
  return normalizeBooking(await clientFetch(`bookings/${bookingId}`));
}

export async function getMyBookings(): Promise<Booking[]> {
  const data = await clientFetch<ApiList<unknown>>("bookings/me");
  return (data.items ?? []).map(normalizeBooking);
}

export async function cancelBooking(bookingId: string) {
  return clientFetch<{ status: string }>(`bookings/${bookingId}/cancel`, { method: "POST" });
}
