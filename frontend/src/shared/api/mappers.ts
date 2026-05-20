import type { Booking, BookingItem, HoldSeatResponse } from "@/entities/booking/types";
import type { Event } from "@/entities/event/types";
import type { Notification } from "@/entities/notification/types";
import type { Payment } from "@/entities/payment/types";
import type { Seat, SeatLayout, SeatMapResponse } from "@/entities/seat/types";
import type { Session } from "@/entities/session/types";
import type { User } from "@/entities/user/types";
import { toSafeExternalUrl } from "@/shared/lib/url";

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
  const venue = asRecord(item.venue);
  const externalLinksValue = item.externalLinks ?? item.external_links;

  const metroRaw = item.metroStations ?? item.metro_stations;
  const metroStations =
    Array.isArray(metroRaw) && metroRaw.every((s) => typeof s === "string")
      ? (metroRaw as string[])
      : undefined;

  const tagsRaw = item.tags;
  const tags =
    Array.isArray(tagsRaw) && tagsRaw.every((s) => typeof s === "string")
      ? (tagsRaw as string[])
      : undefined;

  return {
    id: asString(item.id),
    title: asString(item.title),
    description: (item.description as string | null | undefined) ?? undefined,
    longDescription:
      (item.longDescription as string | null | undefined) ??
      (item.long_description as string | null | undefined) ??
      undefined,
    category: (item.category as string | null | undefined) ?? undefined,
    posterUrl:
      toSafeExternalUrl(
        (item.posterUrl as string | null | undefined) ??
          (item.poster_url as string | null | undefined) ??
          undefined
      ) ?? undefined,
    status: asString(item.status, "published"),
    source: asString(item.source, "manual"),
    externalSource:
      (item.externalSource as string | null | undefined) ??
      (item.external_source as string | null | undefined) ??
      undefined,
    sourceUrl:
      toSafeExternalUrl(
        (item.sourceUrl as string | null | undefined) ??
          (item.source_url as string | null | undefined) ??
          undefined
      ) ?? undefined,
    bookingMode: asString(item.bookingMode ?? item.booking_mode, "reserveflow_managed"),
    startsAt: optionalString(item.startsAt ?? item.starts_at),
    endsAt: optionalString(item.endsAt ?? item.ends_at),
    ageRestriction:
      (item.ageRestriction as string | null | undefined) ??
      (item.age_restriction as string | null | undefined) ??
      undefined,
    priceMin:
      item.priceMin === undefined || item.priceMin === null
        ? (item.price_min === undefined || item.price_min === null ? undefined : asNumber(item.price_min))
        : asNumber(item.priceMin),
    priceMax:
      item.priceMax === undefined || item.priceMax === null
        ? (item.price_max === undefined || item.price_max === null ? undefined : asNumber(item.price_max))
        : asNumber(item.priceMax),
    tags,
    ratingCount:
      item.ratingCount === undefined || item.ratingCount === null
        ? (item.rating_count === undefined || item.rating_count === null
            ? undefined
            : asNumber(item.rating_count))
        : asNumber(item.ratingCount),
    isImported: Boolean(
      item.isImported ?? item.is_imported ?? asString(item.source, "manual") !== "manual"
    ),
    venue:
      venue.id || venue.name
        ? {
            id: asString(venue.id),
            name: asString(venue.name),
            address: asString(venue.address),
            city: asString(venue.city),
            latitude:
              venue.latitude === undefined || venue.latitude === null
                ? undefined
                : asNumber(venue.latitude),
            longitude:
              venue.longitude === undefined || venue.longitude === null
                ? undefined
                : asNumber(venue.longitude),
            metroStations,
            venueTypeCode:
              (venue.venueTypeCode as string | null | undefined) ??
              (venue.venue_type_code as string | null | undefined) ??
              undefined,
            venueTypeName:
              (venue.venueTypeName as string | null | undefined) ??
              (venue.venue_type_name as string | null | undefined) ??
              undefined
          }
        : undefined,
    externalLinks: Array.isArray(externalLinksValue)
      ? externalLinksValue.map((linkValue: unknown) => {
          const link = asRecord(linkValue);
          return {
            id: asString(link.id),
            externalSource: asString(link.externalSource ?? link.external_source),
            externalId: asString(link.externalId ?? link.external_id),
            sourceUrl:
              toSafeExternalUrl(
                (link.sourceUrl as string | null | undefined) ??
                  (link.source_url as string | null | undefined) ??
                  undefined
              ) ?? undefined,
            importedAt: optionalString(link.importedAt ?? link.imported_at)
          };
        })
      : undefined,
    sessions: Array.isArray(item.sessions) ? item.sessions.map(normalizeSession) : undefined,
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
    isBookable: Boolean(item.isBookable ?? item.is_bookable),
    externalSource:
      (item.externalSource as string | null | undefined) ??
      (item.external_source as string | null | undefined) ??
      undefined,
    externalId:
      (item.externalId as string | null | undefined) ??
      (item.external_id as string | null | undefined) ??
      undefined,
    sourceUrl:
      toSafeExternalUrl(
        (item.sourceUrl as string | null | undefined) ??
          (item.source_url as string | null | undefined) ??
          undefined
      ) ?? undefined,
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
    layoutKey: optionalString(item.layoutKey ?? item.layout_key),
    row: asString(item.row),
    number: asNumber(item.number),
    label: optionalString(item.label),
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
      name: asString(hall.name, "Зал уточняется")
    },
    provider: asString(item.provider, "internal_grid") as SeatMapResponse["provider"],
    layout: normalizeSeatLayout(item.layout),
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
  return {
    id: asString(item.id ?? item.bookingId ?? item.booking_id),
    bookingId: optionalString(item.bookingId ?? item.booking_id),
    userId: optionalString(item.userId ?? item.user_id),
    sessionId: asString(item.sessionId ?? item.session_id),
    status: asString(item.status, "pending") as Booking["status"],
    expiresAt:
      (item.expiresAt as string | null | undefined) ??
      (item.expires_at as string | null | undefined),
    totalPrice: asNumber(item.totalPrice ?? item.total_price),
    createdAt: optionalString(item.createdAt ?? item.created_at),
    updatedAt: optionalString(item.updatedAt ?? item.updated_at),
    items: Array.isArray(item.items) ? item.items.map(normalizeBookingItem) : undefined,
    event: item.event ? normalizeEvent(item.event) : undefined,
    session: item.session ? normalizeSession(item.session) : undefined
  };
}

export function normalizeHoldSeatResponse(value: unknown): HoldSeatResponse {
  const item = asRecord(value);
  const seats = Array.isArray(item.seats) ? item.seats : item.seat ? [item.seat] : [];
  return {
    bookingId: asString(item.bookingId ?? item.booking_id),
    status: asString(item.status, "pending") as HoldSeatResponse["status"],
    expiresAt: asString(item.expiresAt ?? item.expires_at),
    seats: seats.map((seatValue) => {
      const seat = asRecord(seatValue);
      return {
        seatId: asString(seat.seatId ?? seat.seat_id),
        row: asString(seat.row),
        number: asNumber(seat.number)
      };
    }),
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

function normalizeSeatLayout(value: unknown): SeatLayout | undefined {
  const item = asRecord(value);
  if (!Object.keys(item).length) {
    return undefined;
  }

  const canvas = asRecord(item.canvas);
  const stage = asRecord(item.stage);
  return {
    version: asNumber(item.version, 1),
    canvas: {
      width: asNumber(canvas.width, 960),
      height: asNumber(canvas.height, 640)
    },
    stage:
      Object.keys(stage).length > 0
        ? {
            label: optionalString(stage.label),
            x: asNumber(stage.x),
            y: asNumber(stage.y),
            width: asNumber(stage.width),
            height: asNumber(stage.height)
          }
        : undefined,
    seats: Array.isArray(item.seats)
      ? item.seats.map((seatValue) => {
          const seat = asRecord(seatValue);
          return {
            key: asString(seat.key),
            label: optionalString(seat.label),
            row: asString(seat.row),
            number: asNumber(seat.number),
            x: asNumber(seat.x),
            y: asNumber(seat.y),
            price: asNumber(seat.price),
            category: optionalString(seat.category)
          };
        })
      : [],
    meta: Object.keys(asRecord(item.meta)).length ? asRecord(item.meta) : undefined
  };
}
