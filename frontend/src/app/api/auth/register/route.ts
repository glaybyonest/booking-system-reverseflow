import { normalizeUser } from "@/shared/api/mappers";
import { backendJSON, jsonWithCookies } from "@/app/api/auth/_backend";
import { extractTokens } from "@/app/api/auth/_cookies";

export async function POST(request: Request) {
  const body = await request.json();
  const { response, payload } = await backendJSON("/api/v1/auth/register", {
    method: "POST",
    body: JSON.stringify(body)
  });
  if (!response.ok) {
    return jsonWithCookies(payload, response.status);
  }
  const tokens = extractTokens(payload);
  const user = normalizeUser((payload as { user?: unknown })?.user);
  return jsonWithCookies({ user, authenticated: Boolean(tokens) }, response.status, tokens);
}
