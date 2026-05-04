import { NotificationList } from "@/features/notifications/notification-list";
import { Header } from "@/widgets/header/header";

export default function NotificationsPage() {
  return (
    <div className="min-h-screen bg-[#F8F9FA]">
      <Header variant="auth" />
      <main className="mx-auto max-w-4xl px-4 py-8 sm:px-6">
        <div className="mb-8">
          <h1 className="text-4xl font-extrabold tracking-tight">Уведомления</h1>
          <p className="mt-3 text-gray-500">События по оплатам, отменам и истекшим броням.</p>
        </div>
        <NotificationList />
      </main>
    </div>
  );
}
