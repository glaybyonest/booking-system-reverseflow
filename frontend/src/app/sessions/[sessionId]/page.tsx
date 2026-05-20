import { SeatSelection } from "@/features/seat-selection/seat-selection";
import { Footer } from "@/widgets/footer/footer";
import { Header } from "@/widgets/header/header";

type SessionPageProps = {
  params: Promise<{
    sessionId: string;
  }>;
};

export default async function SessionPage({ params }: SessionPageProps) {
  const { sessionId } = await params;

  return (
    <div className="flex min-h-screen flex-col bg-bg">
      <Header variant="auto" />
      <main className="mx-auto w-full max-w-[1440px] flex-1 px-6 py-8">
        <SeatSelection sessionId={sessionId} />
      </main>
      <Footer />
    </div>
  );
}
