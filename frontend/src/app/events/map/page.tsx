import { Suspense } from "react";

import { EventsMapPage } from "@/features/events-map/events-map-page";
import { Footer } from "@/widgets/footer/footer";
import { Header } from "@/widgets/header/header";
import { Spinner } from "@/shared/ui/spinner";

export default function EventsMapRoute() {
  return (
    <div className="flex min-h-screen flex-col bg-bg">
      <Header variant="auto" />
      <main className="mx-auto w-full max-w-[1440px] flex-1 px-6 py-8">
        <div className="mb-6">
          <p className="text-[11px] font-bold uppercase tracking-[0.22em] text-mute-2">
            АФИША · КАРТА
          </p>
          <h1 className="mt-1 text-[28px] font-extrabold tracking-tight text-ink">
            Карта мероприятий
          </h1>
          <p className="mt-2 text-sm text-mute">
            Смотрите все актуальные события на карте и переходите к подробностям.
          </p>
        </div>
        <Suspense
          fallback={
            <div className="flex justify-center py-20">
              <Spinner />
            </div>
          }
        >
          <EventsMapPage />
        </Suspense>
      </main>
      <Footer />
    </div>
  );
}
