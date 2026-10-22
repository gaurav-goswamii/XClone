"use client";

import React, { useState } from "react";
import { PostResponse } from "./PostCard";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "";
const CHAR_LIMIT = 280;

interface EditPostModalProps {
  isOpen: boolean;
  post: PostResponse;
  onClose: () => void;
  onUpdated: (updatedPost: PostResponse) => void;
}

export default function EditPostModal({
  isOpen,
  post,
  onClose,
  onUpdated,
}: EditPostModalProps) {
  const [content, setContent] = useState(post.content);
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  if (!isOpen) return null;

  // Spreading counts characters the way Go counts runes, so an emoji costs 1 here too
  const charCount = [...content].length;
  const trimmedContent = content.trim();
  const isUnchanged = trimmedContent === post.content.trim();
  // A post carrying photos is allowed to have no caption, so text is only
  // required when there are no photos. Without this, a photo-only post opens a
  // modal whose Save button can never enable. Mirrors the same rule in
  // PostService.UpdatePost.
  const hasPhotos = (post.media?.length ?? 0) > 0;
  const canSave =
    (trimmedContent.length > 0 || hasPhotos) &&
    charCount <= CHAR_LIMIT &&
    !isSaving &&
    !isUnchanged;

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!canSave) return;

    setIsSaving(true);
    setError(null);

    try {
      const res = await fetch(`${API_BASE}/api/posts/${post.id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ content: trimmedContent }),
      });

      const data = await res.json();
      setIsSaving(false);

      if (res.ok && data.status === "success" && data.data?.post) {
        onUpdated(data.data.post);
        onClose();
      } else {
        setError(data.message || "Failed to save changes");
      }
    } catch {
      setIsSaving(false);
      setError("Network error while saving your post");
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-in fade-in duration-150">
      <form
        id="edit-post-modal"
        onSubmit={handleSave}
        className="w-full max-w-[600px] bg-black border border-[#2f3336] rounded-2xl overflow-hidden shadow-2xl flex flex-col max-h-[90vh]"
      >
        {/* Modal Header */}
        <div className="flex items-center justify-between px-4 py-3 border-b border-[#2f3336]">
          <div className="flex items-center gap-6">
            <button
              type="button"
              onClick={onClose}
              className="w-9 h-9 rounded-full flex items-center justify-center hover:bg-[#181818] transition text-[#eff3f4] cursor-pointer"
              aria-label="Close"
            >
              <svg viewBox="0 0 24 24" className="w-5 h-5 fill-current">
                <path d="M10.59 12L4.54 5.96l1.42-1.42L12 10.59l6.04-6.05 1.42 1.42L13.41 12l6.05 6.04-1.42 1.42L12 13.41 5.96 19.46 4.54 18.04 10.59 12z" />
              </svg>
            </button>
            <h2 className="text-xl font-bold text-[#eff3f4]">Edit post</h2>
          </div>
          <button
            type="submit"
            disabled={!canSave}
            className="px-4 py-1.5 rounded-full bg-[#1d9bf0] text-white font-bold text-sm hover:bg-[#1a8cd8] disabled:opacity-50 disabled:cursor-not-allowed transition cursor-pointer"
          >
            {isSaving ? "Saving..." : "Save"}
          </button>
        </div>

        {/* Modal Body */}
        <div className="overflow-y-auto flex-1 px-4 py-4 space-y-4">
          {error && (
            <div className="p-3 bg-[#f4212e]/10 border border-[#f4212e]/30 rounded-xl text-[#f4212e] text-sm">
              {error}
            </div>
          )}

          <textarea
            id="edit-post-content"
            value={content}
            onChange={(e) => {
              setContent(e.target.value);
              if (error) setError(null);
            }}
            disabled={isSaving}
            rows={5}
            autoFocus
            placeholder="What is happening?!"
            className="w-full bg-transparent text-[#e7e9ea] text-lg placeholder-[#71767b] outline-none resize-none disabled:opacity-60"
          />

          <div className="flex items-center justify-between border-t border-[#2f3336] pt-3">
            <span
              className={`text-xs ${charCount > CHAR_LIMIT ? "text-[#f4212e]" : "text-[#71767b]"}`}
            >
              {charCount} / {CHAR_LIMIT}
            </span>
            {isUnchanged && !isSaving && (
              <span className="text-xs text-[#71767b]">No changes yet</span>
            )}
          </div>
        </div>
      </form>
    </div>
  );
}
