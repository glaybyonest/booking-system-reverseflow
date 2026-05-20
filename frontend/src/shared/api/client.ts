import { AppApiError, isApiErrorPayload } from "@/shared/api/errors";

type ClientFetchOptions = RequestInit & {
  json?: unknown;
};

export async function clientFetch<T>(path: string, options: ClientFetchOptions = {}): Promise<T> {
  const normalizedPath = path.replace(/^\/+/, "");
  const headers = new Headers(options.headers);
  let body = options.body;
  if (options.json !== undefined) {
    headers.set("Content-Type", "application/json");
    body = JSON.stringify(options.json);
  }
  const response = await fetch(`/api/backend/${normalizedPath}`, {
    ...options,
    headers,
    body
  });
  const payload = await readPayload(response);
  if (!response.ok) {
    throw toApiError(response.status, payload);
  }
  return payload as T;
}

async function readPayload(response: Response) {
  const text = await response.text();
  if (!text) return null;
  try {
    return JSON.parse(text) as unknown;
  } catch {
    return text;
  }
}

export function toApiError(status: number, payload: unknown) {
  if (isApiErrorPayload(payload)) {
    return new AppApiError({
      code: payload.error.code,
      message: payload.error.message,
      status,
      details: payload.error.details
    });
  }
  return new AppApiError({
    code: status === 401 ? "UNAUTHORIZED" : "INTERNAL_ERROR",
    message: "Запрос не выполнен",
    status
  });
}
