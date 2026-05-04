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

    expect(screen.getByLabelText("Seat A1, available")).toBeEnabled();
    expect(screen.getByLabelText("Seat A2, held")).toBeDisabled();
    expect(screen.getByLabelText("Seat A3, booked")).toBeDisabled();

    await userEvent.click(screen.getByLabelText("Seat A1, available"));
    expect(onSelect).toHaveBeenCalledWith(seats[0]);
  });
});
