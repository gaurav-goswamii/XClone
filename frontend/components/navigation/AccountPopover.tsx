"use client";

import React, { useEffect, useRef } from "react";
import Link from "next/link";
import UserAvatar from "../ui/UserAvatar";

export interface UserSummary {
  id: string;
  name: string;
  username: string;
  avatar_url?: string;
}

interface AccountPopoverProps {
  user: UserSummary;
  isOpen: boolean;
  onClose: () => void;
  onLogout: () => void;
}

export default function AccountPopover({
  user,
  isOpen,
  onClose,
  onLogout,
}: AccountPopoverProps) {
  const popoverRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!isOpen) return;

    const handleOutsideClick = (e: MouseEvent) => {
      if (popoverRef.current && !popoverRef.current.contains(e.target as Node)) {
        onClose();
      }
    };

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        onClose();
      }
    };

    document.addEventListener("mousedown", handleOutsideClick);
    document.addEventListener("keydown", handleKeyDown);

    return () => {
      document.removeEventListener("mousedown", handleOutsideClick);
      document.removeEventListener("keydown", handleKeyDown);
    };
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  return (
    <div
      ref={popoverRef}
      id="account-popover-menu"
      className="absolute bottom-full left-0 mb-3 w-[280px] xl:w-[300px] bg-black border border-[#2f3336] rounded-2xl shadow-[0_0_15px_rgba(255,255,255,0.18)] py-2 z-50 animate-in fade-in zoom-in-95 duration-100"
    >
      {/* Active account item -> Links to user's profile */}
      <Link
        href={`/${user.username}`}
        onClick={onClose}
        className="flex items-center justify-between px-4 py-3 hover:bg-[#181818] transition cursor-pointer"
      >
        <div className="flex items-center gap-3 min-w-0">
          <UserAvatar src={user.avatar_url} name={user.name} size="md" />
          <div className="min-w-0 text-left">
            <p className="font-bold text-[#eff3f4] text-[15px] truncate">
              {user.name}
            </p>
            <p className="text-[14px] text-[#71767b] truncate">
              @{user.username}
            </p>
          </div>
        </div>
        {/* Blue Checkmark (active account indicator) */}
        <div className="text-[#1d9bf0] shrink-0 ml-2" title="Active account">
          <svg viewBox="0 0 24 24" className="w-5 h-5 fill-current">
            <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z" />
          </svg>
        </div>
      </Link>

      <div className="h-[1px] bg-[#2f3336] my-1" />

      {/* Profile link item */}
      <Link
        href={`/${user.username}`}
        onClick={onClose}
        className="block px-4 py-3 text-left text-[15px] font-medium text-[#eff3f4] hover:bg-[#181818] transition"
      >
        View Profile
      </Link>

      {/* Add existing account */}
      <button
        type="button"
        onClick={() => {
          onClose();
          alert("Multi-account switcher: In this demo, one account is active.");
        }}
        className="w-full text-left px-4 py-3 text-[15px] font-medium text-[#eff3f4] hover:bg-[#181818] transition"
      >
        Add an existing account
      </button>

      {/* Logout option */}
      <button
        type="button"
        id="popover-logout-btn"
        onClick={() => {
          onClose();
          onLogout();
        }}
        className="w-full text-left px-4 py-3 text-[15px] font-medium text-[#eff3f4] hover:bg-[#181818] hover:text-[#f4212e] transition"
      >
        Log out @{user.username}
      </button>

      {/* Authentic bottom triangle arrow pointer */}
      <div className="absolute -bottom-2 left-6 xl:left-8 w-4 h-4 bg-black border-r border-b border-[#2f3336] transform rotate-45" />
    </div>
  );
}
