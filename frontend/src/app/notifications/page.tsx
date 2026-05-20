import { NotificationList } from "@/features/notifications/notification-list";
import { Footer } from "@/widgets/footer/footer";
import { Header } from "@/widgets/header/header";

export default function NotificationsPage() {
  return (
    <div className="flex min-h-screen flex-col bg-bg">
      <Header variant="auth" />
      <main className="mx-auto w-full max-w-4xl flex-1 px-6 py-8">
        <div className="mb-6">
          <p className="text-[11px] font-bold uppercase tracking-[0.22em] text-mute-2">
            ЛИЧНЫЙ КАБИНЕТ
          </p>
          <h1 className="mt-1 text-[28px] font-extrabold tracking-tight text-ink">
            Уведомления
          </h1>
        </div>
        <NotificationList />
      </main>
      <Footer />
    </div>
  );
}
