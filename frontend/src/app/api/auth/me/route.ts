import { cookies } from "next/headers";
import { NextResponse } from "next/server";

import { normalizeUser } from "@/shared/api/mappers";
import { backendJSON, refreshTokens } from "@/app/api/auth/_backend";
import { setAuthCookies } from "@/app/api/auth/_cookies";

export async function GET() {
  const accessToken = (await cookies()).get("access_token")?.value;
  const first = await fetchMe(accessToken);
  if (first.response.ok) {
    return NextResponse.json(normalizeUser(first.payload));
  }
  if (first.response.status !== 401) {
    return NextResponse.json(first.payload, { status: first.response.status });
  }
  const tokens = await refreshTokens();
  if (!tokens) {
    return NextResponse.json(first.payload, { status: 401 });
  }
  const second = await fetchMe(tokens.accessToken);
  const response = NextResponse.json(
    second.response.ok ? normalizeUser(second.payload) : second.payload,
    { status: second.response.status }
  );
  setAuthCookies(response, tokens);
  return response;
}

function fetchMe(accessToken?: string) {
  return backendJSON("/api/v1/auth/me", {
    headers: accessToken ? { Authorization: `Bearer ${accessToken}` } : undefined
  });
}
