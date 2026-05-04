export function getPaymentIdempotencyKey(bookingId: string) {
  const storageKey = `payment-idempotency:${bookingId}`;
  if (typeof window === "undefined") {
    return `${bookingId}:server`;
  }
  const existing = window.sessionStorage.getItem(storageKey);
  if (existing) return existing;
  const key =
    typeof window.crypto?.randomUUID === "function"
      ? window.crypto.randomUUID()
      : `${bookingId}:${Date.now()}:${Math.random().toString(16).slice(2)}`;
  window.sessionStorage.setItem(storageKey, key);
  return key;
}
