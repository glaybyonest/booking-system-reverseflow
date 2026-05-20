"use client";

import { CalendarDays, MapPin } from "lucide-react";
import { useRouter } from "next/navigation";
import { useMemo, useState } from "react";

import type { Seat } from "@/entities/seat/types";
import { HoldSeatButton } from "@/features/seat-selection/hold-seat-button";
import {
  useHoldSeat,
  useSeatMap,
  useSession
} from "@/features/seat-selection/seat-selection.hooks";
import { friendlyApiError } from "@/shared/api/errors";
import { formatDateTime, formatMoney, formatTimeRange } from "@/shared/lib/date";
import { routes } from "@/shared/lib/routes";
import { Alert } from "@/shared/ui/alert";
import { Spinner } from "@/shared/ui/spinner";
import { ExactSeatMap } from "@/widgets/seat-map/exact-seat-map";

const MAX_SELECTED_SEATS = 4;

export function SeatSelection({ sessionId }: { sessionId: string }) {
  const router = useRouter();
  const session = useSession(sessionId);
  const seatMap = useSeatMap(sessionId);
  const hold = useHoldSeat();
  const [selectedSeatIds, setSelectedSeatIds] = useState<string[]>([]);
  const [localError, setLocalError] = useState<string | null>(null);

  const selectedSeats = useMemo(() => {
    const map = new Map((seatMap.data?.seats ?? []).map((seat) => [seat.seatId, seat]));
    return selectedSeatIds
      .map((seatId) => map.get(seatId))
      .filter((seat): seat is Seat => {
        if (!seat) return false;
        return seat.status === "available";
      });
  }, [seatMap.data?.seats, selectedSeatIds]);

  const totalPrice = selectedSeats.reduce((sum, seat) => sum + seat.price, 0);

  if (session.isLoading || seatMap.isLoading) {
    return (
      <div className="flex justify-center py-20">
        <Spinner />
      </div>
    );
  }

  const error = session.error ?? seatMap.error;
  if (error) {
    return <Alert variant="error">{friendlyApiError(error)}</Alert>;
  }

  const title = seatMap.data?.event?.title ?? session.data?.event?.title ?? "Мероприятие";
  const hall = seatMap.data?.hall?.name ?? session.data?.hall?.name ?? "Зал уточняется";

  function toggleSeat(seat: Seat) {
    if (seat.status !== "available") {
      setLocalError("Это место уже недоступно. Выберите другое.");
      return;
    }
    setLocalError(null);
    setSelectedSeatIds((current) => {
      if (current.includes(seat.seatId)) {
        return current.filter((seatId) => seatId !== seat.seatId);
      }
      if (current.length >= MAX_SELECTED_SEATS) {
        setLocalError("За один раз можно выбрать не больше 4 билетов.");
        return current;
      }
      return [...current, seat.seatId];
    });
  }

  function submitHold() {
    if (!selectedSeats.length) return;
    setLocalError(null);
    hold.mutate(
      { sessionId, seatIds: selectedSeats.map((seat) => seat.seatId) },
      {
        onSuccess: (result) => router.push(routes.checkout(result.bookingId)),
        onError: (err) => {
          setLocalError(friendlyApiError(err));
          seatMap.refetch();
          setSelectedSeatIds([]);
        }
      }
    );
  }

  const hasVip = selectedSeats.some((s) => s.price > 3000);

  return (
    <div className="grid grid-cols-1 gap-8 lg:grid-cols-12">
      {/* ── Sidebar ── */}
      <aside className="lg:col-span-4">
        <div className="sticky top-24 overflow-hidden rounded-2xl border border-border bg-white shadow-card">
          {/* Header */}
          <div className="border-b border-border px-5 py-4">
            <p className="text-[11px] font-bold uppercase tracking-[0.22em] text-mute-2">
              ВЫБРАННОЕ МЕСТО УДЕРЖИВАЕТСЯ
            </p>
            <h1 className="mt-1 line-clamp-2 text-[15px] font-bold text-ink">{title}</h1>
          </div>

          {/* Session info */}
          <div className="space-y-2 px-5 py-4">
            <div className="flex items-center gap-2 text-xs text-mute">
              <CalendarDays className="h-3.5 w-3.5 shrink-0 text-mute-2" />
              <span>
                {formatDateTime(session.data?.startsAt)}
                {session.data?.endsAt
                  ? ` · ${formatTimeRange(session.data?.startsAt, session.data?.endsAt)}`
                  : ""}
              </span>
            </div>
            <div className="flex items-center gap-2 text-xs text-mute">
              <MapPin className="h-3.5 w-3.5 shrink-0 text-mute-2" />
              <span>{hall}</span>
            </div>
          </div>

          {/* Selected seats */}
          <div className="border-t border-border px-5 py-4">
            <div className="mb-3 flex items-center justify-between">
              <p className="text-[11px] font-bold uppercase tracking-widest text-mute-2">
                Выбранные места
              </p>
              {hasVip && (
                <span className="rounded-full bg-warn-soft px-2 py-0.5 text-[10px] font-black uppercase tracking-wide text-warn-fg">
                  VIP
                </span>
              )}
            </div>

            <div className="min-h-[80px] rounded-xl border border-border bg-bg p-3">
              {selectedSeats.length ? (
                <div className="space-y-2">
                  {selectedSeats.map((seat) => (
                    <div
                      key={seat.seatId}
                      className="flex items-center justify-between rounded-lg bg-white px-3 py-2 shadow-sm"
                    >
                      <span className="text-sm font-semibold text-ink">
                        Ряд {seat.row}, место {seat.number}
                      </span>
                      <span className="text-sm font-bold text-ink">{formatMoney(seat.price)}</span>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="pt-3 text-center text-xs text-mute">
                  Выберите от 1 до {MAX_SELECTED_SEATS} мест на схеме
                </p>
              )}
            </div>

            {selectedSeats.length > 0 && (
              <div className="mt-3 flex items-center justify-between border-t border-border pt-3">
                <span className="text-xs text-mute">
                  Билетов: {selectedSeats.length} из {MAX_SELECTED_SEATS}
                </span>
                <span className="text-xl font-extrabold text-ink">{formatMoney(totalPrice)}</span>
              </div>
            )}
          </div>

          {/* Actions */}
          <div className="border-t border-border px-5 pb-5 pt-4">
            {localError && (
              <Alert variant="error" className="mb-3 text-xs">
                {localError}
              </Alert>
            )}

            <form
              onSubmit={(e) => {
                e.preventDefault();
                submitHold();
              }}
            >
              <HoldSeatButton disabled={!selectedSeats.length} pending={hold.isPending} />
            </form>

            <p className="mt-2.5 text-center text-[11px] text-mute-2">
              Место удерживается 10 минут после подтверждения
            </p>
          </div>
        </div>
      </aside>

      {/* ── Seat map ── */}
      <section className="overflow-hidden rounded-2xl border border-border bg-white p-6 shadow-sm lg:col-span-8 sm:p-10">
        <ExactSeatMap
          layout={seatMap.data?.layout}
          seats={seatMap.data?.seats ?? []}
          selectedSeatIds={selectedSeatIds}
          onToggleSeat={toggleSeat}
        />
      </section>
    </div>
  );
}
