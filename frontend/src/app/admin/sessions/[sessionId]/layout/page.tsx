import { SessionLayoutPage } from "@/features/admin-layout/session-layout-page";
import { Header } from "@/widgets/header/header";

export default async function AdminSessionLayoutRoute({
  params
}: {
  params: Promise<{ sessionId: string }>;
}) {
  const { sessionId } = await params;

  return (
    <div className="min-h-screen bg-[#F8F9FA]">
      <Header variant="auto" />
      <main className="mx-auto max-w-7xl px-4 py-8 sm:px-6">
        <SessionLayoutPage sessionId={sessionId} />
      </main>
    </div>
  );
}
