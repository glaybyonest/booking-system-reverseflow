import { Seat } from "@/widgets/seat-map/seat";
import type { SeatMapProps } from "@/widgets/seat-map/seat-map.types";
import { groupSeatsByRow } from "@/widgets/seat-map/seat-map.utils";

export function SeatMap({ seats, selectedSeatId, onSelectSeat }: SeatMapProps) {
  const rows = groupSeatsByRow(seats);
  return (
    <>
      <div className="w-full overflow-x-auto pb-6 seat-scroll">
        <div className="flex min-w-[500px] flex-col items-center gap-4">
          {rows.map((row) => (
            <div key={row.row} className="flex items-center gap-3">
              <span className="mr-2 w-6 text-right text-sm font-bold text-gray-400">{row.row}</span>
              {row.seats.map((seat, index) => (
                <div key={seat.seatId} className="flex items-center gap-3">
                  {index === Math.ceil(row.seats.length / 2) ? <div className="w-4" /> : null}
                  <Seat
                    seat={seat}
                    selected={selectedSeatId === seat.seatId}
                    onSelect={() => onSelectSeat(seat)}
                  />
                </div>
              ))}
            </div>
          ))}
        </div>
      </div>
      <div className="mt-auto flex w-full flex-wrap justify-center gap-6 border-t border-gray-100 pt-8 text-sm text-gray-500">
        <LegendSeat label="Свободно" className="border-2 border-gray-200" />
        <LegendSeat label="Выбрано" className="border-2 border-gray-900 bg-gray-900" />
        <LegendSeat label="Удерживается" className="border-2 border-orange-300 bg-orange-50" />
        <LegendSeat label="Занято" className="border-2 border-gray-100 bg-gray-100 opacity-50" />
      </div>
    </>
  );
}

function LegendSeat({ label, className }: { label: string; className: string }) {
  return (
    <div className="flex items-center gap-2">
      <div className={`h-5 w-5 rounded-b-sm rounded-t-md ${className}`} />
      {label}
    </div>
  );
}
