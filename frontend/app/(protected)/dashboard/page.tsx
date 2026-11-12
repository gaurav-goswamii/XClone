import { redirect } from "next/navigation";
import { getCurrentUser } from "@/lib/auth";
import DashboardView from "./DashboardView";

export default async function DashboardPage() {
  const user = await getCurrentUser();

  // The protected layout has already redirected if there is no session; this
  // repeat is what narrows CurrentUser | null to CurrentUser for the view. It
  // costs nothing — getCurrentUser() is cached for the request.
  if (!user) redirect("/login");

  return <DashboardView user={user} />;
}
