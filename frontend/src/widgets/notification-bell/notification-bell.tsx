"use client";

import { Bell } from "lucide-react";
import Link from "next/link";

import { useNotifications } from "@/features/notifications/notifications.hooks";

export function NotificationBell() {
  const notifications = useNotifications();
  const unread = notifications.data?.some((item) => !item.isRead);

  return (
    <Link
      href="/notifications"
      className="relative rounded-full p-2 text-gray-400 transition-colors hover:bg-gray-100 hover:text-gray-900"
      aria-label="Уведомления"
    >
      <Bell className="h-5 w-5" />
      {unread ? (
        <span className="absolute right-1 top-1 h-2 w-2 rounded-full border border-white bg-red-500" />
      ) : null}
    </Link>
  );
}
