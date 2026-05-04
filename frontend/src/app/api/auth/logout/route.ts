import { cookies } from "next/headers";

import { backendJSON, logoutResponse } from "@/app/api/auth/_backend";

export async function POST() {
  const refreshToken = (await cookies()).get("refresh_token")?.value;
  if (refreshToken) {
    await backendJSON("/api/v1/auth/logout", {
      method: "POST",
      body: JSON.stringify({ refreshToken })
    }).catch(() => null);
  }
  return logoutResponse();
}
