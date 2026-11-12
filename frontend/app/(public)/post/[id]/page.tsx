import { getCurrentUser } from "@/lib/auth";
import PostDetailView from "./PostDetailView";

interface PageProps {
  params: Promise<{ id: string }>;
}

export default async function PostDetailPage({ params }: PageProps) {
  const { id } = await params;
  const currentUser = await getCurrentUser();

  return (
    <PostDetailView postId={decodeURIComponent(id)} currentUser={currentUser} />
  );
}
