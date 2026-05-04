import { CheckoutClient } from "@/features/booking-checkout/checkout-client";
import { Header } from "@/widgets/header/header";

export default function CheckoutPage({ params }: { params: { bookingId: string } }) {
  return (
    <div className="min-h-screen bg-[#F8F9FA]">
      <Header variant="auth" />
      <main className="mx-auto max-w-7xl px-4 py-8 sm:px-6">
        <CheckoutClient bookingId={params.bookingId} />
      </main>
    </div>
  );
}
