import { redirect } from "next/navigation";
import { getCurrentUser } from "@/lib/auth";
import MessagesView from "./MessagesView";

export default async function MessagesPage() {
  const user = await getCurrentUser();
  if (!user) redirect("/login");

  return <MessagesView user={user} />;
}
