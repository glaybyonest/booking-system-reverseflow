import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
  loadYandexMapsApi,
  resetYandexMapsLoaderForTests,
  YandexMapsConfigError
} from "@/widgets/event-map/yandex-maps";

describe("yandex-maps loader", () => {
  beforeEach(() => {
    resetYandexMapsLoaderForTests();
    document.head.innerHTML = "";
    delete process.env.NEXT_PUBLIC_YANDEX_MAPS_API_KEY;
    delete window.ymaps3;
  });

  afterEach(() => {
    resetYandexMapsLoaderForTests();
    vi.restoreAllMocks();
    delete window.ymaps3;
  });

  it("fails when the public Yandex Maps key is missing", async () => {
    await expect(loadYandexMapsApi()).rejects.toBeInstanceOf(YandexMapsConfigError);
  });

  it("loads the script once and reuses the resolved API", async () => {
    process.env.NEXT_PUBLIC_YANDEX_MAPS_API_KEY = "test-key";

    const ymapsMock = {
      ready: Promise.resolve()
    } as typeof ymaps3;

    const appendSpy = vi.spyOn(document.head, "appendChild").mockImplementation((node) => {
      const script = node as HTMLScriptElement;
      window.ymaps3 = ymapsMock;
      queueMicrotask(() => {
        script.dispatchEvent(new Event("load"));
      });
      return node;
    });

    await expect(loadYandexMapsApi()).resolves.toBe(ymapsMock);
    await expect(loadYandexMapsApi()).resolves.toBe(ymapsMock);
    expect(appendSpy).toHaveBeenCalledTimes(1);
  });
});
