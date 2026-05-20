import type { SeatLayout } from "@/entities/seat/types";

export type LayoutVenueRef = {
  id: string;
  name: string;
};

export type LayoutHallRef = {
  id: string;
  name: string;
  venue: LayoutVenueRef;
};

export type SessionLayoutState = {
  sessionId: string;
  eventId: string;
  eventTitle: string;
  source: string;
  bookingMode: string;
  isBookable: boolean;
  hall?: LayoutHallRef | null;
  layout?: SeatLayout | null;
  fallbackLayout?: SeatLayout | null;
  effectiveLayout?: SeatLayout | null;
  layoutSource: "none" | "hall" | "session" | string;
};

export type HallLayoutSessionSummary = {
  id: string;
  eventId: string;
  eventTitle: string;
  startsAt?: string | null;
  isBookable: boolean;
};

export type HallLayoutState = {
  hallId: string;
  name: string;
  venue: LayoutVenueRef;
  layout?: SeatLayout | null;
  sessions: HallLayoutSessionSummary[];
};

export type LayoutMutationResult = {
  sessionId?: string | null;
  hallId: string;
  eventId: string;
  eventTitle: string;
  bookingMode: string;
  isBookable: boolean;
  visibleInCatalog: boolean;
  materializedSeats: number;
};
