import { Suspense } from "react";

import { EventList } from "@/features/event-list/event-list";
import { Footer } from "@/widgets/footer/footer";
import { Header } from "@/widgets/header/header";
import { Spinner } from "@/shared/ui/spinner";

export default function EventsPage() {
  return (
    <div className="flex min-h-screen flex-col bg-bg">
      <Header variant="auto" />
      <main className="mx-auto w-full max-w-[1440px] flex-1 px-6 py-8">
        <div className="mb-6">
          <p className="text-[11px] font-bold uppercase tracking-[0.22em] text-mute-2">
            КАТАЛОГ · МЕРОПРИЯТИЯ
          </p>
          <h1 className="mt-1 text-[28px] font-extrabold tracking-tight text-ink">
            Все мероприятия
          </h1>
          <p className="mt-2 max-w-xl text-sm text-mute">
            Полный каталог актуальных событий с прямым переходом к бронированию.
          </p>
        </div>
        <Suspense
          fallback={
            <div className="flex justify-center py-20">
              <Spinner />
            </div>
          }
        >
          <EventList />
        </Suspense>
      </main>
      <Footer />
    </div>
  );
}
