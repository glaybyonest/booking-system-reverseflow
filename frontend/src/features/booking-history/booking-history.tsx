"use client";

import Link from "next/link";
import { useState } from "react";

import { friendlyApiError } from "@/shared/api/errors";
import { formatDateTime, formatMoney } from "@/shared/lib/date";
import { routes } from "@/shared/lib/routes";
import { Alert } from "@/shared/ui/alert";
import { EmptyState } from "@/shared/ui/empty-state";
import { Spinner } from "@/shared/ui/spinner";
import { BookingStatusBadge } from "@/features/booking-history/booking-status-badge";
import { useCancelBooking, useMyBookings } from "@/features/booking-history/booking-history.hooks";
import type { BookingStatus } from "@/entities/booking/types";
import { cn } from "@/shared/lib/cn";

const TABS: { label: string; status?: BookingStatus }[] = [
  { label: "Все" },
  { label: "Ожидают", status: "pending" },
  { label: "Подтверждены", status: "confirmed" },
  { label: "Истекли", status: "expired" },
  { label: "Отменены", status: "cancelled" },
  { label: "Ошибка оплаты", status: "payment_failed" }
];

const COVER_COLORS: Record<string, string> = {
  "0": "bg-[#1A1A2E]",
  "1": "bg-[#5C2E1A]",
  "2": "bg-[#1B3A6B]",
  "3": "bg-[#1A3D2B]",
  "4": "bg-[#2D1B4E]"
};

export function BookingHistory() {
  const bookings = useMyBookings();
  const cancel = useCancelBooking();
  const [activeTab, setActiveTab] = useState<BookingStatus | undefined>(undefined);
  const [cancellingId, setCancellingId] = useState<string | null>(null);

  function handleCancel(bookingId: string) {
    setCancellingId(bookingId);
    cancel.mutate(bookingId, {
      onSettled: () => setCancellingId(null)
    });
  }

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

  const filtered = activeTab
    ? bookings.data.filter((b) => b.status === activeTab)
    : bookings.data;

  const countByStatus = (status: BookingStatus) =>
    bookings.data?.filter((b) => b.status === status).length ?? 0;

  return (
    <div>
      {/* Status tabs */}
      <div className="mb-5 flex items-center gap-1 overflow-x-auto border-b border-border pb-0 scrollbar-none">
        {TABS.map((t) => {
          const count = t.status ? countByStatus(t.status) : bookings.data?.length ?? 0;
          const isActive = activeTab === t.status;
          return (
            <button
              key={t.label}
              onClick={() => setActiveTab(t.status)}
              className={cn(
                "flex shrink-0 items-center gap-1.5 border-b-2 px-4 py-3 text-sm font-medium transition-colors",
                isActive
                  ? "border-ink text-ink"
                  : "border-transparent text-mute hover:text-ink"
              )}
            >
              {t.label}
              {count > 0 && (
                <span
                  className={cn(
                    "rounded-full px-1.5 py-0.5 text-[11px] font-bold",
                    isActive ? "bg-ink text-white" : "bg-bg text-mute-2"
                  )}
                >
                  {count}
                </span>
              )}
            </button>
          );
        })}
      </div>

      {/* Booking rows */}
      <div className="divide-y divide-border overflow-hidden rounded-2xl border border-border bg-white shadow-card">
        {filtered.length === 0 ? (
          <div className="py-12 text-center text-sm text-mute">Броней нет</div>
        ) : (
          filtered.map((booking, idx) => {
            const item = booking.items?.[0];
            const coverColor = COVER_COLORS[String(idx % 5)] ?? COVER_COLORS["0"];
            const seatLabel = item?.row
              ? `Ряд ${item.row}, Место ${item.number}`
              : "место уточняется";

            return (
              <div
                key={booking.id}
                className="flex items-center gap-4 px-5 py-4 hover:bg-bg transition-colors"
              >
                {/* Cover avatar */}
                <div
                  className={cn(
                    "flex h-10 w-10 shrink-0 items-center justify-center rounded-xl text-xs font-black text-white",
                    coverColor
                  )}
                >
                  {booking.id.slice(0, 2).toUpperCase()}
                </div>

                {/* Main info */}
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-bold text-ink">
                    {booking.event?.title ?? booking.session?.event?.title ?? `Бронь ${booking.id.slice(0, 8).toUpperCase()}`}
                  </p>
                  <p className="mt-0.5 truncate text-xs text-mute">
                    {formatDateTime(booking.createdAt)} · {seatLabel}
                  </p>
                </div>

                {/* Booking number */}
                <div className="hidden shrink-0 lg:block">
                  <p className="text-[11px] text-mute-2">Бронь</p>
                  <p className="font-mono text-xs font-medium text-mute">
                    RF-{booking.id.slice(0, 6).toUpperCase()}
                  </p>
                </div>

                {/* Status */}
                <div className="shrink-0">
                  <BookingStatusBadge status={booking.status} />
                </div>

                {/* Price */}
                <div className="shrink-0 text-right">
                  <p className="text-base font-bold text-ink">{formatMoney(booking.totalPrice)}</p>
                  {booking.status === "pending" && (
                    <div className="mt-0.5 flex items-center justify-end gap-2">
                      <Link href={routes.checkout(booking.id)}>
                        <span className="text-[11px] font-medium text-ink hover:underline">
                          Оплатить →
                        </span>
                      </Link>
                      <span className="text-mute-2">·</span>
                      <button
                        onClick={() => handleCancel(booking.id)}
                        disabled={cancellingId === booking.id}
                        className="text-[11px] font-medium text-err transition-colors hover:underline disabled:opacity-50"
                      >
                        {cancellingId === booking.id ? "..." : "Отменить"}
                      </button>
                    </div>
                  )}
                </div>
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}
