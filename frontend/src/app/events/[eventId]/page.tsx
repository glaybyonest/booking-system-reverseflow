import { ChevronRight } from "lucide-react";
import Link from "next/link";

import { EventBreadcrumb } from "@/features/event-list/event-breadcrumb";
import { EventDetails } from "@/widgets/event-details/event-details";
import { Footer } from "@/widgets/footer/footer";
import { Header } from "@/widgets/header/header";
import { parseEventRouteId } from "@/shared/lib/url";

type EventDetailsPageProps = {
  params: Promise<{
    eventId: string;
  }>;
};

export default async function EventDetailsPage({ params }: EventDetailsPageProps) {
  const { eventId } = await params;
  const normalizedEventId = parseEventRouteId(eventId);

  return (
    <div className="flex min-h-screen flex-col bg-bg">
      <Header variant="auto" />
      <main className="mx-auto w-full max-w-[1440px] flex-1 px-6 py-6">
        <nav className="mb-6 flex items-center gap-1.5 text-xs text-mute">
          <Link href="/events" className="transition-colors hover:text-ink">
            Мероприятия
          </Link>
          <ChevronRight className="h-3 w-3" />
          <EventBreadcrumb eventId={normalizedEventId} />
          <ChevronRight className="h-3 w-3" />
          <span className="font-medium text-ink">Подробности</span>
        </nav>
        <EventDetails eventId={normalizedEventId} />
      </main>
      <Footer />
    </div>
  );
}
