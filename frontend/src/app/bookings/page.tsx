import { BookingHistory } from "@/features/booking-history/booking-history";
import { Header } from "@/widgets/header/header";

export default function BookingsPage() {
  return (
    <div className="min-h-screen bg-[#F8F9FA]">
      <Header variant="auth" />
      <main className="mx-auto max-w-7xl px-4 py-8 sm:px-6">
        <div className="mb-8">
          <h1 className="text-4xl font-extrabold tracking-tight">Мои брони</h1>
          <p className="mt-3 text-gray-500">История удержаний, оплат и подтвержденных мест.</p>
        </div>
        <BookingHistory />
      </main>
    </div>
  );
}
