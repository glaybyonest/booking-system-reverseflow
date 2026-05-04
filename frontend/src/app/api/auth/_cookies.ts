import { NextResponse } from "next/server";

const secure = process.env.NODE_ENV === "production";

export type TokenPair = {
  accessToken: string;
  refreshToken: string;
};

export function extractTokens(payload: unknown): TokenPair | null {
  const root = asRecord(payload);
  const tokens = asRecord(root.tokens);
  const accessToken = stringValue(root.accessToken ?? root.access_token ?? tokens.accessToken ?? tokens.access_token);
  const refreshToken = stringValue(
    root.refreshToken ?? root.refresh_token ?? tokens.refreshToken ?? tokens.refresh_token
  );
  if (!accessToken || !refreshToken) return null;
  return { accessToken, refreshToken };
}

export function setAuthCookies(response: NextResponse, tokens: TokenPair) {
  response.cookies.set("access_token", tokens.accessToken, {
    httpOnly: true,
    sameSite: "lax",
    secure,
    path: "/"
  });
  response.cookies.set("refresh_token", tokens.refreshToken, {
    httpOnly: true,
    sameSite: "lax",
    secure,
    path: "/"
  });
}

export function clearAuthCookies(response: NextResponse) {
  response.cookies.set("access_token", "", {
    httpOnly: true,
    sameSite: "lax",
    secure,
    path: "/",
    maxAge: 0
  });
  response.cookies.set("refresh_token", "", {
    httpOnly: true,
    sameSite: "lax",
    secure,
    path: "/",
    maxAge: 0
  });
}

function asRecord(value: unknown): Record<string, unknown> {
  return typeof value === "object" && value !== null ? (value as Record<string, unknown>) : {};
}

function stringValue(value: unknown) {
  return typeof value === "string" && value.length > 0 ? value : null;
}
