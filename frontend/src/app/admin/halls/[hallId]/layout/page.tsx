import { HallLayoutPage } from "@/features/admin-layout/hall-layout-page";
import { Header } from "@/widgets/header/header";

export default async function AdminHallLayoutRoute({
  params
}: {
  params: Promise<{ hallId: string }>;
}) {
  const { hallId } = await params;

  return (
    <div className="min-h-screen bg-[#F8F9FA]">
      <Header variant="auto" />
      <main className="mx-auto max-w-7xl px-4 py-8 sm:px-6">
        <HallLayoutPage hallId={hallId} />
      </main>
    </div>
  );
}
