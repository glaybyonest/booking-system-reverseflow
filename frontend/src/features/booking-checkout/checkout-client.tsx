"use client";

import Link from "next/link";
import { useEffect, useState } from "react";

import { friendlyApiError } from "@/shared/api/errors";
import { getPaymentIdempotencyKey } from "@/shared/lib/idempotency";
import { routes } from "@/shared/lib/routes";
import { Alert } from "@/shared/ui/alert";
import { Button } from "@/shared/ui/button";
import { Card, CardContent } from "@/shared/ui/card";
import { Spinner } from "@/shared/ui/spinner";
import { BookingSummary } from "@/widgets/booking-summary/booking-summary";
import { BookingStatusBadge } from "@/features/booking-history/booking-status-badge";
import { HoldTimer } from "@/features/booking-checkout/hold-timer";
import { PaymentButton } from "@/features/booking-checkout/payment-button";
import { useBooking, useCreatePayment } from "@/features/booking-checkout/booking-checkout.hooks";

export function CheckoutClient({ bookingId }: { bookingId: string }) {
  const booking = useBooking(bookingId);
  const payment = useCreatePayment();
  const [now, setNow] = useState(0);

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

  const remaining = booking.data.expiresAt ? new Date(booking.data.expiresAt).getTime() - now : 0;
  const canPay = booking.data.status === "pending" && remaining > 0;

  function pay(forceStatus: "succeeded" | "failed") {
    payment.mutate(
      {
        bookingId,
        idempotencyKey: getPaymentIdempotencyKey(bookingId),
        forceStatus
      },
      {
        onSuccess: () => booking.refetch()
      }
    );
  }

  return (
    <div className="mx-auto grid max-w-5xl grid-cols-1 gap-8 lg:grid-cols-12">
      <div className="lg:col-span-7">
        <BookingSummary booking={booking.data} />
      </div>
      <Card className="lg:col-span-5">
        <CardContent>
          <div className="flex items-center justify-between gap-3">
            <h1 className="text-2xl font-bold">Checkout</h1>
            <BookingStatusBadge status={booking.data.status} />
          </div>

          {booking.data.status === "pending" ? (
            <div className="mt-6 space-y-4">
              <HoldTimer expiresAt={booking.data.expiresAt} onExpired={() => booking.refetch()} />
              {payment.error ? <Alert variant="error">{friendlyApiError(payment.error)}</Alert> : null}
              <PaymentButton pending={payment.isPending} disabled={!canPay} onClick={pay} />
              <Button
                type="button"
                variant="secondary"
                className="w-full"
                disabled={!canPay || payment.isPending}
                onClick={() => pay("failed")}
              >
                Симулировать ошибку оплаты
              </Button>
              {!canPay ? <Alert variant="info">Время оплаты истекло. Обновляем статус брони.</Alert> : null}
            </div>
          ) : null}

          {booking.data.status === "confirmed" ? (
            <div className="mt-6 space-y-4">
              <Alert variant="success">Бронь подтверждена. Место закреплено за вами.</Alert>
              <Link href={routes.bookings}>
                <Button className="w-full">Перейти в мои брони</Button>
              </Link>
            </div>
          ) : null}

          {booking.data.status === "expired" ? (
            <div className="mt-6 space-y-4">
              <Alert variant="info">Бронь истекла, место снова доступно другим пользователям.</Alert>
              <Link href={routes.events}>
                <Button variant="secondary" className="w-full">
                  Вернуться к мероприятиям
                </Button>
              </Link>
            </div>
          ) : null}

          {booking.data.status === "payment_failed" ? (
            <div className="mt-6 space-y-4">
              <Alert variant="error">Оплата не прошла, место было освобождено.</Alert>
              <Link href={routes.events}>
                <Button variant="secondary" className="w-full">
                  Выбрать другое место
                </Button>
              </Link>
            </div>
          ) : null}
        </CardContent>
      </Card>
    </div>
  );
}
