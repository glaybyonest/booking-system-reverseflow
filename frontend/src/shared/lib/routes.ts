import { buildEventPath } from "@/shared/lib/url";

export const routes = {
  home: "/",
  events: "/events",
  eventsMap: "/events/map",
  event: (eventId: string, title?: string | null) => buildEventPath(eventId, title),
  session: (sessionId: string) => `/sessions/${sessionId}`,
  adminSessionLayout: (sessionId: string) => `/admin/sessions/${sessionId}/layout`,
  adminHallLayout: (hallId: string) => `/admin/halls/${hallId}/layout`,
  checkout: (bookingId: string) => `/checkout/${bookingId}`,
  bookings: "/bookings",
  notifications: "/notifications",
  login: (redirect?: string) => `/login${redirect ? `?redirect=${encodeURIComponent(redirect)}` : ""}`,
  register: "/register"
};
