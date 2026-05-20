const ROUTE_ID_DELIMITER = "--";
const REDIRECT_BASE_URL = "https://reserveflow.local";

export function slugifySegment(value: string | null | undefined, fallback = "event") {
  const normalized = (value ?? "")
    .trim()
    .toLowerCase()
    .normalize("NFKD")
    .replace(/[^\p{L}\p{N}]+/gu, "-")
    .replace(/^-+|-+$/g, "")
    .replace(/-{2,}/g, "-");

  return normalized || fallback;
}

export function buildEventPath(eventId: string, title?: string | null) {
  const slug = slugifySegment(title);
  return encodeURI(`/events/${slug}${ROUTE_ID_DELIMITER}${eventId}`);
}

export function parseEventRouteId(value: string) {
  let normalized = value.trim();
  try {
    normalized = decodeURIComponent(value).trim();
  } catch {
    normalized = value.trim();
  }
  if (!normalized) return "";

  const delimiterIndex = normalized.lastIndexOf(ROUTE_ID_DELIMITER);
  if (delimiterIndex === -1) return normalized;

  const parsedId = normalized.slice(delimiterIndex + ROUTE_ID_DELIMITER.length);
  return parsedId || normalized;
}

export function toSafeExternalUrl(value: string | null | undefined) {
  const normalized = value?.trim();
  if (!normalized) return null;

  try {
    const target = new URL(normalized);
    if (target.protocol !== "http:" && target.protocol !== "https:") {
      return null;
    }
    return target.toString();
  } catch {
    return null;
  }
}

export function toSafeRedirectPath(value: string | null | undefined, fallback = "/events") {
  const normalized = value?.trim();
  if (!normalized || !normalized.startsWith("/") || normalized.startsWith("//")) {
    return fallback;
  }

  try {
    const target = new URL(normalized, REDIRECT_BASE_URL);
    if (target.origin !== REDIRECT_BASE_URL) {
      return fallback;
    }
    return `${target.pathname}${target.search}${target.hash}`;
  } catch {
    return fallback;
  }
}
