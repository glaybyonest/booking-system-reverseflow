import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { backendApiUrl } from "@/shared/config/env";
import { clearAuthCookies, extractTokens, setAuthCookies } from "@/app/api/auth/_cookies";

export async function backendJSON(path: string, init: RequestInit = {}) {
  const headers = new Headers(init.headers);
  if (!headers.has("Content-Type") && init.body) {
    headers.set("Content-Type", "application/json");
  }
  const response = await fetch(`${backendApiUrl()}${path}`, {
    ...init,
    headers,
    cache: "no-store"
  });
  const payload = await response.json().catch(() => null);
  return { response, payload };
}

export async function refreshTokens() {
  const refreshToken = (await cookies()).get("refresh_token")?.value;
  if (!refreshToken) return null;
  const { response, payload } = await backendJSON("/api/v1/auth/refresh", {
    method: "POST",
    body: JSON.stringify({ refreshToken })
  });
  if (!response.ok) return null;
  return extractTokens(payload);
}

export async function jsonWithCookies(payload: unknown, status: number, tokens?: ReturnType<typeof extractTokens>) {
  const res = NextResponse.json(payload, { status });
  if (tokens) setAuthCookies(res, tokens);
  return res;
}

export function logoutResponse(payload: unknown = { success: true }) {
  const res = NextResponse.json(payload);
  clearAuthCookies(res);
  return res;
}
