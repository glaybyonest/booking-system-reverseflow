import type { Event } from "@/entities/event/types";
import type { Session } from "@/entities/session/types";

export type BookingStatus = "pending" | "confirmed" | "cancelled" | "expired" | "payment_failed";

export type BookingItem = {
  id: string;
  seatId: string;
  row?: string;
  number?: number;
  price: number;
  status: string;
};

export type Booking = {
  id: string;
  bookingId?: string;
  userId?: string;
  sessionId: string;
  status: BookingStatus;
  expiresAt?: string | null;
  totalPrice: number;
  createdAt?: string;
  updatedAt?: string;
  items?: BookingItem[];
  event?: Event;
  session?: Session;
};

export type HoldSeatResponse = {
  bookingId: string;
  status: BookingStatus;
  expiresAt: string;
  seats: Array<{
    seatId: string;
    row: string;
    number: number;
  }>;
  totalPrice: number;
};
