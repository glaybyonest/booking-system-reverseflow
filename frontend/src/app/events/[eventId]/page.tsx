import { EventDetails } from "@/widgets/event-details/event-details";
import { Header } from "@/widgets/header/header";
import { SessionList } from "@/features/session-list/session-list";

export default function EventDetailsPage({ params }: { params: { eventId: string } }) {
  return (
    <div className="min-h-screen bg-[#F8F9FA]">
      <Header variant="auto" />
      <main className="mx-auto grid max-w-7xl grid-cols-1 gap-8 px-4 py-8 sm:px-6 lg:grid-cols-12">
        <section className="lg:col-span-7">
          <EventDetails eventId={params.eventId} />
        </section>
        <section className="lg:col-span-5">
          <div className="mb-4">
            <h2 className="text-2xl font-bold">Сеансы</h2>
            <p className="mt-1 text-sm text-gray-500">Выберите удобное время и переходите к местам.</p>
          </div>
          <SessionList eventId={params.eventId} />
        </section>
      </main>
    </div>
  );
}
