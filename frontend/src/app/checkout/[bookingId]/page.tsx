import { CheckoutClient } from "@/features/booking-checkout/checkout-client";
import { Footer } from "@/widgets/footer/footer";
import { Header } from "@/widgets/header/header";

type CheckoutPageProps = {
  params: Promise<{
    bookingId: string;
  }>;
};

export default async function CheckoutPage({ params }: CheckoutPageProps) {
  const { bookingId } = await params;

  return (
    <div className="flex min-h-screen flex-col bg-bg">
      <Header variant="auth" />
      <main className="mx-auto w-full max-w-[1440px] flex-1 px-6 py-8">
        <CheckoutClient bookingId={bookingId} />
      </main>
      <Footer />
    </div>
  );
}
