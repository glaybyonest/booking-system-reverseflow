"use client";

import { friendlyApiError } from "@/shared/api/errors";
import { formatDateTime } from "@/shared/lib/date";
import { Alert } from "@/shared/ui/alert";
import { Button } from "@/shared/ui/button";
import { Card, CardContent } from "@/shared/ui/card";
import { EmptyState } from "@/shared/ui/empty-state";
import { Spinner } from "@/shared/ui/spinner";
import { useMarkNotificationRead, useNotifications } from "@/features/notifications/notifications.hooks";

export function NotificationList() {
  const notifications = useNotifications();
  const markRead = useMarkNotificationRead();

  if (notifications.isLoading) return <Spinner />;
  if (notifications.error) return <Alert variant="error">{friendlyApiError(notifications.error)}</Alert>;
  if (!notifications.data?.length) {
    return <EmptyState title="Уведомлений пока нет" description="Здесь появятся события по вашим броням." />;
  }
  return (
    <div className="space-y-4">
      {notifications.data.map((notification) => (
        <Card key={notification.id} className={notification.isRead ? "opacity-70" : ""}>
          <CardContent className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
            <div className="flex gap-3">
              <span
                className={`mt-2 h-2.5 w-2.5 shrink-0 rounded-full ${
                  notification.isRead ? "bg-gray-200" : "bg-red-500"
                }`}
              />
              <div>
                <h3 className="font-bold">{notification.title}</h3>
                <p className="mt-1 text-sm text-gray-500">{notification.message}</p>
                <p className="mt-2 text-xs text-gray-400">{formatDateTime(notification.createdAt)}</p>
              </div>
            </div>
            {!notification.isRead ? (
              <Button
                variant="secondary"
                size="sm"
                disabled={markRead.isPending}
                onClick={() => markRead.mutate(notification.id)}
              >
                Прочитано
              </Button>
            ) : null}
          </CardContent>
        </Card>
      ))}
    </div>
  );
}
