import { SeatSelection } from "@/features/seat-selection/seat-selection";
import { Header } from "@/widgets/header/header";

export default function SessionPage({ params }: { params: { sessionId: string } }) {
  return (
    <div className="flex min-h-screen flex-col bg-[#F8F9FA]">
      <Header variant="auth" />
      <main className="mx-auto w-full max-w-7xl flex-1 px-4 py-8 sm:px-6">
        <SeatSelection sessionId={params.sessionId} />
      </main>
    </div>
  );
}
