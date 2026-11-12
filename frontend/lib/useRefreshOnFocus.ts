"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";

/**
 * Re-runs the current route's server components whenever the tab regains focus.
 *
 * The pages used to re-fetch /api/auth/me on focus so that a session expiring
 * in another tab would not leave a stale signed-in view. Now that the user is
 * resolved on the server, refreshing the route achieves the same thing: the
 * server re-reads the session and the protected layout redirects if it is gone.
 * Client state in the view (feed posts, drafts) survives a refresh.
 */
export function useRefreshOnFocus(): void {
  const router = useRouter();

  useEffect(() => {
    const handleFocus = () => router.refresh();
    window.addEventListener("focus", handleFocus);
    return () => window.removeEventListener("focus", handleFocus);
  }, [router]);
}
