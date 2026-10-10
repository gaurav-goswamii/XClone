"use client";

import React, { useState } from "react";
import UserAvatar from "../ui/UserAvatar";
import AccountPopover, { UserSummary } from "./AccountPopover";

interface AccountPillProps {
  user: UserSummary;
  onLogout: () => void;
}

export default function AccountPill({ user, onLogout }: AccountPillProps) {
  const [isPopoverOpen, setIsPopoverOpen] = useState(false);

  return (
    <div className="relative w-full">
      <AccountPopover
        user={user}
        isOpen={isPopoverOpen}
        onClose={() => setIsPopoverOpen(false)}
        onLogout={onLogout}
      />

      <button
        type="button"
        id="account-pill-button"
        onClick={() => setIsPopoverOpen((prev) => !prev)}
        className="w-full flex items-center justify-between p-2 xl:p-2.5 rounded-full hover:bg-[#181818] transition duration-150 cursor-pointer text-left focus:outline-none focus-visible:ring-2 focus-visible:ring-[#1d9bf0]"
        aria-label="Account options"
        aria-expanded={isPopoverOpen}
      >
        <div className="flex items-center gap-3 min-w-0">
          <UserAvatar
            src={user.avatar_url}
            name={user.name}
            size="md"
            className="ring-1 ring-white/10"
          />
          <div className="hidden xl:block min-w-0 leading-tight">
            <p className="font-bold text-[#eff3f4] text-[15px] truncate max-w-[125px]">
              {user.name}
            </p>
            <p className="text-[14px] text-[#71767b] truncate max-w-[125px]">
              @{user.username}
            </p>
          </div>
        </div>

        {/* Three dots icon (Horizontal Ellipsis) */}
        <div className="hidden xl:flex items-center justify-center text-[#eff3f4] shrink-0 p-1">
          <svg viewBox="0 0 24 24" aria-hidden="true" className="w-5 h-5 fill-current">
            <path d="M3 12c0-1.1.9-2 2-2s2 .9 2 2-.9 2-2 2-2-.9-2-2zm9 2c1.1 0 2-.9 2-2s-.9-2-2-2-2 .9-2 2 .9 2 2 2zm7 0c1.1 0 2-.9 2-2s-.9-2-2-2-2 .9-2 2 .9 2 2 2z" />
          </svg>
        </div>
      </button>
    </div>
  );
}
