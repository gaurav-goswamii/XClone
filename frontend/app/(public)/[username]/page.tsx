import { getCurrentUser } from "@/lib/auth";
import ProfileView from "./ProfileView";

interface PageProps {
  params: Promise<{ username: string }>;
}

export default async function UserProfilePage({ params }: PageProps) {
  const { username } = await params;

  // A public route: a signed-out visitor gets null and sees the Sign in button
  // plus the non-owner view, exactly as before.
  const currentUser = await getCurrentUser();

  return (
    <ProfileView
      targetUsername={decodeURIComponent(username).replace(/^@/, "")}
      currentUser={currentUser}
    />
  );
}
