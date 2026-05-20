export type SeatStatus = "available" | "held" | "booked";
export type SeatMapProvider = "internal_grid" | "react_seat_toolkit";

export type SeatLayout = {
  version: number;
  canvas: {
    width: number;
    height: number;
  };
  stage?: {
    label?: string;
    x: number;
    y: number;
    width: number;
    height: number;
  } | null;
  seats: SeatLayoutSeat[];
  meta?: Record<string, unknown> | null;
};

export type SeatLayoutSeat = {
  key: string;
  label?: string | null;
  row: string;
  number: number;
  x: number;
  y: number;
  price: number;
  category?: string | null;
};

export type Seat = {
  seatId: string;
  layoutKey?: string;
  row: string;
  number: number;
  label?: string;
  status: SeatStatus;
  price: number;
};

export type SeatMapResponse = {
  sessionId: string;
  event?: {
    id: string;
    title: string;
  };
  hall: {
    id?: string;
    name: string;
  };
  provider?: SeatMapProvider;
  layout?: SeatLayout | null;
  seats: Seat[];
};
