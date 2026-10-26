"use client";
import React, { useState, useEffect, useRef } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import UserAvatar from "../../components/ui/UserAvatar";
import AccountPill from "../../components/navigation/AccountPill";
import PostCard, { PostResponse } from "../../components/post/PostCard";
import EditPostModal from "../../components/post/EditPostModal";
import {
  DraftPhoto,
  MAX_PHOTOS,
  rejectPhotoFile,
  toMediaPayload,
  uploadPostPhoto,
} from "../../lib/postPhotos";

interface UserProfile {
  id: string;
  name: string;
  username: string;
  email: string;
  birth_date?: string;
  avatar_url?: string;
  bio?: string;
  created_at?: string;
}

// The `|| ""` matters: without it an unset NEXT_PUBLIC_API_URL makes every
// request URL start with the literal string "undefined". Every other file in
// the app already has this fallback.
const API_BASE = process.env.NEXT_PUBLIC_API_URL || "";
const CHAR_LIMIT = 280;

export default function DashboardPage() {
  const router = useRouter();

  const [user, setUser] = useState<UserProfile | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);

  const [tweetText, setTweetText] = useState<string>("");
  const [feedPosts, setFeedPosts] = useState<PostResponse[]>([]);
  const [isLoadingFeed, setIsLoadingFeed] = useState<boolean>(true);
  const [isPosting, setIsPosting] = useState<boolean>(false);
  const [postError, setPostError] = useState<string | null>(null);
  const [editingPost, setEditingPost] = useState<PostResponse | null>(null);
  const [feedError, setFeedError] = useState<string | null>(null);
  const [draftPhotos, setDraftPhotos] = useState<DraftPhoto[]>([]);
  const fileInputRef = useRef<HTMLInputElement>(null);
  // Object URLs must be revoked by hand or they leak for the page's lifetime.
  // A ref lets the unmount cleanup see the current list without re-running.
  const previewUrlsRef = useRef<string[]>([]);

  useEffect(() => {
    previewUrlsRef.current = draftPhotos.map((photo) => photo.previewUrl);
  }, [draftPhotos]);

  useEffect(() => {
    return () => {
      previewUrlsRef.current.forEach((url) => URL.revokeObjectURL(url));
    };
  }, []);

  useEffect(() => {
    let isMounted = true;

    const verifySession = async () => {
      try {
        const res = await fetch(`${API_BASE}/api/auth/me`, {
          credentials: "include",
        });
        if (!res.ok) {
          throw new Error("Unauthenticated");
        }
        const data = await res.json();
        if (isMounted) {
          if (data.status === "success" && data.data?.user) {
            setUser(data.data.user);
            setIsLoading(false);
          } else {
            setUser(null);
            router.replace("/login");
          }
        }
      } catch {
        if (isMounted) {
          setUser(null);
          router.replace("/login");
        }
      }
    };

    verifySession();

    // Re-verify session when the window gains focus (e.g., after cookie changes/deletion in DevTools)
    const handleFocus = () => {
      verifySession();
    };

    window.addEventListener("focus", handleFocus);
    return () => {
      isMounted = false;
      window.removeEventListener("focus", handleFocus);
    };
  }, [router]);

  // Fetch live timeline feed
  useEffect(() => {
    let isCancelled = false;

    fetch(`${API_BASE}/api/posts/feed`, {
      credentials: "include",
    })
      .then((res) => (res.ok ? res.json() : null))
      .then((data) => {
        if (!isCancelled && data?.status === "success" && Array.isArray(data.data?.posts)) {
          setFeedPosts(data.data.posts);
        }
      })
      .catch(() => {
        // Keep feed empty on error
      })
      .finally(() => {
        if (!isCancelled) {
          setIsLoadingFeed(false);
        }
      });

    return () => {
      isCancelled = true;
    };
  }, []);

  const handleLogout = async () => {
    // Clear backend session cookie
    try {
      await fetch(`${API_BASE}/api/auth/logout`, {
        method: "POST",
        credentials: "include",
      });
    } catch {
      // Continue client logout even on network error
    }

    // Clear user state and navigate to login
    setUser(null);
    router.replace("/login");
  };

  // Upload one photo and patch its row in place. Rows are addressed by key, so
  // a removal mid-flight simply means the patch finds nothing to update.
  const uploadDraftPhoto = async (key: string, file: File) => {
    const result = await uploadPostPhoto(file);

    setDraftPhotos((prev) =>
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
  };

  const handleFilesSelected = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const picked = Array.from(e.target.files ?? []);
    // Always reset the input, even on an early return, so picking the same file
    // again still fires onChange
    e.target.value = "";
    if (picked.length === 0) return;

    const remaining = MAX_PHOTOS - draftPhotos.length;
    if (remaining <= 0) {
      setPostError(`You can add up to ${MAX_PHOTOS} photos.`);
      return;
    }

    const accepted = picked.slice(0, remaining);
    if (picked.length > remaining) {
      setPostError(`You can add up to ${MAX_PHOTOS} photos.`);
    } else {
      setPostError(null);
    }

    const queued: { key: string; file: File }[] = [];
    for (const file of accepted) {
      const rejection = rejectPhotoFile(file);
      if (rejection) {
        setPostError(rejection);
        continue;
      }
      queued.push({ key: crypto.randomUUID(), file });
    }
    if (queued.length === 0) return;

    setDraftPhotos((prev) => [
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
      await uploadDraftPhoto(key, file);
    }
  };

  const handleRemoveDraftPhoto = (key: string) => {
    setDraftPhotos((prev) => {
      const target = prev.find((photo) => photo.key === key);
      if (target) URL.revokeObjectURL(target.previewUrl);
      return prev.filter((photo) => photo.key !== key);
    });
    setPostError(null);
  };

  const handlePostTweet = async (e: React.FormEvent) => {
    e.preventDefault();
    const content = tweetText.trim();
    const readyPhotos = draftPhotos.filter(
      (photo) => photo.status === "done" && photo.public_id
    );

    if (isPosting) return;
    // Posting mid-upload would silently drop the photos still in flight
    if (draftPhotos.some((photo) => photo.status === "uploading")) return;
    // A post needs text or photos, but no longer both
    if (!content && readyPhotos.length === 0) return;
    // Spread counts code points, matching the server's rune count
    if ([...content].length > CHAR_LIMIT) return;

    setIsPosting(true);
    setPostError(null);

    try {
      const res = await fetch(`${API_BASE}/api/posts`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        credentials: "include",
        body: JSON.stringify({
          content,
          media: toMediaPayload(readyPhotos),
        }),
      });

      const data = await res.json();

      if (!res.ok || data.status !== "success") {
        setPostError(data?.message || "Failed to post. Please try again.");
        return;
      }

      if (data.data?.post) {
        setFeedPosts((prev) => [data.data.post, ...prev]);
        setTweetText("");
        draftPhotos.forEach((photo) => URL.revokeObjectURL(photo.previewUrl));
        setDraftPhotos([]);
      }
    } catch {
      setPostError("Network error. Could not post tweet.");
    } finally {
      setIsPosting(false);
    }
  };

  // Replace the edited post in place, keeping its position in the feed
  const handlePostUpdated = (updatedPost: PostResponse) => {
    setFeedPosts((prev) =>
      prev.map((post) => (post.id === updatedPost.id ? updatedPost : post))
    //if the id of updated post = prev post id, then replace it
    );
  };

  const handleDeletePost = async (postId: string) => {
    if (!window.confirm("Delete this post? This cannot be undone.")) return;

    setFeedError(null);

    try {
      const res = await fetch(`${API_BASE}/api/posts/${postId}`, {
        method: "DELETE",
        credentials: "include",
      });

      // 404 means it is already gone, so drop it from the list either way
      if (res.ok || res.status === 404) {
        setFeedPosts((prev) => prev.filter((post) => post.id !== postId));
        return;
      }

      const data = await res.json().catch(() => null);
      setFeedError(data?.message || "Failed to delete post. Please try again.");
    } catch {
      setFeedError("Network error. Could not delete post.");
    }
  };

  if (isLoading || !user) {
    return (
      <div className="min-h-screen bg-black flex flex-col items-center justify-center text-[#1d9bf0] gap-3">
        <div className="w-8 h-8 border-2 border-current border-t-transparent rounded-full animate-spin" />
        <span className="text-xs text-[#71767b] font-medium">Loading timeline...</span>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-black text-[#e7e9ea] flex justify-center">
      <div className="w-full max-w-[1200px] flex">
        
        {/* Left Sidebar */}
        <aside className="w-[88px] xl:w-[275px] shrink-0 border-r border-[#2f3336] p-3 flex flex-col justify-between h-screen sticky top-0">
          <div className="space-y-4">
            {/* Logo */}
            <Link
              href="/"
              className="w-12 h-12 rounded-full flex items-center justify-center hover:bg-[#181818] transition text-[#eff3f4]"
            >
              <svg viewBox="0 0 24 24" aria-hidden="true" className="h-7 w-7 fill-current">
                <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z" />
              </svg>
            </Link>

            {/* Navigation links */}
            <nav className="space-y-1">
              <div className="flex items-center gap-4 p-3 rounded-full hover:bg-[#181818] font-bold text-lg cursor-pointer text-[#eff3f4]">
                <span>🏠</span>
                <span className="hidden xl:inline">Home</span>
              </div>
              <div className="flex items-center gap-4 p-3 rounded-full hover:bg-[#181818] text-lg cursor-pointer text-[#71767b]">
                <span>🔍</span>
                <span className="hidden xl:inline">Explore</span>
              </div>
              <div className="flex items-center gap-4 p-3 rounded-full hover:bg-[#181818] text-lg cursor-pointer text-[#71767b]">
                <span>🔔</span>
                <span className="hidden xl:inline">Notifications</span>
              </div>
              <Link
                href={`/${user.username}`}
                className="flex items-center gap-4 p-3 rounded-full hover:bg-[#181818] font-bold text-lg cursor-pointer text-[#eff3f4]"
              >
                <span>👤</span>
                <span className="hidden xl:inline">Profile</span>
              </Link>
            </nav>
          </div>

          {/* Left Sidebar Account Pill */}
          <div className="pt-4">
            <AccountPill user={user} onLogout={handleLogout} />
          </div>
        </aside>

        {/* Center Feed Column */}
        <main className="w-full max-w-[600px] border-r border-[#2f3336] min-h-screen">
          {/* Header Tabs */}
          <header className="sticky top-0 bg-black/80 backdrop-blur-md border-b border-[#2f3336] z-10">
            <div className="flex items-center justify-between px-4 py-3">
              <h1 className="text-xl font-bold text-[#eff3f4]">Home</h1>
              <button
                type="button"
                onClick={handleLogout}
                id="header-logout-btn"
                title="Log out of account"
                className="flex items-center gap-1.5 px-3 py-1.5 rounded-full border border-[#333639] hover:border-[#f4212e]/50 hover:bg-[#f4212e]/10 text-[#71767b] hover:text-[#f4212e] text-xs font-semibold transition duration-150 cursor-pointer"
              >
                <svg className="w-4 h-4 shrink-0 fill-none stroke-current" viewBox="0 0 24 24" strokeWidth="2">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 9V5.25A2.25 2.25 0 0013.5 3h-6a2.25 2.25 0 00-2.25 2.25v13.5A2.25 2.25 0 007.5 21h6a2.25 2.25 0 002.25-2.25V15m3 0l3-3m0 0l-3-3m3 3H9" />
                </svg>
                <span>Log out</span>
              </button>
            </div>
            <div className="flex border-t border-[#2f3336]">
              <div className="flex-1 py-3 text-center font-bold border-b-4 border-[#1d9bf0] text-[#eff3f4] cursor-pointer">
                For you
              </div>
              <div className="flex-1 py-3 text-center text-[#71767b] hover:bg-[#181818] cursor-pointer transition">
                Following
              </div>
            </div>
          </header>

          {/* Compose Post Box */}
          <form onSubmit={handlePostTweet} className="p-4 border-b border-[#2f3336] flex gap-3">
            <Link href={`/${user.username}`} className="shrink-0">
              <UserAvatar
                src={user.avatar_url}
                name={user.name}
                size="md"
                className="hover:opacity-90 transition cursor-pointer"
              />
            </Link>
            <div className="flex-1 min-w-0">
              <textarea
                value={tweetText}
                onChange={(e) => {
                  setTweetText(e.target.value);
                  if (postError) setPostError(null);
                }}
                placeholder="What is happening?!"
                maxLength={CHAR_LIMIT}
                disabled={isPosting}
                className="w-full bg-transparent text-[18px] text-[#e7e9ea] outline-none placeholder-[#71767b] resize-none min-h-[60px] disabled:opacity-50"
              />

              {draftPhotos.length > 0 && (
                <div className="mt-2 mb-3 grid grid-cols-2 gap-2 max-w-[400px]">
                  {draftPhotos.map((photo, index) => (
                    <div
                      key={photo.key}
                      className={`relative aspect-square rounded-xl overflow-hidden border ${
                        photo.status === "error" ? "border-[#f4212e]" : "border-[#2f3336]"
                      }`}
                    >
                      {/* eslint-disable-next-line @next/next/no-img-element */}
                      <img
                        src={photo.previewUrl}
                        alt={`Selected photo ${index + 1} of ${draftPhotos.length}`}
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
                        onClick={() => handleRemoveDraftPhoto(photo.key)}
                        disabled={isPosting}
                        aria-label={`Remove photo ${index + 1}`}
                        className="absolute top-1 right-1 w-7 h-7 rounded-full bg-black/70 hover:bg-black/90 text-white flex items-center justify-center transition cursor-pointer disabled:opacity-50"
                      >
                        <svg viewBox="0 0 24 24" className="w-4 h-4 fill-none stroke-current" strokeWidth="2.2">
                          <path strokeLinecap="round" d="M6 6l12 12M18 6L6 18" />
                        </svg>
                      </button>
                    </div>
                  ))}
                </div>
              )}

              {postError && (
                <div className="mb-2 text-xs text-[#f4212e] bg-[#f4212e]/10 py-1.5 px-3 rounded-lg border border-[#f4212e]/20">
                  {postError}
                </div>
              )}

              <div className="flex items-center justify-between pt-2 border-t border-[#2f3336]">
                <div className="flex items-center gap-3 text-[#71767b] text-xs">
                  {/* The narrow accept list also makes iOS transcode HEIC to JPEG */}
                  <input
                    ref={fileInputRef}
                    type="file"
                    multiple
                    accept="image/jpeg,image/png,image/webp,image/gif"
                    onChange={handleFilesSelected}
                    className="hidden"
                    id="post-photo-input"
                  />
                  <button
                    type="button"
                    onClick={() => fileInputRef.current?.click()}
                    disabled={isPosting || draftPhotos.length >= MAX_PHOTOS}
                    aria-label="Add photos"
                    title={
                      draftPhotos.length >= MAX_PHOTOS
                        ? `You can add up to ${MAX_PHOTOS} photos`
                        : "Add photos"
                    }
                    className="w-8 h-8 -ml-1 rounded-full flex items-center justify-center text-[#1d9bf0] hover:bg-[#1d9bf0]/10 transition cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed"
                  >
                    <svg viewBox="0 0 24 24" className="w-5 h-5 fill-none stroke-current" strokeWidth="1.8">
                      <rect x="3" y="3" width="18" height="18" rx="2" />
                      <circle cx="8.5" cy="8.5" r="1.6" />
                      <path strokeLinecap="round" strokeLinejoin="round" d="M21 15l-5-5L5 21" />
                    </svg>
                  </button>
                  {draftPhotos.length > 0 && (
                    <span className="text-[#71767b]">
                      {draftPhotos.length}/{MAX_PHOTOS}
                    </span>
                  )}
                  {tweetText.length > 0 && (
                    <span
                      className={`font-mono text-xs ${
                        CHAR_LIMIT - tweetText.length <= 20
                          ? "text-[#f4212e] font-bold"
                          : "text-[#71767b]"
                      }`}
                    >
                      {CHAR_LIMIT - tweetText.length}
                    </span>
                  )}
                </div>
                <button
                  type="submit"
                  disabled={
                    isPosting ||
                    draftPhotos.some((photo) => photo.status === "uploading") ||
                    (!tweetText.trim() &&
                      !draftPhotos.some((photo) => photo.status === "done")) ||
                    [...tweetText].length > CHAR_LIMIT
                  }
                  className="px-4 py-1.5 rounded-full bg-[#1d9bf0] text-white font-bold text-sm hover:bg-[#1a8cd8] disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer transition flex items-center gap-2"
                >
                  {isPosting ? (
                    <>
                      <div className="w-3.5 h-3.5 border-2 border-white border-t-transparent rounded-full animate-spin" />
                      <span>Posting...</span>
                    </>
                  ) : (
                    <span>Post</span>
                  )}
                </button>
              </div>
            </div>
          </form>

          {/* Posts Feed */}
          <div>
            {feedError && (
              <div className="mx-4 mt-3 p-3 bg-[#f4212e]/10 border border-[#f4212e]/30 rounded-xl text-[#f4212e] text-sm">
                {feedError}
              </div>
            )}
            {isLoadingFeed ? (
              <div className="py-12 flex justify-center items-center">
                <div className="w-6 h-6 border-2 border-[#1d9bf0] border-t-transparent rounded-full animate-spin" />
              </div>
            ) : feedPosts.length === 0 ? (
              <div className="py-16 px-6 text-center">
                <p className="text-xl font-bold text-[#eff3f4]">Welcome to your timeline!</p>
                <p className="text-sm text-[#71767b] mt-1.5">
                  Be the first to share an update. Write your thoughts above and click Post.
                </p>
              </div>
            ) : (
              feedPosts.map((post) => (
                <PostCard
                  key={post.id}
                  post={post}
                  isOwner={post.user_id === user.id}
                  onEdit={setEditingPost}
                  onDelete={handleDeletePost}
                />
              ))
            )}
          </div>
        </main>

        {/* Right Widget Area */}
        <aside className="w-[350px] shrink-0 p-4 hidden lg:block">
          <div className="bg-[#16181c] rounded-2xl p-4 border border-[#2f3336]">
            <h2 className="text-lg font-extrabold text-[#eff3f4] mb-3">What&apos;s happening</h2>
            <div className="space-y-4 text-sm">
              <div>
                <p className="text-xs text-[#71767b]">Technology · Trending</p>
                <p className="font-bold text-[#eff3f4]">#NextJS16</p>
                <p className="text-xs text-[#71767b]">24.8K posts</p>
              </div>
              <div>
                <p className="text-xs text-[#71767b]">Software · Trending</p>
                <p className="font-bold text-[#eff3f4]">#GolangCleanArchitecture</p>
                <p className="text-xs text-[#71767b]">15.2K posts</p>
              </div>
            </div>
          </div>
        </aside>

      </div>

      {/* Edit Post Modal */}
      {editingPost && (
        <EditPostModal
          isOpen={Boolean(editingPost)}
          post={editingPost}
          onClose={() => setEditingPost(null)}
          onUpdated={handlePostUpdated}
        />
      )}
    </div>
  );
}
