"use client";

import React, { useState, useEffect } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import UserAvatar from "../../components/ui/UserAvatar";
import AccountPill from "../../components/navigation/AccountPill";

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

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "";

export default function DashboardPage() {
  const router = useRouter();

  const [user, setUser] = useState<UserProfile | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);

  const [tweetText, setTweetText] = useState<string>("");
  const [tweets, setTweets] = useState<Array<{ id: string; author: string; handle: string; time: string; text: string; likes: number }>>([
    {
      id: "post_1",
      author: "X Clone Team",
      handle: "xclone",
      time: "10m",
      text: "Welcome to the full-stack X (Twitter) clone! Sign-up flow, layered Go backend, and modern Next.js 16 frontend are now live 🚀",
      likes: 42,
    },
  ]);

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

  const handlePostTweet = (e: React.FormEvent) => {
    e.preventDefault();
    if (!tweetText.trim()) return;

    const newPost = {
      id: `post_${Date.now()}`,
      author: user?.name || "Anonymous",
      handle: user?.username || "user",
      time: "Just now",
      text: tweetText.trim(),
      likes: 0,
    };

    setTweets([newPost, ...tweets]);
    setTweetText("");
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
            <div className="flex-1">
              <textarea
                value={tweetText}
                onChange={(e) => setTweetText(e.target.value)}
                placeholder="What is happening?!"
                className="w-full bg-transparent text-[18px] text-[#e7e9ea] outline-none placeholder-[#71767b] resize-none min-h-[60px]"
              />
              <div className="flex items-center justify-between pt-2 border-t border-[#2f3336]">
                <div className="flex gap-3 text-[#1d9bf0]">
                  <span className="cursor-pointer hover:opacity-80">🖼️</span>
                  <span className="cursor-pointer hover:opacity-80">📊</span>
                  <span className="cursor-pointer hover:opacity-80">😊</span>
                </div>
                <button
                  type="submit"
                  disabled={!tweetText.trim()}
                  className="px-4 py-1.5 rounded-full bg-[#1d9bf0] text-white font-bold text-sm hover:bg-[#1a8cd8] disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer transition"
                >
                  Post
                </button>
              </div>
            </div>
          </form>

          {/* Posts Feed */}
          <div>
            {tweets.map((post) => (
              <article key={post.id} className="p-4 border-b border-[#2f3336] hover:bg-[#080808] transition flex gap-3">
                <Link href={`/${post.handle}`} className="shrink-0">
                  <UserAvatar
                    name={post.author}
                    size="md"
                    className="hover:opacity-90 transition cursor-pointer"
                  />
                </Link>
                <div className="flex-1">
                  <div className="flex items-center gap-1.5 text-sm">
                    <Link href={`/${post.handle}`} className="font-bold text-[#eff3f4] hover:underline cursor-pointer">
                      {post.author}
                    </Link>
                    <Link href={`/${post.handle}`} className="text-[#71767b] hover:underline cursor-pointer">
                      @{post.handle}
                    </Link>
                    <span className="text-[#71767b]">·</span>
                    <span className="text-[#71767b]">{post.time}</span>
                  </div>
                  <p className="mt-1 text-[15px] leading-relaxed text-[#e7e9ea]">{post.text}</p>
                  <div className="mt-3 flex items-center gap-12 text-[#71767b] text-xs">
                    <span className="hover:text-[#1d9bf0] cursor-pointer">💬 2</span>
                    <span className="hover:text-[#00ba7c] cursor-pointer">🔁 5</span>
                    <span className="hover:text-[#f4212e] cursor-pointer">❤️ {post.likes}</span>
                    <span className="hover:text-[#1d9bf0] cursor-pointer">📊 1.2K</span>
                  </div>
                </div>
              </article>
            ))}
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
    </div>
  );
}
