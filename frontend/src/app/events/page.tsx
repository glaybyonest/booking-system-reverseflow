import { EventList } from "@/features/event-list/event-list";
import { Header } from "@/widgets/header/header";

export default function EventsPage() {
  return (
    <div className="min-h-screen bg-[#F8F9FA]">
      <Header variant="auto" />
      <main className="mx-auto max-w-7xl px-4 py-8 sm:px-6">
        <div className="mb-8">
          <h1 className="text-4xl font-extrabold tracking-tight">Мероприятия</h1>
          <p className="mt-3 max-w-2xl text-gray-500">
            Выберите событие, затем сеанс и свободное место на интерактивной схеме.
          </p>
        </div>
        <EventList />
      </main>
    </div>
  );
}
