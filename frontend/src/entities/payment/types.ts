export type PaymentStatus = "pending" | "succeeded" | "failed";

export type Payment = {
  id: string;
  paymentId?: string;
  bookingId: string;
  status: PaymentStatus;
  amount: number;
  provider: string;
};
