"use client";

import React, { useState, useEffect, use } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import UserAvatar from "../../components/ui/UserAvatar";
import AccountPill from "../../components/navigation/AccountPill";
import EditProfileModal, { UserProfileData } from "../../components/profile/EditProfileModal";
import PostCard, { PostResponse } from "../../components/post/PostCard";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "";

interface PageProps {
  params: Promise<{ username: string }>;
}

export default function UserProfilePage({ params }: PageProps) {
  const router = useRouter();
  const resolvedParams = use(params);
  const targetUsername = decodeURIComponent(resolvedParams.username).replace(/^@/, "");

  // Authenticated user session
  const [currentUser, setCurrentUser] = useState<UserProfileData | null>(null);
  // Target profile user
  const [profileUser, setProfileUser] = useState<UserProfileData | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [userNotFound, setUserNotFound] = useState<boolean>(false);

  // Profile posts
  const [userPosts, setUserPosts] = useState<PostResponse[]>([]);
  const [isLoadingPosts, setIsLoadingPosts] = useState<boolean>(true);

  // Profile Edit Modal State
  const [isEditModalOpen, setIsEditModalOpen] = useState<boolean>(false);
  const [activeTab, setActiveTab] = useState<string>("Posts");

  // Fetch current authenticated session
  useEffect(() => {
    const fetchSession = () => {
      fetch(`${API_BASE}/api/auth/me`, { credentials: "include" })
        .then((res) => (res.ok ? res.json() : null))
        .then((data) => {
          if (data?.status === "success" && data.data?.user) {
            setCurrentUser(data.data.user);
          } else {
            setCurrentUser(null);
          }
        })
        .catch(() => {
          setCurrentUser(null);
        });
    };

    fetchSession();

    const handleFocus = () => {
      fetchSession();
    };

    window.addEventListener("focus", handleFocus);
    return () => {
      window.removeEventListener("focus", handleFocus);
    };
  }, []);

  // Fetch target profile by dynamic username
  useEffect(() => {
    let isCancelled = false;

    fetch(`${API_BASE}/api/users/profile?username=${targetUsername}`, {
      credentials: "include",
    })
      .then(async (res) => {
        if (!res.ok) {
          throw new Error("User not found");
        }
        return res.json();
      })
      .then((data) => {
        if (!isCancelled) {
          setIsLoading(false);
          if (data.status === "success" && data.data?.user) {
            setProfileUser(data.data.user);
            setUserNotFound(false);
          } else {
            setUserNotFound(true);
          }
        }
      })
      .catch(() => {
        if (!isCancelled) {
          setIsLoading(false);
          setUserNotFound(true);
        }
      });

    return () => {
      isCancelled = true;
    };
  }, [targetUsername]);

  // Fetch posts created by target user
  useEffect(() => {
    let isCancelled = false;

    fetch(`${API_BASE}/api/posts/user?username=${targetUsername}`, {
      credentials: "include",
    })
      .then(async (res) => {
        if (!res.ok) {
          throw new Error("Failed to load user posts");
        }
        return res.json();
      })
      .then((data) => {
        if (!isCancelled) {
          if (data.status === "success" && Array.isArray(data.data?.posts)) {
            setUserPosts(data.data.posts);
          } else {
            setUserPosts([]);
          }
          setIsLoadingPosts(false);
        }
      })
      .catch(() => {
        if (!isCancelled) {
          setUserPosts([]);
          setIsLoadingPosts(false);
        }
      });

    return () => {
      isCancelled = true;
    };
  }, [targetUsername]);

  const handleLogout = async () => {
    try {
      await fetch(`${API_BASE}/api/auth/logout`, {
        method: "POST",
        credentials: "include",
      });
    } catch {
      // Continue
    }
    setCurrentUser(null);
    router.push("/login");
  };

  const isOwnProfile = Boolean(
    currentUser &&
      profileUser &&
      (currentUser.id === profileUser.id ||
        currentUser.username.toLowerCase() === profileUser.username.toLowerCase())
  );

  return (
    <div className="min-h-screen bg-black text-[#e7e9ea] flex justify-center selection:bg-[#1d9bf0]/30 selection:text-white">
      <div className="w-full max-w-[1200px] flex">

        {/* Left Sidebar */}
        <aside className="w-[88px] xl:w-[275px] shrink-0 border-r border-[#2f3336] p-3 flex flex-col justify-between h-screen sticky top-0">
          <div className="space-y-4">
            {/* Logo */}
            <Link
              href="/dashboard"
              className="w-12 h-12 rounded-full flex items-center justify-center hover:bg-[#181818] transition text-[#eff3f4]"
            >
              <svg viewBox="0 0 24 24" aria-hidden="true" className="h-7 w-7 fill-current">
                <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z" />
              </svg>
            </Link>

            {/* Navigation links */}
            <nav className="space-y-1">
              <Link
                href="/dashboard"
                className="flex items-center gap-4 p-3 rounded-full hover:bg-[#181818] text-lg cursor-pointer text-[#eff3f4]"
              >
                <span>🏠</span>
                <span className="hidden xl:inline">Home</span>
              </Link>
              <div className="flex items-center gap-4 p-3 rounded-full hover:bg-[#181818] text-lg cursor-pointer text-[#71767b]">
                <span>🔍</span>
                <span className="hidden xl:inline">Explore</span>
              </div>
              <div className="flex items-center gap-4 p-3 rounded-full hover:bg-[#181818] text-lg cursor-pointer text-[#71767b]">
                <span>🔔</span>
                <span className="hidden xl:inline">Notifications</span>
              </div>
              {currentUser && (
                <Link
                  href={`/${currentUser.username}`}
                  className="flex items-center gap-4 p-3 rounded-full hover:bg-[#181818] font-bold text-lg cursor-pointer text-[#eff3f4]"
                >
                  <span>👤</span>
                  <span className="hidden xl:inline">Profile</span>
                </Link>
              )}
            </nav>
          </div>

          {/* Left Sidebar Account Pill */}
          <div className="pt-4">
            {currentUser ? (
              <AccountPill user={currentUser} onLogout={handleLogout} />
            ) : (
              <Link
                href="/login"
                className="w-full py-2 bg-white text-black font-bold text-center rounded-full block text-sm hover:bg-[#d7dbdc] transition"
              >
                Sign in
              </Link>
            )}
          </div>
        </aside>

        {/* Center Main Profile Column */}
        <main className="w-full max-w-[600px] border-r border-[#2f3336] min-h-screen">
          {/* Sticky Top Header */}
          <header className="sticky top-0 bg-black/80 backdrop-blur-md border-b border-[#2f3336] z-10 flex items-center gap-6 px-4 py-1.5">
            <button
              type="button"
              onClick={() => router.back()}
              className="w-9 h-9 rounded-full flex items-center justify-center hover:bg-[#181818] transition text-[#eff3f4]"
              aria-label="Back"
            >
              <svg viewBox="0 0 24 24" className="w-5 h-5 fill-current">
                <path d="M20 11H7.414l4.293-4.293a1 1 0 00-1.414-1.414l-6 6a1 1 0 000 1.414l6 6a1 1 0 001.414-1.414L7.414 13H20a1 1 0 000-2z" />
              </svg>
            </button>
            <div className="leading-tight">
              <h1 className="text-xl font-bold text-[#eff3f4] flex items-center gap-1">
                {profileUser ? profileUser.name : `@${targetUsername}`}
                {profileUser?.is_verified && (
                  <span className="text-[#1d9bf0] text-sm" title="Verified">
                    ✓
                  </span>
                )}
              </h1>
              <p className="text-xs text-[#71767b]">
                {profileUser ? "12 Posts" : "User Profile"}
              </p>
            </div>
          </header>

          {isLoading ? (
            <div className="py-24 flex items-center justify-center text-[#1d9bf0]">
              <div className="w-8 h-8 border-2 border-current border-t-transparent rounded-full animate-spin" />
            </div>
          ) : userNotFound ? (
            /* Authentic X User Not Found / 404 state */
            <div className="p-8 text-center space-y-4">
              <div className="h-44 bg-[#333639]" />
              <div className="px-4 -mt-16 mb-4 flex justify-between items-end">
                <UserAvatar size="xl" className="border-4 border-black" />
              </div>
              <div className="px-4 text-left">
                <h2 className="text-2xl font-extrabold text-[#eff3f4]">@{targetUsername}</h2>
                <div className="mt-8 text-center py-10">
                  <h3 className="text-3xl font-extrabold text-[#eff3f4] mb-2">This account doesn’t exist</h3>
                  <p className="text-[#71767b] text-[15px]">Try searching for another account.</p>
                </div>
              </div>
            </div>
          ) : profileUser ? (
            <div>
              {/* Cover Banner */}
              <div className="h-48 sm:h-52 bg-[#333639] relative">
                <div className="absolute inset-0 bg-gradient-to-b from-transparent to-black/40" />
              </div>

              {/* Avatar & Action Button Row */}
              <div className="px-4 -mt-16 sm:-mt-20 mb-3 flex justify-between items-end relative z-0">
                <UserAvatar
                  src={profileUser.avatar_url}
                  name={profileUser.name}
                  size="xl"
                  className="border-4 border-black shadow-2xl ring-1 ring-white/10"
                />

                <div>
                  {isOwnProfile ? (
                    <button
                      type="button"
                      id="edit-profile-trigger-btn"
                      onClick={() => setIsEditModalOpen(true)}
                      className="px-4 py-1.5 rounded-full border border-[#536471] font-bold text-sm text-[#eff3f4] hover:bg-white/10 transition cursor-pointer"
                    >
                      Edit profile
                    </button>
                  ) : (
                    <button
                      type="button"
                      className="px-5 py-1.5 rounded-full bg-[#eff3f4] text-black font-bold text-sm hover:bg-[#d7dbdc] transition cursor-pointer"
                    >
                      Follow
                    </button>
                  )}
                </div>
              </div>

              {/* User Bio & Metadata */}
              <div className="px-4 space-y-3">
                <div>
                  <h2 className="text-xl font-extrabold text-[#eff3f4] flex items-center gap-1.5">
                    <span>{profileUser.name}</span>
                    {profileUser.is_verified && (
                      <svg viewBox="0 0 24 24" className="w-5 h-5 fill-[#1d9bf0]">
                        <path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z" />
                      </svg>
                    )}
                  </h2>
                  <p className="text-sm text-[#71767b]">@{profileUser.username}</p>
                </div>

                {profileUser.bio && (
                  <p className="text-[15px] leading-relaxed text-[#e7e9ea] whitespace-pre-line">
                    {profileUser.bio}
                  </p>
                )}

                {/* Metadata badges */}
                <div className="flex flex-wrap items-center gap-x-4 gap-y-1 text-sm text-[#71767b]">
                  <div className="flex items-center gap-1">
                    <span>📅</span>
                    <span>Joined October 2026</span>
                  </div>
                  {profileUser.birth_date && (
                    <div className="flex items-center gap-1">
                      <span>🎈</span>
                      <span>Born {profileUser.birth_date}</span>
                    </div>
                  )}
                </div>

                {/* Follower metrics */}
                <div className="flex items-center gap-5 text-sm pt-1">
                  <div className="hover:underline cursor-pointer">
                    <span className="font-bold text-[#eff3f4]">85</span>{" "}
                    <span className="text-[#71767b]">Following</span>
                  </div>
                  <div className="hover:underline cursor-pointer">
                    <span className="font-bold text-[#eff3f4]">240</span>{" "}
                    <span className="text-[#71767b]">Followers</span>
                  </div>
                </div>
              </div>

              {/* Profile Sub-tabs */}
              <div className="flex border-b border-[#2f3336] mt-4">
                {["Posts", "Replies", "Highlights", "Media", "Likes"].map((tab) => (
                  <button
                    key={tab}
                    type="button"
                    onClick={() => setActiveTab(tab)}
                    className={`flex-1 py-3.5 text-center text-sm font-semibold transition cursor-pointer relative ${
                      activeTab === tab
                        ? "text-[#eff3f4]"
                        : "text-[#71767b] hover:bg-[#181818]"
                    }`}
                  >
                    {tab}
                    {activeTab === tab && (
                      <span className="absolute bottom-0 left-1/2 -translate-x-1/2 w-12 h-1 bg-[#1d9bf0] rounded-full" />
                    )}
                  </button>
                ))}
              </div>

              {/* Profile Feed Content */}
              <div>
                {activeTab === "Posts" ? (
                  isLoadingPosts ? (
                    <div className="py-12 flex justify-center items-center">
                      <div className="w-6 h-6 border-2 border-[#1d9bf0] border-t-transparent rounded-full animate-spin" />
                    </div>
                  ) : userPosts.length === 0 ? (
                    <div className="py-12 px-6 text-center">
                      <p className="text-xl font-bold text-[#eff3f4]">
                        @{profileUser.username} hasn&apos;t posted yet
                      </p>
                      <p className="text-sm text-[#71767b] mt-1">
                        When they post, their posts will show up here.
                      </p>
                    </div>
                  ) : (
                    userPosts.map((post) => (
                      <PostCard key={post.id} post={post} />
                    ))
                  )
                ) : (
                  <div className="py-12 px-6 text-center text-[#71767b]">
                    <p className="text-sm">Nothing to see here yet in {activeTab}.</p>
                  </div>
                )}
              </div>
            </div>
          ) : null}
        </main>

        {/* Right Widget Column */}
        <aside className="w-[350px] shrink-0 p-4 hidden lg:block">
          <div className="bg-[#16181c] rounded-2xl p-4 border border-[#2f3336]">
            <h2 className="text-lg font-extrabold text-[#eff3f4] mb-3">You might like</h2>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <UserAvatar
                    src="https://abs.twimg.com/responsive-web/client-web/icon-default.ee25327a.png"
                    name="X"
                    size="md"
                  />
                  <div className="leading-tight">
                    <p className="font-bold text-[#eff3f4] text-sm">X</p>
                    <p className="text-xs text-[#71767b]">@x</p>
                  </div>
                </div>
                <Link
                  href="/x"
                  className="px-4 py-1.5 rounded-full bg-white text-black font-bold text-xs hover:bg-[#d7dbdc] transition"
                >
                  Follow
                </Link>
              </div>
            </div>
          </div>
        </aside>

      </div>

      {/* Edit Profile Modal */}
      {isOwnProfile && profileUser && (
        <EditProfileModal
          isOpen={isEditModalOpen}
          user={profileUser}
          onClose={() => setIsEditModalOpen(false)}
          onUpdated={(updated) => {
            setProfileUser(updated);
            setCurrentUser(updated);
          }}
        />
      )}
    </div>
  );
}
