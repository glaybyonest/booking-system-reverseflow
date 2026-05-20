import type { Booking } from "@/entities/booking/types";
import { BookingStatusBadge } from "@/features/booking-history/booking-status-badge";
import { formatDateTime, formatMoney } from "@/shared/lib/date";
import { Card, CardContent } from "@/shared/ui/card";

export function BookingSummary({ booking }: { booking: Booking }) {
  return (
    <Card>
      <CardContent>
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <p className="text-xs font-bold uppercase tracking-wider text-gray-400">Бронь</p>
            <h2 className="mt-1 break-all text-xl font-bold">{booking.id}</h2>
          </div>
          <BookingStatusBadge status={booking.status} />
        </div>
        <div className="mt-6 grid gap-3 text-sm text-gray-500">
          <p>Сеанс: {booking.sessionId}</p>
          <div className="space-y-1">
            <p>Места:</p>
            {(booking.items ?? []).length ? (
              <div className="space-y-1">
                {(booking.items ?? []).map((item) => (
                  <p key={item.id}>{item.row ? `Ряд ${item.row}, место ${item.number}` : item.seatId}</p>
                ))}
              </div>
            ) : (
              <p>Уточняется</p>
            )}
          </div>
          {booking.expiresAt ? <p>Удержание до: {formatDateTime(booking.expiresAt)}</p> : null}
        </div>
        <div className="mt-6 rounded-2xl border border-gray-200 bg-[#F8F9FA] p-4">
          <div className="flex items-center justify-between">
            <span className="text-sm text-gray-500">Итого</span>
            <span className="text-2xl font-bold">{formatMoney(booking.totalPrice)}</span>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
