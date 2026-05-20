import { afterEach, describe, expect, it, vi } from "vitest";

import {
  BackendApiResolutionError,
  resetBackendApiUrlCacheForTests,
  resolveBackendApiUrl
} from "@/shared/config/env";

describe("resolveBackendApiUrl", () => {
  afterEach(() => {
    delete process.env.BACKEND_API_URL;
    resetBackendApiUrlCacheForTests();
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it("uses BACKEND_API_URL override without probing local ports", async () => {
    process.env.BACKEND_API_URL = "http://configured.example";
    const fetchSpy = vi.fn();
    vi.stubGlobal("fetch", fetchSpy);

    await expect(resolveBackendApiUrl()).resolves.toBe("http://configured.example");
    expect(fetchSpy).not.toHaveBeenCalled();
  });

  it("prefers healthy backend on localhost:18080 and caches the result", async () => {
    const fetchSpy = vi.fn(async (input: string | URL | Request) => {
      expect(String(input)).toBe("http://localhost:18080/health");
      return new Response(JSON.stringify({ status: "ok" }), {
        status: 200,
        headers: {
          "Content-Type": "application/json"
        }
      });
    });
    vi.stubGlobal("fetch", fetchSpy);

    await expect(resolveBackendApiUrl()).resolves.toBe("http://localhost:18080");
    await expect(resolveBackendApiUrl()).resolves.toBe("http://localhost:18080");
    expect(fetchSpy).toHaveBeenCalledTimes(1);
  });

  it("falls back to localhost:8080 when localhost:18080 is unavailable", async () => {
    const fetchSpy = vi.fn(async (input: string | URL | Request) => {
      const url = String(input);
      if (url === "http://localhost:18080/health") {
        throw new TypeError("connect ECONNREFUSED");
      }
      if (url === "http://localhost:8080/health") {
        return new Response(JSON.stringify({ status: "ok" }), {
          status: 200,
          headers: {
            "Content-Type": "application/json"
          }
        });
      }
      throw new Error(`Unexpected probe: ${url}`);
    });
    vi.stubGlobal("fetch", fetchSpy);

    await expect(resolveBackendApiUrl()).resolves.toBe("http://localhost:8080");
    expect(fetchSpy).toHaveBeenCalledTimes(2);
  });

  it("ignores a foreign service on localhost:8080 that does not expose ReserveFlow health", async () => {
    const fetchSpy = vi.fn(async (input: string | URL | Request) => {
      const url = String(input);
      if (url === "http://localhost:18080/health") {
        throw new TypeError("connect ECONNREFUSED");
      }
      if (url === "http://localhost:8080/health") {
        return new Response("<html>Server is up</html>", {
          status: 200,
          headers: {
            "Content-Type": "text/html"
          }
        });
      }
      throw new Error(`Unexpected probe: ${url}`);
    });
    vi.stubGlobal("fetch", fetchSpy);

    await expect(resolveBackendApiUrl()).rejects.toBeInstanceOf(BackendApiResolutionError);
  });

  it("throws a helpful error when no healthy backend is available", async () => {
    const fetchSpy = vi.fn(async () => new Response("not found", { status: 404 }));
    vi.stubGlobal("fetch", fetchSpy);

    await expect(resolveBackendApiUrl()).rejects.toThrow(
      "ReserveFlow backend недоступен. Проверьте BACKEND_API_URL"
    );
    expect(fetchSpy).toHaveBeenCalledTimes(2);
  });
});
