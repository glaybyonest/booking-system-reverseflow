"use client";

import { AlertTriangle, ArrowRight, CreditCard } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";

import { friendlyApiError } from "@/shared/api/errors";
import { formatMoney } from "@/shared/lib/date";
import { getPaymentIdempotencyKey } from "@/shared/lib/idempotency";
import { routes } from "@/shared/lib/routes";
import { Alert } from "@/shared/ui/alert";
import { Button } from "@/shared/ui/button";
import { Spinner } from "@/shared/ui/spinner";
import { HoldTimer } from "@/features/booking-checkout/hold-timer";
import { useBooking, useCreatePayment } from "@/features/booking-checkout/booking-checkout.hooks";
import { useCancelBooking } from "@/features/booking-history/booking-history.hooks";
import { cn } from "@/shared/lib/cn";

const PAYMENT_TABS = [
  { id: "card", label: "Карта" },
  { id: "sbp", label: "СБП" },
  { id: "wallet", label: "Кошелёк" }
] as const;

type PaymentTab = (typeof PAYMENT_TABS)[number]["id"];

export function CheckoutClient({ bookingId }: { bookingId: string }) {
  const booking = useBooking(bookingId);
  const payment = useCreatePayment();
  const cancel = useCancelBooking();
  const router = useRouter();
  const [now, setNow] = useState(0);
  const [tab, setTab] = useState<PaymentTab>("card");
  const [demoError, setDemoError] = useState(false);

  useEffect(() => {
    const id = window.setInterval(() => setNow(Date.now()), 1000);
    return () => window.clearInterval(id);
  }, []);

  if (booking.isLoading) {
    return (
      <div className="flex justify-center py-20">
        <Spinner />
      </div>
    );
  }
  if (booking.error) {
    return <Alert variant="error">{friendlyApiError(booking.error)}</Alert>;
  }
  if (!booking.data) return null;

  const remaining = booking.data.expiresAt
    ? new Date(booking.data.expiresAt).getTime() - now
    : 0;
  const canPay = booking.data.status === "pending" && remaining > 0;

  function pay() {
    payment.mutate(
      {
        bookingId,
        idempotencyKey: getPaymentIdempotencyKey(bookingId),
        forceStatus: demoError ? "failed" : "succeeded"
      },
      { onSuccess: () => booking.refetch() }
    );
  }

  const item = booking.data.items?.[0];
  const seatLabel = item?.row
    ? `Ряд ${item.row}, место ${item.number}`
    : "место уточняется";

  /* ── Confirmed state ── */
  if (booking.data.status === "confirmed") {
    return (
      <div className="mx-auto max-w-[560px] py-12 text-center">
        <div className="mx-auto mb-4 flex h-20 w-20 items-center justify-center rounded-full border-4 border-ok bg-ok-soft">
          <svg className="h-9 w-9 text-ok" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
          </svg>
        </div>
        <span className="inline-flex items-center gap-1.5 rounded-full bg-ok-soft px-3 py-1 text-xs font-bold uppercase tracking-widest text-ok-fg">
          ✓ CONFIRMED
        </span>
        <h1 className="mt-4 text-3xl font-extrabold text-ink">Бронь подтверждена</h1>
        <p className="mt-2 text-sm text-mute">
          Оплата прошла успешно. Подробности доступны в разделе «Мои брони».
        </p>

        <div className="mt-8 rounded-2xl border border-border bg-white p-6 text-left shadow-card">
          <div className="flex items-start justify-between">
            <div>
              <p className="text-[11px] font-bold uppercase tracking-widest text-mute-2">Номер брони</p>
              <p className="mt-1 font-mono text-lg font-bold text-ink">
                RF-{bookingId.slice(0, 7).toUpperCase()}
              </p>
            </div>
            <div className="flex h-16 w-16 items-center justify-center rounded-xl border border-border bg-bg text-[10px] text-mute-2">
              QR
            </div>
          </div>
          <div className="mt-6 grid grid-cols-2 gap-x-6 gap-y-4 border-t border-border pt-6 text-sm">
            {[
              { label: "Место", value: seatLabel },
              { label: "Итого", value: formatMoney(booking.data.totalPrice) },
              { label: "Статус", value: "✓ CONFIRMED" }
            ].map(({ label, value }) => (
              <div key={label}>
                <p className="text-[11px] font-medium uppercase tracking-wide text-mute-2">{label}</p>
                <p className="mt-0.5 font-semibold text-ink">{value}</p>
              </div>
            ))}
          </div>
        </div>

        <div className="mt-6 flex flex-col gap-3 sm:flex-row sm:justify-center">
          <Link href={routes.bookings}>
            <Button className="w-full sm:w-auto">Мои брони</Button>
          </Link>
          <Link href={routes.events}>
            <Button variant="ghost" className="w-full sm:w-auto">К мероприятиям</Button>
          </Link>
        </div>
      </div>
    );
  }

  /* ── Expired / failed state ── */
  if (booking.data.status === "expired" || booking.data.status === "payment_failed") {
    return (
      <div className="mx-auto max-w-[480px] py-12 text-center">
        <h1 className="text-2xl font-bold text-ink">
          {booking.data.status === "expired" ? "Бронь истекла" : "Оплата не прошла"}
        </h1>
        <p className="mt-2 text-sm text-mute">
          {booking.data.status === "expired"
            ? "Место снова доступно другим пользователям."
            : "Попробуйте снова или выберите другой способ оплаты."}
        </p>
        <Link href={routes.events} className="mt-6 inline-block">
          <Button variant="ghost">Вернуться к мероприятиям</Button>
        </Link>
      </div>
    );
  }

  /* ── Pending payment state ── */
  return (
    <div>
      <nav className="mb-1 text-xs text-mute">
        <span>Бронь</span>
        <span className="mx-1.5">·</span>
        <span className="font-mono font-medium text-ink">
          RF-{bookingId.slice(0, 7).toUpperCase()}
        </span>
      </nav>

      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-2xl font-extrabold text-ink">Подтверждение брони</h1>
        {booking.data.expiresAt && (
          <div className="flex items-center gap-2">
            <span className="text-xs text-mute">Время до истечения</span>
            <HoldTimer expiresAt={booking.data.expiresAt} onExpired={() => booking.refetch()} />
          </div>
        )}
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-12">
        {/* Payment form */}
        <div className="lg:col-span-7">
          {/* Test payment notice */}
          <div className="mb-5 flex items-start gap-3 rounded-xl border border-warn/30 bg-warn-soft px-4 py-3">
            <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0 text-warn" />
            <div>
              <p className="text-xs font-bold text-warn-fg">Тестовая оплата (mock)</p>
              <p className="mt-0.5 text-xs text-warn-fg/80">
                Это mock-оплата для демонстрации MVP. Реальные карты не списываются.
              </p>
            </div>
          </div>

          {/* Payment tabs */}
          <div className="mb-5 flex gap-1 rounded-xl border border-border bg-bg p-1">
            {PAYMENT_TABS.map((t) => (
              <button
                key={t.id}
                onClick={() => setTab(t.id)}
                className={cn(
                  "flex-1 rounded-lg py-2 text-sm font-medium transition-colors",
                  tab === t.id
                    ? "bg-white text-ink shadow-sm"
                    : "text-mute hover:text-ink"
                )}
              >
                {t.label}
              </button>
            ))}
          </div>

          {tab === "card" && (
            <div className="space-y-3">
              <div>
                <label className="mb-1.5 block text-sm font-medium text-ink">Номер карты</label>
                <div className="relative">
                  <CreditCard className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-mute-2" />
                  <input
                    className="w-full rounded-xl border border-border bg-white py-3 pl-10 pr-4 text-sm font-mono text-ink placeholder:text-mute-2 focus:border-ink focus:outline-none focus:ring-2 focus:ring-ink/8"
                    placeholder="4242 4242 4242 4242"
                  />
                </div>
              </div>
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="mb-1.5 block text-sm font-medium text-ink">Срок</label>
                  <input
                    className="w-full rounded-xl border border-border bg-white px-4 py-3 text-sm font-mono text-ink placeholder:text-mute-2 focus:border-ink focus:outline-none focus:ring-2 focus:ring-ink/8"
                    placeholder="MM / YY"
                  />
                </div>
                <div>
                  <label className="mb-1.5 block text-sm font-medium text-ink">CVV</label>
                  <input
                    className="w-full rounded-xl border border-border bg-white px-4 py-3 text-sm font-mono text-ink placeholder:text-mute-2 focus:border-ink focus:outline-none focus:ring-2 focus:ring-ink/8"
                    placeholder="•••"
                  />
                </div>
              </div>
              <div>
                <label className="mb-1.5 block text-sm font-medium text-ink">Имя владельца</label>
                <input
                  className="w-full rounded-xl border border-border bg-white px-4 py-3 text-sm text-ink placeholder:text-mute-2 focus:border-ink focus:outline-none focus:ring-2 focus:ring-ink/8"
                  placeholder="IVAN IVANOV"
                />
              </div>

              <label className="flex cursor-pointer items-center gap-2 text-sm text-mute">
                <input
                  type="checkbox"
                  checked={demoError}
                  onChange={(e) => setDemoError(e.target.checked)}
                  className="rounded border-border accent-ink"
                />
                Demo-режим: вернуть ошибку оплаты
              </label>
            </div>
          )}

          {tab !== "card" && (
            <div className="flex h-32 items-center justify-center rounded-xl border border-border bg-bg text-sm text-mute">
              Способ оплаты недоступен в MVP
            </div>
          )}

          {payment.error ? (
            <Alert variant="error" className="mt-4">
              {friendlyApiError(payment.error)}
            </Alert>
          ) : null}
        </div>

        {/* Booking summary */}
        <div className="lg:col-span-5">
          <div className="rounded-2xl border border-border bg-white p-5 shadow-card">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-ink text-sm font-bold text-white">
                {(booking.data.sessionId ?? "?")[0]?.toUpperCase()}
              </div>
              <div>
                <p className="text-sm font-bold text-ink">Мероприятие</p>
                <p className="text-xs text-mute">
                  Сеанс {booking.data.sessionId?.slice(0, 8)}
                </p>
              </div>
            </div>

            <div className="mt-4 space-y-2 border-t border-border pt-4 text-sm">
              <div className="flex justify-between">
                <span className="text-mute">Место</span>
                <span className="font-medium text-ink">{seatLabel}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-mute">Билет</span>
                <span className="font-medium text-ink">{formatMoney(booking.data.totalPrice)}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-mute">Сервисный сбор</span>
                <span className="font-medium text-ink">0 ₽</span>
              </div>
              <div className="flex justify-between border-t border-border pt-2">
                <span className="font-bold text-ink">Итого</span>
                <span className="text-lg font-extrabold text-ink">
                  {formatMoney(booking.data.totalPrice)}
                </span>
              </div>
            </div>

            {canPay && (
              <>
                <button
                  onClick={pay}
                  disabled={payment.isPending}
                  className="mt-5 flex w-full items-center justify-center gap-2 rounded-full bg-ink py-3.5 text-sm font-bold text-white transition-colors hover:bg-ink-2 disabled:opacity-50"
                >
                  {payment.isPending ? (
                    "Оплачиваем..."
                  ) : (
                    <>
                      Оплатить {formatMoney(booking.data.totalPrice)}
                      <ArrowRight className="h-4 w-4" />
                    </>
                  )}
                </button>
                <button
                  onClick={() =>
                    cancel.mutate(bookingId, {
                      onSuccess: () => router.push(routes.events)
                    })
                  }
                  disabled={cancel.isPending}
                  className="mt-3 w-full text-center text-sm text-err transition-colors hover:underline disabled:opacity-50"
                >
                  {cancel.isPending ? "Отменяем..." : "Отменить бронь"}
                </button>
              </>
            )}

            {!canPay && booking.data.status === "pending" && (
              <Alert variant="warning" className="mt-4">
                Время оплаты истекло. Обновляем статус брони.
              </Alert>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
