"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import PostCard, { PostResponse } from "@/components/post/PostCard";
import EditPostModal from "@/components/post/EditPostModal";
import type { CurrentUser } from "@/lib/user";
import CommentCard, {
  type CommentResponse,
} from "@/components/comment/CommentCard"
import CommentComposer from "@/components/comment/CommentComposer";
import { setPostLike } from "@/lib/postLikes";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "";

interface PostDetailViewProps {
  postId: string;
  /** The signed-in viewer, or null when nobody is signed in. */
  currentUser: CurrentUser | null;
}

export default function PostDetailView({
  postId,
  currentUser,
}: PostDetailViewProps) {
  const router = useRouter();

  const [post, setPost] = useState<PostResponse | null>(null);
  const [comments, setComments] = useState<CommentResponse[]>([]);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [notFound, setNotFound] = useState<boolean>(false);
  const [actionError, setActionError] = useState<string | null>(null);
  const [isEditOpen, setIsEditOpen] = useState<boolean>(false);
  const [isLikePending, setIsLikePending] = useState(false);

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


useEffect(() => {
  let isCancelled = false;

  fetch(
    `${API_BASE}/api/posts/${encodeURIComponent(postId)}/comments`,
    { credentials: "include" }
  )
    .then(async (res) => {
      if (!res.ok) return null;
      return res.json();
    })
    .then((data) => {
      if (
        !isCancelled &&
        data?.status === "success" &&
        Array.isArray(data.data?.comments)
      ) {
        setComments(data.data.comments);
      }
    })
    .catch(() => {
      if (!isCancelled) setComments([]);
    });

  return () => {
    isCancelled = true;
  };
}, [postId]);


const handleLike = async (postId: string, nextLiked: boolean) => {
  if (!currentUser) {
    router.push("/login")
    return
  }

  if (!post || post.id !== postId) return

  const previousPost = post
  setIsLikePending(true)

  setPost({
    ...post,
    liked_by_viewer: nextLiked,
    likes: Math.max(
      0,
      post.likes + (nextLiked ? 1 : -1)
    )
  })

  try {const result = await setPostLike(postId, nextLiked)

    setPost((current) => current ? {
      ...current,
      liked_by_viewer: result.liked,
      likes: result.likes
    } : current)
  } catch {
    setPost(previousPost)
    setActionError("Could not update the like. please try again.")
  } finally {
    setIsLikePending(false)
  }
}

  

  const handleDelete = async () => {
    if (!window.confirm("Delete this post? This cannot be undone.")) return;

    setActionError(null);

    try {
      const res = await fetch(`${API_BASE}/api/posts/${encodeURIComponent(postId)}`, {
        method: "DELETE",
        credentials: "include",
      });

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

  const handleCommentPosted = (comment: CommentResponse) => {
    setComments((existingComments) => [...existingComments, comment])

    setPost((existingPost) => 
    existingPost?{...existingPost, replies: existingPost.replies + 1}
  : existingPost)  
}

const handleCommentDelete = async (commentId: string) => {
  if(!window.confirm("Delete this comment? This cannot be undone.")) return

  setActionError(null)

  try{
    const res = await fetch(`${API_BASE}/api/comments/${encodeURIComponent(commentId)}`,
  {
    method: "DELETE",
    credentials: "include"
  })
  if (res.ok) {
    setComments((existingComments) => existingComments.filter((comment) => comment.id !== commentId))

    setPost((existingPost) => existingPost?{...existingPost, replies:Math.max(0, existingPost.replies -1)}
  :existingPost)
  return
  }

  const data = await res.json().catch(() => null)
  setActionError(data?.message || "Failed to delete comment. Please try again")
  } catch {
    setActionError("Network error. Could not delete comment.")
  }
}



  const isOwner = Boolean(post && currentUser && post.user_id === currentUser.id);
  const wasEdited = Boolean(
    post && post.updated_at && new Date(post.updated_at) > new Date(post.created_at)
  );

  return (
    <div className="min-h-screen bg-black text-[#e7e9ea] flex justify-center selection:bg-[#1d9bf0]/30 selection:text-white">
      <main className="w-full max-w-[600px] border-x border-[#2f3336] min-h-screen">
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
              isLikePending={isLikePending}
              onLike={handleLike}
              onEdit={() => setIsEditOpen(true)}
              onDelete={handleDelete}
            />

            <div className="px-4 py-3 text-xs text-[#71767b] border-b border-[#2f3336] space-y-1">
              <p>Posted {new Date(post.created_at).toLocaleString()}</p>
              {wasEdited && post.updated_at && (
                <p>Edited {new Date(post.updated_at).toLocaleString()}</p>
              )}
            </div>
            {currentUser && (
              <CommentComposer postId={postId} onPosted={handleCommentPosted}/>
            )}

            {comments.map((comment) => (
              <CommentCard key={comment.id} comment={comment} isOwner={comment.user_id === currentUser?.id}
              onDelete={handleCommentDelete}/>
              
            ))}
          </>
        )}
      </main>

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
