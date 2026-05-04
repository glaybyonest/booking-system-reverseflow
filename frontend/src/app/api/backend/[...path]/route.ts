import { cookies } from "next/headers";
import { NextRequest, NextResponse } from "next/server";

import { backendApiUrl } from "@/shared/config/env";
import { refreshTokens } from "@/app/api/auth/_backend";
import { setAuthCookies, type TokenPair } from "@/app/api/auth/_cookies";

type Params = {
  params: Promise<{
    path: string[];
  }>;
};

export async function GET(request: NextRequest, context: Params) {
  return proxy(request, context);
}

export async function POST(request: NextRequest, context: Params) {
  return proxy(request, context);
}

export async function PUT(request: NextRequest, context: Params) {
  return proxy(request, context);
}

export async function PATCH(request: NextRequest, context: Params) {
  return proxy(request, context);
}

export async function DELETE(request: NextRequest, context: Params) {
  return proxy(request, context);
}

async function proxy(request: NextRequest, context: Params) {
  const accessToken = (await cookies()).get("access_token")?.value;
  const params = await context.params;
  const body =
    request.method === "GET" || request.method === "HEAD" ? undefined : await request.text();
  const first = await forward(request, params.path, body, accessToken);
  if (first.status !== 401) return first;

  const tokens = await refreshTokens();
  if (!tokens) return first;

  const second = await forward(request, params.path, body, tokens.accessToken, tokens);
  return second;
}

async function forward(
  request: NextRequest,
  path: string[],
  body?: string,
  accessToken?: string,
  refreshedTokens?: TokenPair
) {
  const target = new URL(`${backendApiUrl()}/api/v1/${path.map(encodeURIComponent).join("/")}`);
  request.nextUrl.searchParams.forEach((value, key) => target.searchParams.append(key, value));

  const headers = new Headers();
  headers.set("Accept", "application/json");
  const contentType = request.headers.get("Content-Type");
  if (contentType) headers.set("Content-Type", contentType);
  if (accessToken) headers.set("Authorization", `Bearer ${accessToken}`);

  const backendResponse = await fetch(target, {
    method: request.method,
    headers,
    body,
    cache: "no-store"
  });
  const text = await backendResponse.text();
  const response = new NextResponse(text, {
    status: backendResponse.status,
    headers: {
      "Content-Type": backendResponse.headers.get("Content-Type") ?? "application/json"
    }
  });
  if (refreshedTokens) setAuthCookies(response, refreshedTokens);
  return response;
}
