import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import React from "react";
import { describe, expect, it, vi } from "vitest";

import { SeatMap } from "@/widgets/seat-map/seat-map";
import type { Seat } from "@/entities/seat/types";

const seats: Seat[] = [
  { seatId: "a1", row: "A", number: 1, status: "available", price: 500 },
  { seatId: "a2", row: "A", number: 2, status: "held", price: 500 },
  { seatId: "a3", row: "A", number: 3, status: "booked", price: 500 }
];

describe("SeatMap", () => {
  it("renders seats and disables held/booked seats", async () => {
    const onSelect = vi.fn();
    render(<SeatMap seats={seats} onSelectSeat={onSelect} />);

    expect(screen.getByLabelText("Место A1, свободно")).toBeEnabled();
    expect(screen.getByLabelText("Место A2, удерживается")).toBeDisabled();
    expect(screen.getByLabelText("Место A3, занято")).toBeDisabled();

    await userEvent.click(screen.getByLabelText("Место A1, свободно"));
    expect(onSelect).toHaveBeenCalledWith(seats[0]);
  });
});
