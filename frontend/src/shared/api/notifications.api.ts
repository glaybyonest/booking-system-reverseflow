import type { Notification } from "@/entities/notification/types";
import { clientFetch } from "@/shared/api/client";
import { normalizeNotification } from "@/shared/api/mappers";
import type { ApiList } from "@/shared/types/api";

export async function getNotifications(): Promise<Notification[]> {
  const data = await clientFetch<ApiList<unknown>>("notifications");
  return (data.items ?? []).map(normalizeNotification);
}

export async function markNotificationRead(id: string) {
  return clientFetch<{ status: string }>(`notifications/${id}/read`, { method: "POST" });
}
