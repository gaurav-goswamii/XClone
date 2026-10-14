"use client";

import React from "react";
import Link from "next/link";
import UserAvatar from "../ui/UserAvatar";

export interface PostResponse {
  id: string;
  user_id: string;
  author_name: string;
  author_username: string;
  author_avatar?: string;
  content: string;
  media_url?: string;
  likes: number;
  retweets: number;
  replies: number;
  created_at: string;
  updated_at?: string;
}

interface PostCardProps {
  post: PostResponse;
  onLike?: (postId: string) => void;
  onRetweet?: (postId: string) => void;
  onReply?: (postId: string) => void;
}

function formatRelativeTime(dateString: string): string {
  try {
    const postDate = new Date(dateString);
    if (isNaN(postDate.getTime())) {
      return "";
    }
    const now = new Date();
    const diffSeconds = Math.max(0, Math.floor((now.getTime() - postDate.getTime()) / 1000));

    if (diffSeconds < 60) {
      return "now";
    }
    const diffMinutes = Math.floor(diffSeconds / 60);
    if (diffMinutes < 60) {
      return `${diffMinutes}m`;
    }
    const diffHours = Math.floor(diffMinutes / 60);
    if (diffHours < 24) {
      return `${diffHours}h`;
    }
    const diffDays = Math.floor(diffHours / 24);
    if (diffDays < 7) {
      return `${diffDays}d`;
    }
    return postDate.toLocaleDateString("en-US", { month: "short", day: "numeric" });
  } catch {
    return "";
  }
}

export default function PostCard({ post, onLike, onRetweet, onReply }: PostCardProps) {
  const timeFormatted = formatRelativeTime(post.created_at);

  return (
    <article
      id={`post-${post.id}`}
      className="p-4 border-b border-[#2f3336] hover:bg-[#080808]/80 transition duration-150 flex gap-3 select-text"
    >
      {/* Author Avatar */}
      <Link
        href={`/${post.author_username}`}
        className="shrink-0 group"
        aria-label={`View ${post.author_name}'s profile`}
      >
        <UserAvatar
          src={post.author_avatar}
          name={post.author_name}
          size="md"
          className="group-hover:opacity-90 transition cursor-pointer"
        />
      </Link>

      {/* Post Content & Meta */}
      <div className="flex-1 min-w-0">
        {/* Author Header */}
        <div className="flex items-center gap-1.5 text-[15px] leading-5 truncate">
          <Link
            href={`/${post.author_username}`}
            className="font-bold text-[#eff3f4] hover:underline truncate cursor-pointer"
          >
            {post.author_name}
          </Link>
          <Link
            href={`/${post.author_username}`}
            className="text-[#71767b] hover:underline truncate cursor-pointer"
          >
            @{post.author_username}
          </Link>
          {timeFormatted && (
            <>
              <span className="text-[#71767b]">·</span>
              <span className="text-[#71767b] shrink-0" title={new Date(post.created_at).toLocaleString()}>
                {timeFormatted}
              </span>
            </>
          )}
        </div>

        {/* Text-Only Post Body */}
        <p className="mt-1 text-[15px] leading-relaxed text-[#e7e9ea] break-words whitespace-pre-line">
          {post.content}
        </p>

        {/* Action Buttons Bar */}
        <div className="mt-3 flex items-center justify-between max-w-md text-[#71767b] text-xs">
          {/* Reply */}
          <button
            type="button"
            onClick={() => onReply?.(post.id)}
            className="flex items-center gap-2 group hover:text-[#1d9bf0] transition cursor-pointer p-1 -m-1 rounded-full"
            aria-label="Reply"
          >
            <div className="p-1.5 rounded-full group-hover:bg-[#1d9bf0]/10 transition">
              <svg className="w-4 h-4 fill-none stroke-current" viewBox="0 0 24 24" strokeWidth="1.8">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M12 20.25c4.97 0 9-3.694 9-8.25s-4.03-8.25-9-8.25S3 7.444 3 12c0 2.104.859 4.023 2.273 5.48.432.447.336 1.157-.098 1.636l-1.07 1.18c-.347.382-.047.989.467.923l2.87-.367a1.47 1.47 0 011.08.318C9.728 19.866 10.835 20.25 12 20.25z"
                />
              </svg>
            </div>
            <span>{post.replies > 0 ? post.replies : ""}</span>
          </button>

          {/* Retweet */}
          <button
            type="button"
            onClick={() => onRetweet?.(post.id)}
            className="flex items-center gap-2 group hover:text-[#00ba7c] transition cursor-pointer p-1 -m-1 rounded-full"
            aria-label="Repost"
          >
            <div className="p-1.5 rounded-full group-hover:bg-[#00ba7c]/10 transition">
              <svg className="w-4 h-4 fill-none stroke-current" viewBox="0 0 24 24" strokeWidth="1.8">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M19.5 12c0-1.232-.046-2.453-.138-3.662a4.006 4.006 0 00-3.7-3.7 48.678 48.678 0 00-7.324 0 4.006 4.006 0 00-3.7 3.7c-.017.228-.032.457-.045.686M4.5 12c0 1.232.046 2.453.138 3.662a4.006 4.006 0 003.7 3.7 48.656 48.656 0 007.324 0 4.006 4.006 0 003.7-3.7c.017-.228.032-.457.045-.686M7.5 4.5l-3 3.5 3 3.5M16.5 19.5l3-3.5-3-3.5"
                />
              </svg>
            </div>
            <span>{post.retweets > 0 ? post.retweets : ""}</span>
          </button>

          {/* Like */}
          <button
            type="button"
            onClick={() => onLike?.(post.id)}
            className="flex items-center gap-2 group hover:text-[#f4212e] transition cursor-pointer p-1 -m-1 rounded-full"
            aria-label="Like"
          >
            <div className="p-1.5 rounded-full group-hover:bg-[#f4212e]/10 transition">
              <svg className="w-4 h-4 fill-none stroke-current" viewBox="0 0 24 24" strokeWidth="1.8">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M21 8.25c0-2.485-2.099-4.5-4.688-4.5-1.935 0-3.597 1.126-4.312 2.733-.715-1.607-2.377-2.733-4.313-2.733C5.1 3.75 3 5.765 3 8.25c0 7.22 9 12 9 12s9-4.78 9-12z"
                />
              </svg>
            </div>
            <span>{post.likes > 0 ? post.likes : ""}</span>
          </button>

          {/* Share */}
          <button
            type="button"
            className="flex items-center gap-2 group hover:text-[#1d9bf0] transition cursor-pointer p-1 -m-1 rounded-full"
            aria-label="Share"
          >
            <div className="p-1.5 rounded-full group-hover:bg-[#1d9bf0]/10 transition">
              <svg className="w-4 h-4 fill-none stroke-current" viewBox="0 0 24 24" strokeWidth="1.8">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  d="M3 16.5v2.25A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75V16.5m-13.5-9L12 3m0 0l4.5 4.5M12 3v13.5"
                />
              </svg>
            </div>
          </button>
        </div>
      </div>
    </article>
  );
}
