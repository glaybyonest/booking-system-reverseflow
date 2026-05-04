import type { ApiErrorPayload } from "@/shared/types/api";

export class AppApiError extends Error {
  code: string;
  status: number;
  details?: unknown;

  constructor({ code, message, status, details }: { code: string; message: string; status: number; details?: unknown }) {
    super(message);
    this.name = "AppApiError";
    this.code = code;
    this.status = status;
    this.details = details;
  }
}

export function isApiErrorPayload(value: unknown): value is ApiErrorPayload {
  return (
    typeof value === "object" &&
    value !== null &&
    "error" in value &&
    typeof (value as ApiErrorPayload).error?.code === "string"
  );
}

export function friendlyApiError(error: unknown) {
  if (!(error instanceof AppApiError)) {
    return "Что-то пошло не так. Попробуйте еще раз.";
  }
  const messages: Record<string, string> = {
    SEAT_NOT_AVAILABLE: "Это место уже недоступно. Выберите другое место.",
    SEAT_ALREADY_HELD: "Это место временно удерживается другим пользователем.",
    BOOKING_EXPIRED: "Бронь истекла. Выберите место заново.",
    BOOKING_NOT_PENDING: "Эту бронь уже нельзя оплатить.",
    IDEMPOTENCY_CONFLICT: "Платеж с таким ключом уже использовался для другой операции.",
    UNAUTHORIZED: "Войдите, чтобы продолжить.",
    FORBIDDEN: "У вас нет доступа к этому действию.",
    VALIDATION_ERROR: "Проверьте введенные данные."
  };
  return messages[error.code] ?? error.message;
}
