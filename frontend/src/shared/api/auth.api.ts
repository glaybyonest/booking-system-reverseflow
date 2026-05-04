import type { User } from "@/entities/user/types";

type AuthResult = {
  user: User;
  authenticated: boolean;
};

async function authFetch<T>(path: string, payload?: unknown): Promise<T> {
  const response = await fetch(`/api/auth/${path}`, {
    method: payload ? "POST" : "GET",
    headers: payload ? { "Content-Type": "application/json" } : undefined,
    body: payload ? JSON.stringify(payload) : undefined
  });
  const data = await response.json().catch(() => null);
  if (!response.ok) {
    const { toApiError } = await import("@/shared/api/client");
    throw toApiError(response.status, data);
  }
  return data as T;
}

export function login(input: { email: string; password: string }) {
  return authFetch<AuthResult>("login", input);
}

export function register(input: { name: string; email: string; password: string }) {
  return authFetch<AuthResult>("register", input);
}

export function logout() {
  return authFetch<{ success: boolean }>("logout", {});
}

export function getMe() {
  return authFetch<User>("me");
}
