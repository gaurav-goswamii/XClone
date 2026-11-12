import { cache } from "react";
import { cookies } from "next/headers";
import type { CurrentUser } from "./user";

/**
 * Server-side only. Do not import this file from a client component — it reads
 * request cookies, which only exist on the server. Client components that need
 * the user receive it as a prop; see app/(protected)/layout.tsx.
 *
 * On the server there is no dev-server rewrite to lean on, so the backend has
 * to be addressed absolutely rather than via a relative /api path.
 */
const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

/**
 * Reads the signed-in user, or null when the request carries no valid session.
 *
 * Wrapped in React's cache(): the first caller during a request performs the
 * HTTP call and every later caller in that same render gets the memoised
 * result. That is what lets the protected layout and the page inside it both
 * ask for the user while only one request reaches the Go backend.
 */
export const getCurrentUser = cache(async (): Promise<CurrentUser | null> => {
  const token = (await cookies()).get("x_auth_token")?.value;
  if (!token) return null;

  try {
    const res = await fetch(`${API_BASE}/api/auth/me`, {
      headers: { Cookie: `x_auth_token=${token}` },
      cache: "no-store",
    });
    if (!res.ok) return null;

    const data = await res.json();
    if (data?.status === "success" && data.data?.user) {
      return data.data.user as CurrentUser;
    }
    return null;
  } catch {
    return null;
  }
});
