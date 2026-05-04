import type { BookingStatus } from "@/entities/booking/types";
import { Badge } from "@/shared/ui/badge";

const labels: Record<BookingStatus, string> = {
  pending: "Ожидает оплаты",
  confirmed: "Подтверждена",
  cancelled: "Отменена",
  expired: "Истекла",
  payment_failed: "Ошибка оплаты"
};

export function BookingStatusBadge({ status }: { status: BookingStatus }) {
  const variant =
    status === "confirmed"
      ? "dark"
      : status === "payment_failed"
        ? "danger"
        : status === "expired"
          ? "muted"
          : "default";
  return <Badge variant={variant}>{labels[status] ?? status}</Badge>;
}
