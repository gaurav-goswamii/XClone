"use client";

import React, { useEffect, useRef, useState } from "react";
import { PostResponse } from "./PostCard";
import {
  DraftPhoto,
  MAX_PHOTOS,
  rejectPhotoFile,
  toMediaPayload,
  uploadPostPhoto,
} from "../../lib/postPhotos";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "";
const CHAR_LIMIT = 280;

// Callers mount this modal only while it is open, which is what resets a
// cancelled edit: the draft lives in component state, so unmounting discards
// it and the next opening seeds fresh values from the post again.
interface EditPostModalProps {
  isOpen: boolean;
  post: PostResponse;
  onClose: () => void;
  onUpdated: (updatedPost: PostResponse) => void;
}

// initialPhotos seeds the editor with the photos the post already has. They
// are "done" from the start: they live on the server, so their public IDs can
// be sent straight back. The public ID doubles as the key because it is unique
// within a post.
function initialPhotos(post: PostResponse): DraftPhoto[] {
  return (post.media ?? []).map((media) => ({
    key: media.public_id,
    previewUrl: media.url,
    status: "done" as const,
    public_id: media.public_id,
    width: media.width,
    height: media.height,
  }));
}

export default function EditPostModal({
  isOpen,
  post,
  onClose,
  onUpdated,
}: EditPostModalProps) {
  const [content, setContent] = useState(post.content);
  const [photos, setPhotos] = useState<DraftPhoto[]>(() => initialPhotos(post));
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  // Mirrors photos so the cleanup paths below can see the current list without
  // taking it as a dependency and re-running on every edit
  const photosRef = useRef<DraftPhoto[]>([]);

  useEffect(() => {
    photosRef.current = photos;
  }, [photos]);

  // Object URLs for newly picked files must be revoked by hand or they leak
  // for the page's lifetime
  useEffect(() => {
    return () => {
     
    };
  }, []);

  if (!isOpen) return null;

  // Spreading counts characters the way Go counts runes, so an emoji costs 1 here too
  const charCount = [...content].length;
  const trimmedContent = content.trim();

  const readyPhotos = photos.filter(
    (photo) => photo.status === "done" && photo.public_id
  );
  const isUploading = photos.some((photo) => photo.status === "uploading");

  // Joining the IDs compares the set *and* its order, so reordering photos
  // counts as a change the same way adding or removing one does
  const originalPhotoIds = (post.media ?? [])
    .map((media) => media.public_id)
    .join("|");
  const currentPhotoIds = readyPhotos.map((photo) => photo.public_id).join("|");

  const isUnchanged =
    trimmedContent === post.content.trim() &&
    currentPhotoIds === originalPhotoIds;

  // A post may be text, photos, or both, but it has to carry something.
  // Mirrors the same rule in PostService.UpdatePost.
  const canSave =
    (trimmedContent.length > 0 || readyPhotos.length > 0) &&
    charCount <= CHAR_LIMIT &&
    !isSaving &&
    !isUploading &&
    !isUnchanged;

  const handleFilesSelected = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const picked = Array.from(e.target.files ?? []);
    // Always reset the input, even on an early return, so picking the same file
    // again still fires onChange
    e.target.value = "";
    if (picked.length === 0) return;

    const remaining = MAX_PHOTOS - photos.length;
    if (remaining <= 0) {
      setError(`You can add up to ${MAX_PHOTOS} photos.`);
      return;
    }

    const accepted = picked.slice(0, remaining);
    if (picked.length > remaining) {
      setError(`You can add up to ${MAX_PHOTOS} photos.`);
    } else {
      setError(null);
    }

    const queued: { key: string; file: File }[] = [];
    for (const file of accepted) {
      const rejection = rejectPhotoFile(file);
      if (rejection) {
        setError(rejection);
        continue;
      }
      queued.push({ key: crypto.randomUUID(), file });
    }
    if (queued.length === 0) return;

    setPhotos((prev) => [
      ...prev,
      ...queued.map(({ key, file }) => ({
        key,
        previewUrl: URL.createObjectURL(file),
        status: "uploading" as const,
      })),
    ]);

    // Sequentially, one request in flight at a time: bounds memory and makes
    // failures easy to attribute. Display order comes from the array, not from
    // whichever upload finishes first.
    for (const { key, file } of queued) {
      const result = await uploadPostPhoto(file);
      setPhotos((prev) =>
        prev.map((photo) => {
          if (photo.key !== key) return photo;
          return result.ok
            ? {
                ...photo,
                status: "done" as const,
                error: undefined,
                public_id: result.public_id,
                width: result.width,
                height: result.height,
              }
            : { ...photo, status: "error" as const, error: result.error };
        })
      );
    }
  };

  const handleRemovePhoto = (key: string) => {
    setPhotos((prev) => prev.filter((photo) => photo.key !== key));
    setError(null);
  };

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
        // media is always sent, so an empty array genuinely means "no photos".
        // Omitting the field would tell the server to keep the existing ones,
        // which is the opposite of what removing them all should do.
        body: JSON.stringify({
          content: trimmedContent,
          media: toMediaPayload(photos),
        }),
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

          {photos.length > 0 && (
            <div className="grid grid-cols-2 gap-2 max-w-[400px]">
              {photos.map((photo, index) => (
                <div
                  key={photo.key}
                  className={`relative aspect-square rounded-xl overflow-hidden border ${
                    photo.status === "error"
                      ? "border-[#f4212e]"
                      : "border-[#2f3336]"
                  }`}
                >
                  {/* eslint-disable-next-line @next/next/no-img-element */}
                  <img
                    src={photo.previewUrl}
                    alt={`Photo ${index + 1} of ${photos.length}`}
                    className="w-full h-full object-cover"
                  />

                  {photo.status === "uploading" && (
                    <div className="absolute inset-0 bg-black/55 flex items-center justify-center">
                      <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin" />
                    </div>
                  )}

                  {photo.status === "error" && (
                    <div className="absolute inset-0 bg-[#f4212e]/25 flex items-end">
                      <p className="w-full px-2 py-1 text-[11px] leading-tight text-white bg-black/70">
                        {photo.error || "Upload failed"}
                      </p>
                    </div>
                  )}

                  <button
                    type="button"
                    onClick={() => handleRemovePhoto(photo.key)}
                    disabled={isSaving}
                    aria-label={`Remove photo ${index + 1}`}
                    className="absolute top-1 right-1 w-7 h-7 rounded-full bg-black/70 hover:bg-black/90 text-white flex items-center justify-center transition cursor-pointer disabled:opacity-50"
                  >
                    <svg
                      viewBox="0 0 24 24"
                      className="w-4 h-4 fill-none stroke-current"
                      strokeWidth="2.2"
                    >
                      <path strokeLinecap="round" d="M6 6l12 12M18 6L6 18" />
                    </svg>
                  </button>
                </div>
              ))}
            </div>
          )}

          <div className="flex items-center justify-between border-t border-[#2f3336] pt-3">
            <div className="flex items-center gap-3">
              {/* The narrow accept list also makes iOS transcode HEIC to JPEG */}
              <input
                ref={fileInputRef}
                type="file"
                multiple
                accept="image/jpeg,image/png,image/webp,image/gif"
                onChange={handleFilesSelected}
                className="hidden"
                id="edit-post-photo-input"
              />
              <button
                type="button"
                onClick={() => fileInputRef.current?.click()}
                disabled={isSaving || photos.length >= MAX_PHOTOS}
                aria-label="Add photos"
                title={
                  photos.length >= MAX_PHOTOS
                    ? `You can add up to ${MAX_PHOTOS} photos`
                    : "Add photos"
                }
                className="w-8 h-8 -ml-1 rounded-full flex items-center justify-center text-[#1d9bf0] hover:bg-[#1d9bf0]/10 transition cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed"
              >
                <svg
                  viewBox="0 0 24 24"
                  className="w-5 h-5 fill-none stroke-current"
                  strokeWidth="1.8"
                >
                  <rect x="3" y="3" width="18" height="18" rx="2" />
                  <circle cx="8.5" cy="8.5" r="1.6" />
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    d="M21 15l-5-5L5 21"
                  />
                </svg>
              </button>
              {photos.length > 0 && (
                <span className="text-xs text-[#71767b]">
                  {photos.length}/{MAX_PHOTOS}
                </span>
              )}
              <span
                className={`text-xs ${charCount > CHAR_LIMIT ? "text-[#f4212e]" : "text-[#71767b]"}`}
              >
                {charCount} / {CHAR_LIMIT}
              </span>
            </div>
            {isUploading ? (
              <span className="text-xs text-[#71767b]">Uploading photos...</span>
            ) : (
              isUnchanged &&
              !isSaving && (
                <span className="text-xs text-[#71767b]">No changes yet</span>
              )
            )}
          </div>
        </div>
      </form>
    </div>
  );
}
