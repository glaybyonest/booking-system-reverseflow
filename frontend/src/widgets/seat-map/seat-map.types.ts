import type { Seat } from "@/entities/seat/types";

export type SeatMapProps = {
  seats: Seat[];
  selectedSeatId?: string;
  onSelectSeat(seat: Seat): void;
};
