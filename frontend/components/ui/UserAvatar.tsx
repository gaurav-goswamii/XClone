"use client";

import React, { useState } from "react";

export type AvatarSize = "xs" | "sm" | "md" | "lg" | "xl";

interface UserAvatarProps {
  src?: string | null;
  name?: string;
  size?: AvatarSize;
  className?: string;
  onClick?: () => void;
}

const sizeClasses: Record<AvatarSize, { container: string; svg: string }> = {
  xs: {
    container: "w-6 h-6 min-w-6 min-h-6 text-[10px]",
    svg: "w-3.5 h-3.5",
  },
  sm: {
    container: "w-8 h-8 min-w-8 min-h-8 text-xs",
    svg: "w-5 h-5",
  },
  md: {
    container: "w-10 h-10 min-w-10 min-h-10 text-sm",
    svg: "w-6 h-6",
  },
  lg: {
    container: "w-12 h-12 min-w-12 min-h-12 text-base",
    svg: "w-7 h-7",
  },
  xl: {
    container: "w-28 h-28 sm:w-36 sm:h-36 min-w-28 min-h-28 sm:min-w-36 sm:min-h-36 text-3xl",
    svg: "w-16 h-16 sm:w-20 sm:h-20",
  },
};

export default function UserAvatar({
  src,
  name = "User",
  size = "md",
  className = "",
  onClick,
}: UserAvatarProps) {
  const [hasError, setHasError] = useState(false);
  const sizeConfig = sizeClasses[size] || sizeClasses.md;

  const showImage = Boolean(src && !hasError && src.trim() !== "");

  return (
    <div
      onClick={onClick}
      className={`relative rounded-full overflow-hidden shrink-0 select-none bg-[#202327] flex items-center justify-center border border-[#2f3336]/60 ${sizeConfig.container} ${className}`}
    >
      {showImage ? (
        // eslint-disable-next-line @next/next/no-img-element
        <img
          src={src as string}
          alt={name}
          onError={() => setHasError(true)}
          className="w-full h-full object-cover rounded-full"
          loading="lazy"
        />
      ) : (
        // Authentic X / Twitter Default Geometric Silhouette SVG
        <svg
          viewBox="0 0 24 24"
          aria-hidden="true"
          className={`fill-[#71767b] ${sizeConfig.svg}`}
        >
          <path d="M12 12c2.21 0 4-1.79 4-4s-1.79-4-4-4-4 1.79-4 4 1.79 4 4 4zm0 2c-2.67 0-8 1.34-8 4v2h16v-2c0-2.66-5.33-4-8-4z" />
        </svg>
      )}
    </div>
  );
}
