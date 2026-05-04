export type SeatStatus = "available" | "held" | "booked";

export type Seat = {
  seatId: string;
  row: string;
  number: number;
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
  seats: Seat[];
};
