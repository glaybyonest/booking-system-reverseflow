"use client";

import {
  AlertTriangle,
  Bell,
  CheckCircle,
  Info,
  XCircle
} from "lucide-react";

import { friendlyApiError } from "@/shared/api/errors";
import { formatDateTime } from "@/shared/lib/date";
import { Alert } from "@/shared/ui/alert";
import { Button } from "@/shared/ui/button";
import { EmptyState } from "@/shared/ui/empty-state";
import { Spinner } from "@/shared/ui/spinner";
import { useMarkNotificationRead, useNotifications } from "@/features/notifications/notifications.hooks";
import { cn } from "@/shared/lib/cn";

type NotificationType = "confirmed" | "held" | "expiring" | "failed" | "cancelled" | "default";

// Map backend notification type strings to our display types.
// Backend types: booking_confirmed, booking_held, booking_expiring,
//                payment_failed, booking_cancelled
function detectType(type: string): NotificationType {
  if (type === "booking_confirmed") return "confirmed";
  if (type === "booking_held") return "held";
  if (type === "booking_expiring") return "expiring";
  if (type === "payment_failed") return "failed";
  if (type === "booking_cancelled") return "cancelled";
  // Fallback: try title-based heuristics for any unknown types
  return "default";
}

const TYPE_CONFIG: Record<
  NotificationType,
  { icon: React.ElementType; iconClass: string; dotClass: string }
> = {
  confirmed: {
    icon: CheckCircle,
    iconClass: "text-ok bg-ok-soft",
    dotClass: "bg-ok"
  },
  held: {
    icon: Info,
    iconClass: "text-info bg-info-soft",
    dotClass: "bg-info"
  },
  expiring: {
    icon: AlertTriangle,
    iconClass: "text-warn bg-warn-soft",
    dotClass: "bg-warn"
  },
  failed: {
    icon: XCircle,
    iconClass: "text-err bg-err-soft",
    dotClass: "bg-err"
  },
  cancelled: {
    icon: Info,
    iconClass: "text-mute bg-bg",
    dotClass: "bg-mute-2"
  },
  default: {
    icon: Bell,
    iconClass: "text-mute bg-bg",
    dotClass: "bg-mute-2"
  }
};

export function NotificationList() {
  const notifications = useNotifications();
  const markRead = useMarkNotificationRead();

  if (notifications.isLoading) return <Spinner />;
  if (notifications.error)
    return <Alert variant="error">{friendlyApiError(notifications.error)}</Alert>;
  if (!notifications.data?.length) {
    return (
      <EmptyState
        title="Уведомлений пока нет"
        description="Здесь появятся события по вашим броням."
      />
    );
  }

  return (
    <div className="overflow-hidden rounded-2xl border border-border bg-white shadow-card">
      {notifications.data.map((notification, idx) => {
        const type = detectType(notification.type);
        const config = TYPE_CONFIG[type];
        const Icon = config.icon;

        return (
          <div
            key={notification.id}
            className={cn(
              "flex items-start gap-4 px-5 py-4 transition-colors",
              idx !== 0 && "border-t border-border",
              notification.isRead ? "bg-white" : "bg-bg"
            )}
          >
            {/* Unread dot */}
            <div className="mt-1 flex h-5 w-5 shrink-0 items-center justify-center">
              {!notification.isRead && (
                <span className={cn("h-2 w-2 rounded-full", config.dotClass)} />
              )}
            </div>

            {/* Icon */}
            <div
              className={cn(
                "flex h-8 w-8 shrink-0 items-center justify-center rounded-full",
                config.iconClass
              )}
            >
              <Icon className="h-4 w-4" />
            </div>

            {/* Content */}
            <div className="min-w-0 flex-1">
              <p className={cn("text-sm font-semibold", notification.isRead ? "text-mute" : "text-ink")}>
                {notification.title}
              </p>
              <p className="mt-0.5 text-xs leading-relaxed text-mute">
                {notification.message}
              </p>
            </div>

            {/* Time + action */}
            <div className="flex shrink-0 flex-col items-end gap-2">
              <span className="text-[11px] text-mute-2">
                {formatDateTime(notification.createdAt)}
              </span>
              {!notification.isRead ? (
                <Button
                  variant="ghost"
                  size="sm"
                  disabled={markRead.isPending}
                  onClick={() => markRead.mutate(notification.id)}
                  className="px-3 py-1 text-[11px]"
                >
                  Прочитано
                </Button>
              ) : (
                <span className="text-[11px] text-mute-2">Прочитано</span>
              )}
            </div>
          </div>
        );
      })}
    </div>
  );
}
