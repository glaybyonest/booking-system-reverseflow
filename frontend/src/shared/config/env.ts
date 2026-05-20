const LOCAL_BACKEND_CANDIDATES = ["http://localhost:18080", "http://localhost:8080"] as const;
const BACKEND_URL_CACHE_TTL_MS = 5_000;
const BACKEND_HEALTH_PATH = "/health";
const BACKEND_HEALTH_TIMEOUT_MS = 1_500;

let cachedBackendApiUrl: { expiresAt: number; value: string } | null = null;
let pendingBackendApiUrl: null | Promise<string> = null;

export class BackendApiResolutionError extends Error {
  code = "BACKEND_UNAVAILABLE";

  constructor(message = defaultBackendApiUrlMessage()) {
    super(message);
    this.name = "BackendApiResolutionError";
  }
}

export async function resolveBackendApiUrl() {
  const configuredUrl = process.env.BACKEND_API_URL?.trim();
  if (configuredUrl) {
    return configuredUrl;
  }

  const now = Date.now();
  if (cachedBackendApiUrl && cachedBackendApiUrl.expiresAt > now) {
    return cachedBackendApiUrl.value;
  }

  if (!pendingBackendApiUrl) {
    pendingBackendApiUrl = probeLocalBackendApiUrl();
  }

  try {
    const resolvedUrl = await pendingBackendApiUrl;
    cachedBackendApiUrl = {
      value: resolvedUrl,
      expiresAt: Date.now() + BACKEND_URL_CACHE_TTL_MS
    };
    return resolvedUrl;
  } finally {
    pendingBackendApiUrl = null;
  }
}

export function backendUnavailablePayload(error?: unknown) {
  return {
    error: {
      code: "BACKEND_UNAVAILABLE",
      message:
        error instanceof BackendApiResolutionError
          ? error.message
          : defaultBackendApiUrlMessage(),
      details: {
        candidates: [...LOCAL_BACKEND_CANDIDATES]
      }
    }
  };
}

export function resetBackendApiUrlCacheForTests() {
  cachedBackendApiUrl = null;
  pendingBackendApiUrl = null;
}

async function probeLocalBackendApiUrl() {
  for (const candidate of LOCAL_BACKEND_CANDIDATES) {
    if (await isReserveFlowBackend(candidate)) {
      return candidate;
    }
  }

  throw new BackendApiResolutionError();
}

async function isReserveFlowBackend(baseUrl: string) {
  try {
    const response = await fetch(`${baseUrl}${BACKEND_HEALTH_PATH}`, {
      method: "GET",
      cache: "no-store",
      headers: {
        Accept: "application/json"
      },
      signal: getProbeAbortSignal()
    });
    if (!response.ok) {
      return false;
    }

    const contentType = response.headers.get("Content-Type")?.toLowerCase() ?? "";
    if (!contentType.includes("application/json")) {
      return false;
    }

    const payload = (await response.json().catch(() => null)) as unknown;
    return isHealthyBackendPayload(payload);
  } catch {
    return false;
  }
}

function isHealthyBackendPayload(payload: unknown): payload is { status: string } {
  return (
    typeof payload === "object" &&
    payload !== null &&
    "status" in payload &&
    (payload as { status?: unknown }).status === "ok"
  );
}

function getProbeAbortSignal() {
  if (typeof AbortSignal === "undefined" || typeof AbortSignal.timeout !== "function") {
    return undefined;
  }
  return AbortSignal.timeout(BACKEND_HEALTH_TIMEOUT_MS);
}

function defaultBackendApiUrlMessage() {
  return "ReserveFlow backend недоступен. Проверьте BACKEND_API_URL или локальный API на http://localhost:18080 либо http://localhost:8080.";
}
