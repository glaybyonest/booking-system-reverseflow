import { BookingHistory } from "@/features/booking-history/booking-history";
import { Footer } from "@/widgets/footer/footer";
import { Header } from "@/widgets/header/header";

export default function BookingsPage() {
  return (
    <div className="flex min-h-screen flex-col bg-bg">
      <Header variant="auth" />
      <main className="mx-auto w-full max-w-[1440px] flex-1 px-6 py-8">
        <div className="mb-6">
          <p className="text-[11px] font-bold uppercase tracking-[0.22em] text-mute-2">
            ЛИЧНЫЙ КАБИНЕТ
          </p>
          <div className="mt-1 flex items-center justify-between">
            <h1 className="text-[28px] font-extrabold tracking-tight text-ink">Мои брони</h1>
            <div className="flex gap-2">
              <button className="rounded-full border border-border bg-white px-4 py-2 text-[13px] font-medium text-ink-2 hover:border-ink/30 hover:text-ink">
                Фильтры
              </button>
              <button className="rounded-full border border-border bg-white px-4 py-2 text-[13px] font-medium text-ink-2 hover:border-ink/30 hover:text-ink">
                Период
              </button>
            </div>
          </div>
        </div>
        <BookingHistory />
      </main>
      <Footer />
    </div>
  );
}
