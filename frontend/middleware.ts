import { NextRequest, NextResponse } from "next/server";

const protectedPrefixes = ["/checkout", "/bookings", "/notifications", "/admin"];
const securityHeaders = {
  "Content-Security-Policy": "base-uri 'self'; frame-ancestors 'none'; object-src 'none';",
  "Referrer-Policy": "strict-origin-when-cross-origin",
  "X-Content-Type-Options": "nosniff",
  "X-Frame-Options": "DENY"
} as const;

export function middleware(request: NextRequest) {
  const isProtectedRoute = protectedPrefixes.some((prefix) =>
    request.nextUrl.pathname.startsWith(prefix)
  );
  const accessToken = request.cookies.get("access_token")?.value;

  if (isProtectedRoute && !accessToken) {
    const login = new URL("/login", request.url);
    login.searchParams.set("redirect", request.nextUrl.pathname + request.nextUrl.search);

    const response = NextResponse.redirect(login);
    setSecurityHeaders(response);
    return response;
  }

  const response = NextResponse.next();
  setSecurityHeaders(response);
  return response;
}

function setSecurityHeaders(response: NextResponse) {
  Object.entries(securityHeaders).forEach(([key, value]) => {
    response.headers.set(key, value);
  });
}

export const config = {
  matcher: ["/((?!api|_next/static|_next/image|favicon.ico|robots.txt|sitemap.xml).*)"]
};
