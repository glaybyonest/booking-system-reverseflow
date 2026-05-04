export const routes = {
  home: "/",
  events: "/events",
  event: (eventId: string) => `/events/${eventId}`,
  session: (sessionId: string) => `/sessions/${sessionId}`,
  checkout: (bookingId: string) => `/checkout/${bookingId}`,
  bookings: "/bookings",
  notifications: "/notifications",
  login: (redirect?: string) => `/login${redirect ? `?redirect=${encodeURIComponent(redirect)}` : ""}`,
  register: "/register"
};
