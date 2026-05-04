"use client";

import Link from "next/link";

import { friendlyApiError } from "@/shared/api/errors";
import { formatDateTime, formatMoney } from "@/shared/lib/date";
import { routes } from "@/shared/lib/routes";
import { Alert } from "@/shared/ui/alert";
import { Button } from "@/shared/ui/button";
import { Card, CardContent } from "@/shared/ui/card";
import { EmptyState } from "@/shared/ui/empty-state";
import { Spinner } from "@/shared/ui/spinner";
import { BookingStatusBadge } from "@/features/booking-history/booking-status-badge";
import { useCancelBooking, useMyBookings } from "@/features/booking-history/booking-history.hooks";

export function BookingHistory() {
  const bookings = useMyBookings();
  const cancel = useCancelBooking();

  if (bookings.isLoading) return <Spinner />;
  if (bookings.error) return <Alert variant="error">{friendlyApiError(bookings.error)}</Alert>;
  if (!bookings.data?.length) {
    return (
      <EmptyState
        title="У вас пока нет броней"
        description="Откройте мероприятия и выберите свободное место."
      />
    );
  }

  return (
    <div className="space-y-4">
      {bookings.data.map((booking) => {
        const item = booking.items?.[0];
        return (
          <Card key={booking.id}>
            <CardContent className="flex flex-col gap-5 lg:flex-row lg:items-center lg:justify-between">
              <div className="min-w-0">
                <BookingStatusBadge status={booking.status} />
                <h3 className="mt-3 break-all text-xl font-bold">Бронь {booking.id}</h3>
                <p className="mt-2 text-sm text-gray-500">
                  Сеанс {booking.sessionId} ·{" "}
                  {item?.row ? `Ряд ${item.row}, место ${item.number}` : item?.seatId ?? "место уточняется"}
                </p>
                <p className="mt-1 text-xs text-gray-400">
                  Создана {booking.createdAt ? formatDateTime(booking.createdAt) : "только что"}
                </p>
              </div>
              <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
                <div className="text-2xl font-bold">{formatMoney(booking.totalPrice)}</div>
                {booking.status === "pending" ? (
                  <>
                    <Link href={routes.checkout(booking.id)}>
                      <Button variant="secondary">К оплате</Button>
                    </Link>
                    <Button
                      variant="danger"
                      disabled={cancel.isPending}
                      onClick={() => cancel.mutate(booking.id)}
                    >
                      Отменить
                    </Button>
                  </>
                ) : null}
              </div>
            </CardContent>
          </Card>
        );
      })}
    </div>
  );
}
