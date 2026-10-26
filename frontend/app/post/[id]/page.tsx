"use client";

import React, { useState, useEffect, use } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import PostCard, { PostResponse } from "../../../components/post/PostCard";
import EditPostModal from "../../../components/post/EditPostModal";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "";

interface CurrentUser {
  id: string;
  username: string;
}

interface PageProps {
  params: Promise<{ id: string }>;
}

export default function PostDetailPage({ params }: PageProps) {
  const router = useRouter();
  const resolvedParams = use(params);
  const postId = decodeURIComponent(resolvedParams.id);

  const [post, setPost] = useState<PostResponse | null>(null);
  const [currentUser, setCurrentUser] = useState<CurrentUser | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [notFound, setNotFound] = useState<boolean>(false);
  const [actionError, setActionError] = useState<string | null>(null);
  const [isEditOpen, setIsEditOpen] = useState<boolean>(false);

  // Who is viewing? Only needed to decide whether to offer Edit/Delete
  useEffect(() => {
    let isCancelled = false;

    fetch(`${API_BASE}/api/auth/me`, { credentials: "include" })
      .then((res) => (res.ok ? res.json() : null))
      .then((data) => {
        if (!isCancelled && data?.status === "success" && data.data?.user) {
          setCurrentUser(data.data.user);
        }
      })
      .catch(() => {
        // Viewing while logged out is fine: the post itself is public
      });

    return () => {
      isCancelled = true;
    };
  }, []);

  // The post itself, from the public single-post endpoint
  useEffect(() => {
    let isCancelled = false;

    fetch(`${API_BASE}/api/posts/${encodeURIComponent(postId)}`, {
      credentials: "include",
    })
      .then(async (res) => {
        if (res.status === 404) {
          if (!isCancelled) setNotFound(true);
          return null;
        }
        return res.ok ? res.json() : null;
      })
      .then((data) => {
        if (!isCancelled && data?.status === "success" && data.data?.post) {
          setPost(data.data.post);
        }
      })
      .catch(() => {
        if (!isCancelled) setNotFound(true);
      })
      .finally(() => {
        if (!isCancelled) setIsLoading(false);
      });

    return () => {
      isCancelled = true;
    };
  }, [postId]);

  const handleDelete = async () => {
    if (!window.confirm("Delete this post? This cannot be undone.")) return;

    setActionError(null);

    try {
      const res = await fetch(`${API_BASE}/api/posts/${encodeURIComponent(postId)}`, {
        method: "DELETE",
        credentials: "include",
      });

      // Gone is gone, whether we deleted it now or it had already been removed
      if (res.ok || res.status === 404) {
        router.push("/dashboard");
        return;
      }

      const data = await res.json().catch(() => null);
      setActionError(data?.message || "Failed to delete post. Please try again.");
    } catch {
      setActionError("Network error. Could not delete post.");
    }
  };

  const isOwner = Boolean(post && currentUser && post.user_id === currentUser.id);
  const wasEdited = Boolean(
    post && post.updated_at && new Date(post.updated_at) > new Date(post.created_at)
  );

  return (
    <div className="min-h-screen bg-black text-[#e7e9ea] flex justify-center selection:bg-[#1d9bf0]/30 selection:text-white">
      <main className="w-full max-w-[600px] border-x border-[#2f3336] min-h-screen">
        {/* Sticky Header */}
        <div className="sticky top-0 z-20 flex items-center gap-6 px-4 py-3 bg-black/80 backdrop-blur-md border-b border-[#2f3336]">
          <button
            type="button"
            onClick={() => router.back()}
            className="w-9 h-9 rounded-full flex items-center justify-center hover:bg-[#181818] transition text-[#eff3f4] cursor-pointer"
            aria-label="Go back"
          >
            <svg viewBox="0 0 24 24" className="w-5 h-5 fill-current">
              <path d="M7.414 13l5.043 5.04-1.414 1.42L3.586 12l7.457-7.46 1.414 1.42L7.414 11H21v2H7.414z" />
            </svg>
          </button>
          <h1 className="text-xl font-bold text-[#eff3f4]">Post</h1>
        </div>

        {actionError && (
          <div className="mx-4 mt-3 p-3 bg-[#f4212e]/10 border border-[#f4212e]/30 rounded-xl text-[#f4212e] text-sm">
            {actionError}
          </div>
        )}

        {isLoading ? (
          <div className="py-16 flex justify-center items-center">
            <div className="w-6 h-6 border-2 border-[#1d9bf0] border-t-transparent rounded-full animate-spin" />
          </div>
        ) : notFound || !post ? (
          <div className="py-16 px-6 text-center">
            <p className="text-xl font-bold text-[#eff3f4]">This post doesn&apos;t exist</p>
            <p className="text-sm text-[#71767b] mt-1.5">
              It may have been deleted by its author.
            </p>
            <Link
              href="/dashboard"
              className="inline-block mt-5 px-4 py-2 rounded-full bg-[#1d9bf0] text-white font-bold text-sm hover:bg-[#1a8cd8] transition"
            >
              Back to timeline
            </Link>
          </div>
        ) : (
          <>
            <PostCard
              post={post}
              isOwner={isOwner}
              onEdit={() => setIsEditOpen(true)}
              onDelete={handleDelete}
            />

            {/* Exact timestamps, which the compact card view omits */}
            <div className="px-4 py-3 text-xs text-[#71767b] border-b border-[#2f3336] space-y-1">
              <p>Posted {new Date(post.created_at).toLocaleString()}</p>
              {wasEdited && post.updated_at && (
                <p>Edited {new Date(post.updated_at).toLocaleString()}</p>
              )}
            </div>
          </>
        )}
      </main>

      {/* Edit Post Modal */}
      {post && isEditOpen && (
        <EditPostModal
          isOpen={isEditOpen}
          post={post}
          onClose={() => setIsEditOpen(false)}
          onUpdated={(updatedPost) => setPost(updatedPost)}
        />
      )}
    </div>
  );
}
