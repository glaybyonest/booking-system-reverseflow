import { X } from "lucide-react";

import type { Seat as SeatModel } from "@/entities/seat/types";
import { cn } from "@/shared/lib/cn";

export function Seat({
  seat,
  selected,
  onSelect
}: {
  seat: SeatModel;
  selected: boolean;
  onSelect(): void;
}) {
  const disabled = seat.status !== "available";
  return (
    <button
      type="button"
      aria-label={`Seat ${seat.row}${seat.number}, ${seat.status}`}
      disabled={disabled}
      title={seat.status === "held" ? "Временно удерживается" : undefined}
      onClick={onSelect}
      className={cn(
        "flex h-10 w-10 items-center justify-center rounded-b-md rounded-t-xl border-2 text-sm font-medium transition-colors",
        seat.status === "available" &&
          !selected &&
          "border-gray-200 hover:border-gray-900 hover:bg-gray-50",
        selected && "border-gray-900 bg-gray-900 text-white shadow-md",
        seat.status === "held" && "cursor-not-allowed border-orange-300 bg-orange-50",
        seat.status === "booked" &&
          "cursor-not-allowed border-gray-100 bg-gray-100 text-gray-400 opacity-50"
      )}
    >
      {seat.status === "booked" ? <X className="h-4 w-4" /> : selected ? seat.number : null}
    </button>
  );
}
