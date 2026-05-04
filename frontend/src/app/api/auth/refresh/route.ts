import { NextResponse } from "next/server";

import { refreshTokens } from "@/app/api/auth/_backend";
import { setAuthCookies } from "@/app/api/auth/_cookies";

export async function POST() {
  const tokens = await refreshTokens();
  if (!tokens) {
    return NextResponse.json(
      { error: { code: "UNAUTHORIZED", message: "Unable to refresh session", details: {} } },
      { status: 401 }
    );
  }
  const response = NextResponse.json({ success: true });
  setAuthCookies(response, tokens);
  return response;
}
