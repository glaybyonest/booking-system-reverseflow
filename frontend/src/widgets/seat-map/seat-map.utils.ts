import type { Seat } from "@/entities/seat/types";

export function groupSeatsByRow(seats: Seat[]) {
  const groups = new Map<string, Seat[]>();
  for (const seat of seats) {
    groups.set(seat.row, [...(groups.get(seat.row) ?? []), seat]);
  }
  return Array.from(groups.entries()).map(([row, rowSeats]) => ({
    row,
    seats: rowSeats.sort((a, b) => a.number - b.number)
  }));
}
