/**
 * The shape of the signed-in user, as returned by GET /api/auth/me.
 *
 * This file holds types only — no server imports — so client components can
 * import it freely. The fetching lives in lib/auth.ts, which is server-only.
 */
export interface CurrentUser {
  id: string;
  name: string;
  username: string;
  email?: string;
  birth_date?: string;
  avatar_url?: string;
  bio?: string;
  is_verified?: boolean;
  created_at?: string;
}
