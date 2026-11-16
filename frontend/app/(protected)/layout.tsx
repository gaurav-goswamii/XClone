import { redirect } from "next/navigation";
import { getCurrentUser } from "@/lib/auth";

/**
 * Server-side gate for every route in the (protected) group.
 *
 * proxy.ts already bounces unauthenticated visitors, but it only decodes the
 * JWT's exp without verifying the signature — a UX gate, not a security one.
 * This layout is the real check: it asks the backend who the caller is and
 * redirects when the answer is nobody.
 *
 * A layout cannot hand props to the page it wraps, so the user is not passed
 * down from here. Each page re-asks getCurrentUser(), which is cached per
 * request, so the whole group still costs exactly one call to /api/auth/me.
 */
export default async function ProtectedLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const user = await getCurrentUser();
  if (!user) redirect("/login");

  return <>{children}</>;
}