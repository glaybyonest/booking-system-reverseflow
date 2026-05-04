import type { Payment } from "@/entities/payment/types";
import { clientFetch } from "@/shared/api/client";
import { normalizePayment } from "@/shared/api/mappers";

export async function createPayment(input: {
  bookingId: string;
  idempotencyKey: string;
  forceStatus?: "succeeded" | "failed";
}): Promise<Payment> {
  return normalizePayment(
    await clientFetch("payments", {
      method: "POST",
      json: input
    })
  );
}

export async function getPayment(paymentId: string): Promise<Payment> {
  return normalizePayment(await clientFetch(`payments/${paymentId}`));
}
