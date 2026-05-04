import type { Booking, BookingItem, HoldSeatResponse } from "@/entities/booking/types";
import type { Event } from "@/entities/event/types";
import type { Notification } from "@/entities/notification/types";
import type { Payment } from "@/entities/payment/types";
import type { Seat, SeatMapResponse } from "@/entities/seat/types";
import type { Session } from "@/entities/session/types";
import type { User } from "@/entities/user/types";

type AnyRecord = Record<string, unknown>;

function asRecord(value: unknown): AnyRecord {
  return typeof value === "object" && value !== null ? (value as AnyRecord) : {};
}

function asString(value: unknown, fallback = "") {
  return typeof value === "string" ? value : fallback;
}

function optionalString(value: unknown) {
  return typeof value === "string" ? value : undefined;
}

function asNumber(value: unknown, fallback = 0) {
  if (typeof value === "number") return value;
  if (typeof value === "string") {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : fallback;
  }
  return fallback;
}

export function normalizeUser(value: unknown): User {
  const item = asRecord(value);
  return {
    id: asString(item.id),
    email: asString(item.email),
    name: asString(item.name),
    role: asString(item.role, "user")
  };
}

export function normalizeEvent(value: unknown): Event {
  const item = asRecord(value);
  return {
    id: asString(item.id),
    title: asString(item.title),
    description: (item.description as string | null | undefined) ?? undefined,
    category: (item.category as string | null | undefined) ?? undefined,
    posterUrl:
      (item.posterUrl as string | null | undefined) ??
      (item.poster_url as string | null | undefined) ??
      undefined,
    status: asString(item.status),
    createdAt: optionalString(item.createdAt ?? item.created_at),
    updatedAt: optionalString(item.updatedAt ?? item.updated_at)
  };
}

export function normalizeSession(value: unknown): Session {
  const item = asRecord(value);
  const event = asRecord(item.event);
  const hall = asRecord(item.hall);
  return {
    id: asString(item.id),
    eventId: optionalString(item.eventId ?? item.event_id),
    hallId: optionalString(item.hallId ?? item.hall_id),
    startsAt: optionalString(item.startsAt ?? item.starts_at),
    endsAt: optionalString(item.endsAt ?? item.ends_at),
    status: asString(item.status),
    event:
      event.id || event.title
        ? {
            id: asString(event.id),
            title: asString(event.title)
          }
        : undefined,
    hall:
      hall.id || hall.name
        ? {
            id: optionalString(hall.id),
            name: asString(hall.name),
            venue: optionalString(hall.venue)
          }
        : undefined,
    hallName: optionalString(item.hallName ?? item.hall_name)
  };
}

export function normalizeSeat(value: unknown): Seat {
  const item = asRecord(value);
  return {
    seatId: asString(item.seatId ?? item.seat_id),
    row: asString(item.row),
    number: asNumber(item.number),
    status: asString(item.status, "available") as Seat["status"],
    price: asNumber(item.price)
  };
}

export function normalizeSeatMap(value: unknown): SeatMapResponse {
  const item = asRecord(value);
  const event = asRecord(item.event);
  const hall = asRecord(item.hall);
  return {
    sessionId: asString(item.sessionId ?? item.session_id),
    event:
      event.id || event.title
        ? {
            id: asString(event.id),
            title: asString(event.title)
          }
        : undefined,
    hall: {
      id: optionalString(hall.id),
      name: asString(hall.name)
    },
    seats: Array.isArray(item.seats) ? item.seats.map(normalizeSeat) : []
  };
}

export function normalizeBookingItem(value: unknown): BookingItem {
  const item = asRecord(value);
  return {
    id: asString(item.id),
    seatId: asString(item.seatId ?? item.seat_id),
    row: optionalString(item.row),
    number: item.number === undefined ? undefined : asNumber(item.number),
    price: asNumber(item.price),
    status: asString(item.status)
  };
}

export function normalizeBooking(value: unknown): Booking {
  const item = asRecord(value);
  const seat = asRecord(item.seat);
  return {
    id: asString(item.id ?? item.bookingId ?? item.booking_id),
    bookingId: optionalString(item.bookingId ?? item.booking_id),
    userId: optionalString(item.userId ?? item.user_id),
    sessionId: asString(item.sessionId ?? item.session_id),
    status: asString(item.status, "pending") as Booking["status"],
    expiresAt: (item.expiresAt as string | null | undefined) ?? (item.expires_at as string | null | undefined),
    totalPrice: asNumber(item.totalPrice ?? item.total_price),
    createdAt: optionalString(item.createdAt ?? item.created_at),
    updatedAt: optionalString(item.updatedAt ?? item.updated_at),
    items: Array.isArray(item.items) ? item.items.map(normalizeBookingItem) : undefined,
    event: item.event ? normalizeEvent(item.event) : undefined,
    session: item.session ? normalizeSession(item.session) : undefined,
    seat:
      seat.seatId || seat.seat_id
        ? {
            seatId: asString(seat.seatId ?? seat.seat_id),
            row: asString(seat.row),
            number: asNumber(seat.number)
          }
        : undefined
  };
}

export function normalizeHoldSeatResponse(value: unknown): HoldSeatResponse {
  const item = asRecord(value);
  const seat = asRecord(item.seat);
  return {
    bookingId: asString(item.bookingId ?? item.booking_id),
    status: asString(item.status, "pending") as HoldSeatResponse["status"],
    expiresAt: asString(item.expiresAt ?? item.expires_at),
    seat: {
      seatId: asString(seat.seatId ?? seat.seat_id),
      row: asString(seat.row),
      number: asNumber(seat.number)
    },
    totalPrice: asNumber(item.totalPrice ?? item.total_price)
  };
}

export function normalizePayment(value: unknown): Payment {
  const item = asRecord(value);
  return {
    id: asString(item.id ?? item.paymentId ?? item.payment_id),
    paymentId: optionalString(item.paymentId ?? item.payment_id),
    bookingId: asString(item.bookingId ?? item.booking_id),
    status: asString(item.status, "pending") as Payment["status"],
    amount: asNumber(item.amount),
    provider: asString(item.provider, "mock")
  };
}

export function normalizeNotification(value: unknown): Notification {
  const item = asRecord(value);
  return {
    id: asString(item.id),
    type: asString(item.type),
    title: asString(item.title),
    message: asString(item.message),
    isRead: Boolean(item.isRead ?? item.is_read),
    createdAt: optionalString(item.createdAt ?? item.created_at)
  };
}
