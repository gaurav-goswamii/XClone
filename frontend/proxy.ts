import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

function isTokenValid(token: string | undefined): boolean {
  if (!token) return false;
  try {
    const parts = token.split(".");
    if (parts.length !== 3) return false;

    let base64 = parts[1].replace(/-/g, "+").replace(/_/g, "/");
    while (base64.length % 4 !== 0) {
      base64 += "=";
    }

    const jsonStr =
      typeof atob === "function"
        ? atob(base64)
        : Buffer.from(base64, "base64").toString("utf-8");

    const payload = JSON.parse(jsonStr);

    if (payload.exp && typeof payload.exp === "number") {
      const nowInSeconds = Math.floor(Date.now() / 1000);
      if (payload.exp <= nowInSeconds) {
        return false;
      }
    }
    return true;
  } catch {
    return false;
  }
}

export function proxy(request: NextRequest) {
  const token = request.cookies.get("x_auth_token")?.value;
  const isValid = isTokenValid(token);
  const { pathname } = request.nextUrl;

  // Protected routes: redirect unauthenticated or expired users to /login
  if (pathname.startsWith("/dashboard") || pathname.startsWith("/messages")) {
    if (!isValid) {
      const loginUrl = new URL("/login", request.url);
      const response = NextResponse.redirect(loginUrl);
      if (token) {
        response.cookies.delete("x_auth_token");
      }
      return response;
    }
  }

  // Auth pages: redirect already authenticated users to /dashboard
  if (pathname === "/login" || pathname === "/signup") {
    if (isValid) {
      const dashboardUrl = new URL("/dashboard", request.url);
      return NextResponse.redirect(dashboardUrl);
    }
    // If token exists but is expired/malformed, clear it so user can cleanly authenticate
    if (token && !isValid) {
      const response = NextResponse.next();
      response.cookies.delete("x_auth_token");
      return response;
    }
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    "/dashboard",
    "/dashboard/:path*",
    "/messages",
    "/messages/:path*",
    "/login",
    "/signup",
  ],
};
