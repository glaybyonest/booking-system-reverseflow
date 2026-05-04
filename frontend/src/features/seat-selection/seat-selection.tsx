"use client";

import { CalendarDays, MapPin } from "lucide-react";
import { useRouter } from "next/navigation";
import { useMemo, useState } from "react";

import type { Seat } from "@/entities/seat/types";
import { friendlyApiError } from "@/shared/api/errors";
import { formatDateTime, formatMoney, formatTimeRange } from "@/shared/lib/date";
import { routes } from "@/shared/lib/routes";
import { Alert } from "@/shared/ui/alert";
import { Badge } from "@/shared/ui/badge";
import { Card } from "@/shared/ui/card";
import { Spinner } from "@/shared/ui/spinner";
import { SeatMap } from "@/widgets/seat-map/seat-map";
import { HoldSeatButton } from "@/features/seat-selection/hold-seat-button";
import { useHoldSeat, useSeatMap, useSession } from "@/features/seat-selection/seat-selection.hooks";

export function SeatSelection({ sessionId }: { sessionId: string }) {
  const router = useRouter();
  const session = useSession(sessionId);
  const seatMap = useSeatMap(sessionId);
  const hold = useHoldSeat();
  const [selectedSeatId, setSelectedSeatId] = useState<string | undefined>();
  const [localError, setLocalError] = useState<string | null>(null);

  const rawSelectedSeat = useMemo(
    () => seatMap.data?.seats.find((seat) => seat.seatId === selectedSeatId),
    [seatMap.data?.seats, selectedSeatId]
  );
  const selectedSeat = rawSelectedSeat?.status === "available" ? rawSelectedSeat : undefined;

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

  function submitHold() {
    if (!selectedSeat) return;
    setLocalError(null);
    hold.mutate(
      { sessionId, seatId: selectedSeat.seatId },
      {
        onSuccess: (result) => router.push(routes.checkout(result.bookingId)),
        onError: (err) => {
          setLocalError(friendlyApiError(err));
          seatMap.refetch();
          setSelectedSeatId(undefined);
        }
      }
    );
  }

  return (
    <div className="grid grid-cols-1 gap-8 lg:grid-cols-12">
      <aside className="lg:col-span-4">
        <Card className="sticky top-24 p-6">
          <Badge className="mb-4">Сеанс</Badge>
          <h1 className="mb-2 text-2xl font-bold">{title}</h1>
          <div className="mb-8 space-y-2 text-sm text-gray-500">
            <p className="flex gap-2">
              <CalendarDays className="h-4 w-4 shrink-0" />
              <span>
                {formatDateTime(session.data?.startsAt)} ·{" "}
                {formatTimeRange(session.data?.startsAt, session.data?.endsAt)}
              </span>
            </p>
            <p className="flex gap-2">
              <MapPin className="h-4 w-4 shrink-0" />
              <span>{hall}</span>
            </p>
          </div>

          <div className="border-t border-gray-100 pt-6">
            <h3 className="mb-4 text-xs font-bold uppercase tracking-wider text-gray-400">
              Выбранное место
            </h3>
            <div className="mb-6 flex items-center justify-between rounded-2xl border border-gray-200 bg-[#F8F9FA] p-4">
              {selectedSeat ? (
                <>
                  <div>
                    <div className="text-lg font-bold">
                      Ряд {selectedSeat.row}, место {selectedSeat.number}
                    </div>
                    <div className="mt-0.5 text-xs text-gray-500">Стандартный билет</div>
                  </div>
                  <div className="text-xl font-bold">{formatMoney(selectedSeat.price)}</div>
                </>
              ) : (
                <p className="text-sm text-gray-500">Выберите свободное место на схеме</p>
              )}
            </div>
            {localError ? (
              <Alert variant="error" className="mb-4">
                {localError}
              </Alert>
            ) : null}
            <form
              onSubmit={(event) => {
                event.preventDefault();
                submitHold();
              }}
            >
              <HoldSeatButton disabled={!selectedSeat} pending={hold.isPending} />
            </form>
            <p className="mt-3 text-center text-xs text-gray-400">
              После нажатия место будет удержано на 10 минут
            </p>
          </div>
        </Card>
      </aside>

      <section className="flex flex-col items-center overflow-hidden rounded-[2rem] border border-gray-100 bg-white p-6 shadow-sm sm:p-10 lg:col-span-8">
        <div className="mb-16 flex h-12 w-full max-w-sm items-start justify-center rounded-t-[50%] border-t-4 border-gray-200 bg-gradient-to-b from-gray-100 to-transparent pt-2 text-xs font-bold uppercase tracking-widest text-gray-400">
          Сцена
        </div>
        <SeatMap
          seats={seatMap.data?.seats ?? []}
          selectedSeatId={selectedSeat?.seatId}
          onSelectSeat={(seat: Seat) => {
            if (seat.status === "available") {
              setSelectedSeatId(seat.seatId);
              setLocalError(null);
            }
          }}
        />
      </section>
    </div>
  );
}
