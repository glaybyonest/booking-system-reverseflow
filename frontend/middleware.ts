import { NextRequest, NextResponse } from "next/server";

export function middleware(request: NextRequest) {
  const accessToken = request.cookies.get("access_token")?.value;
  if (accessToken) return NextResponse.next();

  const login = new URL("/login", request.url);
  login.searchParams.set("redirect", request.nextUrl.pathname + request.nextUrl.search);
  return NextResponse.redirect(login);
}

export const config = {
  matcher: ["/checkout/:path*", "/bookings/:path*", "/notifications/:path*"]
};
