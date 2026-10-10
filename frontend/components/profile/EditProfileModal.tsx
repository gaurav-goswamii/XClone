"use client";

import React, { useState, useRef } from "react";
import UserAvatar from "../ui/UserAvatar";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "";

export interface UserProfileData {
  id: string;
  name: string;
  username: string;
  email?: string;
  birth_date?: string;
  avatar_url?: string;
  bio?: string;
  is_verified?: boolean;
}

interface EditProfileModalProps {
  isOpen: boolean;
  user: UserProfileData;
  onClose: () => void;
  onUpdated: (updatedUser: UserProfileData) => void;
}

export default function EditProfileModal({
  isOpen,
  user,
  onClose,
  onUpdated,
}: EditProfileModalProps) {
  const [name, setName] = useState(user.name || "");
  const [bio, setBio] = useState(user.bio || "");
  const [avatarUrl, setAvatarUrl] = useState(user.avatar_url || "");
  const [isUploading, setIsUploading] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fileInputRef = useRef<HTMLInputElement>(null);

  if (!isOpen) return null;

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    // Validate size (max 5MB)
    if (file.size > 5 * 1024 * 1024) {
      setError("Image size cannot exceed 5MB");
      return;
    }

    // Validate type
    const validTypes = ["image/jpeg", "image/png", "image/webp", "image/gif"];
    if (!validTypes.includes(file.type)) {
      setError("Please select a JPG, PNG, WEBP, or GIF image");
      return;
    }

    setIsUploading(true);
    setError(null);

    // Immediate local preview
    const previewUrl = URL.createObjectURL(file);
    setAvatarUrl(previewUrl);

    try {
      const formData = new FormData();
      formData.append("avatar", file);

      const res = await fetch(`${API_BASE}/api/upload/avatar`, {
        method: "POST",
        credentials: "include",
        body: formData,
      });

      const data = await res.json();
      setIsUploading(false);

      if (res.ok && data.status === "success" && data.data?.avatar_url) {
        setAvatarUrl(data.data.avatar_url);
      } else {
        setError(data.message || "Failed to upload avatar to Cloudinary");
      }
    } catch {
      setIsUploading(false);
      setError("Network error while uploading avatar");
    }
  };

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim()) {
      setError("Name cannot be empty");
      return;
    }

    setIsSaving(true);
    setError(null);

    try {
      const res = await fetch(`${API_BASE}/api/users/profile`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({
          name: name.trim(),
          avatar_url: avatarUrl.trim(),
          bio: bio.trim(),
        }),
      });

      const data = await res.json();
      setIsSaving(false);

      if (res.ok && data.status === "success" && data.data?.user) {
        onUpdated(data.data.user);
        onClose();
      } else {
        setError(data.message || "Failed to save profile changes");
      }
    } catch {
      setIsSaving(false);
      setError("Network error while saving profile");
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-in fade-in duration-150">
      <div
        id="edit-profile-modal"
        className="w-full max-w-[600px] bg-black border border-[#2f3336] rounded-2xl overflow-hidden shadow-2xl flex flex-col max-h-[90vh]"
      >
        {/* Modal Header */}
        <div className="flex items-center justify-between px-4 py-3 border-b border-[#2f3336] sticky top-0 bg-black/80 backdrop-blur-md z-10">
          <div className="flex items-center gap-6">
            <button
              type="button"
              onClick={onClose}
              className="w-9 h-9 rounded-full flex items-center justify-center hover:bg-[#181818] transition text-[#eff3f4]"
              aria-label="Close"
            >
              <svg viewBox="0 0 24 24" className="w-5 h-5 fill-current">
                <path d="M10.59 12L4.54 5.96l1.42-1.42L12 10.59l6.04-6.05 1.42 1.42L13.41 12l6.05 6.04-1.42 1.42L12 13.41 5.96 19.46 4.54 18.04 10.59 12z" />
              </svg>
            </button>
            <h2 className="text-xl font-bold text-[#eff3f4]">Edit profile</h2>
          </div>
          <button
            type="button"
            onClick={handleSave}
            disabled={isSaving || isUploading || !name.trim()}
            className="px-4 py-1.5 rounded-full bg-white text-black font-bold text-sm hover:bg-[#d7dbdc] disabled:opacity-50 disabled:cursor-not-allowed transition cursor-pointer"
          >
            {isSaving ? "Saving..." : "Save"}
          </button>
        </div>

        {/* Modal Scrollable Body */}
        <div className="overflow-y-auto flex-1 pb-6">
          {/* Header Banner */}
          <div className="relative h-44 bg-[#333639] flex items-center justify-center">
            <div className="absolute inset-0 bg-black/20" />
            <button
              type="button"
              title="Add header photo"
              className="relative z-10 w-11 h-11 rounded-full bg-black/60 hover:bg-black/80 flex items-center justify-center text-white transition cursor-pointer"
            >
              <svg viewBox="0 0 24 24" className="w-5 h-5 fill-current">
                <path d="M19.708 5.688L16.292 2.272A1.99 1.99 0 0014.879 1.688H9.121c-.53 0-1.039.211-1.414.586L4.292 5.688A1.99 1.99 0 003.707 7.101V19c0 1.103.897 2 2 2h14c1.103 0 2-.897 2-2V7.101a1.99 1.99 0 00-.585-1.413zM12 17.5c-2.485 0-4.5-2.015-4.5-4.5s2.015-4.5 4.5-4.5 4.5 2.015 4.5 4.5-2.015 4.5-4.5 4.5z" />
              </svg>
            </button>
          </div>

          {/* Avatar Section (overlapping banner) */}
          <div className="px-4 -mt-16 mb-4 flex justify-between items-end relative z-20">
            <div className="relative group">
              <UserAvatar
                src={avatarUrl}
                name={name}
                size="xl"
                className="border-4 border-black ring-2 ring-[#2f3336]/60 shadow-xl"
              />

              {/* Camera Icon Overlay to trigger Cloudinary Upload */}
              <button
                type="button"
                onClick={() => fileInputRef.current?.click()}
                disabled={isUploading}
                title="Change profile picture (Cloudinary)"
                className="absolute inset-0 m-auto w-11 h-11 rounded-full bg-black/60 hover:bg-black/80 flex items-center justify-center text-white transition cursor-pointer z-30"
              >
                {isUploading ? (
                  <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin" />
                ) : (
                  <svg viewBox="0 0 24 24" className="w-5 h-5 fill-current">
                    <path d="M19.708 5.688L16.292 2.272A1.99 1.99 0 0014.879 1.688H9.121c-.53 0-1.039.211-1.414.586L4.292 5.688A1.99 1.99 0 003.707 7.101V19c0 1.103.897 2 2 2h14c1.103 0 2-.897 2-2V7.101a1.99 1.99 0 00-.585-1.413zM12 17.5c-2.485 0-4.5-2.015-4.5-4.5s2.015-4.5 4.5-4.5 4.5 2.015 4.5 4.5-2.015 4.5-4.5 4.5z" />
                  </svg>
                )}
              </button>

              <input
                ref={fileInputRef}
                type="file"
                accept="image/png, image/jpeg, image/webp, image/gif"
                onChange={handleFileChange}
                className="hidden"
                id="cloudinary-avatar-file-input"
              />
            </div>
          </div>

          {/* Form Fields */}
          <div className="px-4 space-y-5">
            {error && (
              <div className="p-3 bg-[#f4212e]/10 border border-[#f4212e]/30 rounded-xl text-[#f4212e] text-sm">
                {error}
              </div>
            )}

            {/* Display Name Input */}
            <div className="relative border border-[#333639] focus-within:border-[#1d9bf0] focus-within:ring-1 focus-within:ring-[#1d9bf0] rounded-lg p-3 transition">
              <div className="flex justify-between text-xs text-[#71767b]">
                <label htmlFor="edit-name">Name</label>
                <span>{name.length} / 50</span>
              </div>
              <input
                id="edit-name"
                type="text"
                maxLength={50}
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="Name"
                className="w-full bg-transparent text-[#eff3f4] text-base outline-none pt-1"
              />
            </div>

            {/* Bio Input */}
            <div className="relative border border-[#333639] focus-within:border-[#1d9bf0] focus-within:ring-1 focus-within:ring-[#1d9bf0] rounded-lg p-3 transition">
              <div className="flex justify-between text-xs text-[#71767b]">
                <label htmlFor="edit-bio">Bio</label>
                <span>{bio.length} / 160</span>
              </div>
              <textarea
                id="edit-bio"
                maxLength={160}
                rows={3}
                value={bio}
                onChange={(e) => setBio(e.target.value)}
                placeholder="Add your bio"
                className="w-full bg-transparent text-[#eff3f4] text-base outline-none pt-1 resize-none"
              />
            </div>

            {/* Cloudinary URL input for direct custom URL */}
            <div className="relative border border-[#333639] focus-within:border-[#1d9bf0] focus-within:ring-1 focus-within:ring-[#1d9bf0] rounded-lg p-3 transition">
              <div className="flex justify-between text-xs text-[#71767b]">
                <label htmlFor="edit-avatar-url">Profile Picture URL (Cloudinary or Direct)</label>
              </div>
              <input
                id="edit-avatar-url"
                type="text"
                value={avatarUrl}
                onChange={(e) => setAvatarUrl(e.target.value)}
                placeholder="https://res.cloudinary.com/..."
                className="w-full bg-transparent text-[#eff3f4] text-sm outline-none pt-1"
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
