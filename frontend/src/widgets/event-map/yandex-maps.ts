const SCRIPT_ID = "reserveflow-yandex-maps";
const API_URL = "https://api-maps.yandex.ru/v3/";
const READY_TIMEOUT_MS = 15_000;

let cachedYandexMapsPromise: Promise<typeof ymaps3> | null = null;

export class YandexMapsConfigError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "YandexMapsConfigError";
  }
}

/**
 * Wraps window.ymaps3.ready in a race with a timeout.
 *
 * In production, ymaps3.ready resolves once the library finishes init.
 * However, if the API key is not authorized for the current domain
 * (e.g. localhost not added in Yandex Console), the Promise silently hangs.
 * The timeout surfaces a clear error instead of an infinite spinner.
 */
function waitForReady(ymaps: NonNullable<typeof window.ymaps3>): Promise<void> {
  return Promise.race([
    Promise.resolve(ymaps.ready),
    new Promise<never>((_, reject) =>
      setTimeout(
        () =>
          reject(
            new Error(
              "Яндекс.Карты не инициализировались за 15 с. " +
                "Проверьте API-ключ и авторизацию домена на console.yandex.ru."
            )
          ),
        READY_TIMEOUT_MS
      )
    )
  ]);
}

export async function loadYandexMapsApi() {
  if (typeof window === "undefined") {
    throw new YandexMapsConfigError("Yandex Maps API can only be loaded in the browser.");
  }

  if (window.ymaps3) {
    await waitForReady(window.ymaps3);
    return window.ymaps3;
  }

  const apiKey = process.env.NEXT_PUBLIC_YANDEX_MAPS_API_KEY?.trim();
  if (!apiKey) {
    throw new YandexMapsConfigError(
      "Ключ Яндекс.Карт не задан. Добавьте NEXT_PUBLIC_YANDEX_MAPS_API_KEY в frontend/.env.local."
    );
  }

  if (!cachedYandexMapsPromise) {
    cachedYandexMapsPromise = new Promise<typeof ymaps3>((resolve, reject) => {
      const existingScript = document.getElementById(SCRIPT_ID) as HTMLScriptElement | null;
      if (existingScript) {
        if (window.ymaps3) {
          void handleReady();
          return;
        }
        existingScript.addEventListener("load", handleReady, { once: true });
        existingScript.addEventListener("error", handleError, { once: true });
        return;
      }

      const script = document.createElement("script");
      script.id = SCRIPT_ID;
      script.async = true;
      script.src = `${API_URL}?apikey=${encodeURIComponent(apiKey)}&lang=ru_RU`;
      script.addEventListener("load", handleReady, { once: true });
      script.addEventListener("error", handleError, { once: true });
      document.head.appendChild(script);

      async function handleReady() {
        if (!window.ymaps3) {
          reject(new Error("Yandex Maps API script loaded without window.ymaps3."));
          cachedYandexMapsPromise = null;
          return;
        }

        try {
          await waitForReady(window.ymaps3);
          resolve(window.ymaps3);
        } catch (error) {
          reject(error);
          cachedYandexMapsPromise = null;
        }
      }

      function handleError() {
        reject(new Error("Failed to load Yandex Maps API script."));
        cachedYandexMapsPromise = null;
      }
    });
  }

  return cachedYandexMapsPromise;
}

export function resetYandexMapsLoaderForTests() {
  cachedYandexMapsPromise = null;
}
