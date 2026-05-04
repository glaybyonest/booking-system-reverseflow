import { cookies } from "next/headers";

import { backendApiUrl } from "@/shared/config/env";

export async function serverFetch<T>(path: string, options: RequestInit = {}): Promise<T | null> {
  const token = (await cookies()).get("access_token")?.value;
  const headers = new Headers(options.headers);
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }
  const response = await fetch(`${backendApiUrl()}/api/v1/${path.replace(/^\/+/, "")}`, {
    ...options,
    headers,
    cache: "no-store"
  });
  if (response.status === 401) {
    return null;
  }
  if (!response.ok) {
    throw new Error(`Backend request failed: ${response.status}`);
  }
  return (await response.json()) as T;
}
