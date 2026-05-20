"use client";

import type { Seat, SeatLayout, SeatLayoutSeat } from "@/entities/seat/types";
import { cn } from "@/shared/lib/cn";
import { seatStatusLabels } from "@/shared/lib/labels";

type ExactSeatMapProps = {
  layout?: SeatLayout | null;
  seats: Seat[];
  selectedSeatIds?: string[];
  interactive?: boolean;
  onToggleSeat?: (seat: Seat) => void;
};

export function ExactSeatMap({
  layout,
  seats,
  selectedSeatIds = [],
  interactive = true,
  onToggleSeat
}: ExactSeatMapProps) {
  const normalizedLayout = layout ?? buildFallbackLayout(seats);
  const selected = new Set(selectedSeatIds);
  const seatByKey = new Map<string, Seat>();

  seats.forEach((seat) => {
    const fallbackKey = `${seat.row}-${seat.number}`;
    if (seat.layoutKey) {
      seatByKey.set(seat.layoutKey, seat);
    }
    if (seat.label) {
      seatByKey.set(seat.label, seat);
    }
    seatByKey.set(fallbackKey, seat);
  });

  return (
    <div className="space-y-6">
      <div
        className="relative w-full overflow-hidden rounded-[2rem] border border-gray-100 bg-white"
        style={{ aspectRatio: `${normalizedLayout.canvas.width} / ${normalizedLayout.canvas.height}` }}
      >
        {normalizedLayout.stage ? (
          <div
            className="absolute rounded-full border border-gray-200 bg-gradient-to-b from-gray-100 to-white text-center text-xs font-semibold uppercase tracking-[0.24em] text-gray-500"
            style={rectStyle(normalizedLayout.stage, normalizedLayout.canvas)}
          >
            <div className="flex h-full items-center justify-center px-3">
              {normalizedLayout.stage.label ?? "Сцена"}
            </div>
          </div>
        ) : null}

        {normalizedLayout.seats.map((layoutSeat) => {
          const seat = seatByKey.get(layoutSeat.key) ?? seatByKey.get(fallbackLayoutKey(layoutSeat));
          const isSelected = seat ? selected.has(seat.seatId) : false;
          const isDisabled = !interactive || !seat || seat.status !== "available";

          return (
            <button
              key={layoutSeat.key}
              type="button"
              aria-label={seatAriaLabel(seat, layoutSeat)}
              disabled={isDisabled}
              title={seat?.status === "held" ? "Место временно удерживается" : undefined}
              onClick={() => {
                if (seat && onToggleSeat) {
                  onToggleSeat(seat);
                }
              }}
              className={cn(
                "absolute flex h-10 w-10 -translate-x-1/2 -translate-y-1/2 items-center justify-center rounded-full border text-[11px] font-semibold shadow-sm transition-colors",
                !seat && "cursor-default border-dashed border-gray-200 bg-gray-50 text-gray-400",
                seat?.status === "available" &&
                  !isSelected &&
                  "border-gray-300 bg-white text-gray-700 hover:border-gray-900 hover:text-gray-900",
                seat?.status === "available" &&
                  isSelected &&
                  "border-gray-900 bg-gray-900 text-white",
                seat?.status === "held" && "cursor-not-allowed border-orange-200 bg-orange-50 text-orange-500",
                seat?.status === "booked" && "cursor-not-allowed border-gray-100 bg-gray-100 text-gray-400"
              )}
              style={pointStyle(layoutSeat, normalizedLayout.canvas)}
            >
              {seat?.status === "booked" ? "×" : layoutSeat.number}
            </button>
          );
        })}
      </div>

      <div className="flex flex-wrap justify-center gap-6 border-t border-gray-100 pt-4 text-sm text-gray-500">
        <LegendSeat label="Свободно" className="border border-gray-300 bg-white" />
        <LegendSeat label="Выбрано" className="border border-gray-900 bg-gray-900" />
        <LegendSeat label="Удерживается" className="border border-orange-200 bg-orange-50" />
        <LegendSeat label="Занято" className="border border-gray-100 bg-gray-100" />
      </div>
    </div>
  );
}

function LegendSeat({ label, className }: { label: string; className: string }) {
  return (
    <div className="flex items-center gap-2">
      <div className={cn("h-5 w-5 rounded-full", className)} />
      <span>{label}</span>
    </div>
  );
}

function seatAriaLabel(seat: Seat | undefined, layoutSeat: SeatLayoutSeat) {
  if (!seat) {
    return `Место ${layoutSeat.row}${layoutSeat.number}, схема не привязана`;
  }
  return `Место ${seat.row}${seat.number}, ${seatStatusLabels[seat.status]}`;
}

function fallbackLayoutKey(seat: SeatLayoutSeat) {
  return `${seat.row}-${seat.number}`;
}

function rectStyle(
  rect: { x: number; y: number; width: number; height: number },
  canvas: SeatLayout["canvas"]
) {
  return {
    left: `${(rect.x / canvas.width) * 100}%`,
    top: `${(rect.y / canvas.height) * 100}%`,
    width: `${(rect.width / canvas.width) * 100}%`,
    height: `${(rect.height / canvas.height) * 100}%`
  };
}

function pointStyle(seat: SeatLayoutSeat, canvas: SeatLayout["canvas"]) {
  return {
    left: `${(seat.x / canvas.width) * 100}%`,
    top: `${(seat.y / canvas.height) * 100}%`
  };
}

function buildFallbackLayout(seats: Seat[]): SeatLayout {
  const grouped = new Map<string, Seat[]>();
  seats.forEach((seat) => {
    grouped.set(seat.row, [...(grouped.get(seat.row) ?? []), seat]);
  });

  const rows = Array.from(grouped.entries()).sort(([a], [b]) => a.localeCompare(b));
  const layoutSeats: SeatLayoutSeat[] = [];
  rows.forEach(([row, rowSeats], rowIndex) => {
    rowSeats
      .sort((a, b) => a.number - b.number)
      .forEach((seat, seatIndex) => {
        layoutSeats.push({
          key: seat.layoutKey ?? `${seat.row}-${seat.number}`,
          label: seat.label,
          row,
          number: seat.number,
          x: 120 + seatIndex * 54,
          y: 170 + rowIndex * 58,
          price: seat.price
        });
      });
  });

  const maxSeats = Math.max(...rows.map(([, rowSeats]) => rowSeats.length), 1);
  return {
    version: 1,
    canvas: {
      width: Math.max(720, 180 + maxSeats * 54),
      height: Math.max(520, 250 + rows.length * 58)
    },
    stage: {
      label: "Сцена",
      x: 160,
      y: 40,
      width: Math.max(260, maxSeats * 36),
      height: 56
    },
    seats: layoutSeats
  };
}
