"use client";

import React, { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "";

export default function LoginPage() {
  const router = useRouter();

  // Form Fields
  const [email, setEmail] = useState<string>("");
  const [password, setPassword] = useState<string>("");
  const [showPassword, setShowPassword] = useState<boolean>(false);

  // UI & Loading States
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [focusedField, setFocusedField] = useState<string | null>(null);
  const [isSuccess, setIsSuccess] = useState<boolean>(false);

  const isValidEmail = (val: string) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(val.trim());

  const canSubmit = Boolean(email.trim() && password && isValidEmail(email));

  const handleLoginSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!canSubmit || isLoading) return;

    setIsLoading(true);
    setError(null);

    try {
      const res = await fetch(`${API_BASE}/api/auth/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({
          email: email.trim().toLowerCase(),
          password: password,
        }),
      });

      const data = await res.json();
      setIsLoading(false);

      if (res.ok && data.status === "success") {
        setIsSuccess(true);
        setTimeout(() => {
          router.push("/dashboard");
        }, 800);
      } else {
        setError(data.message || "Invalid email or password. Please try again.");
      }
    } catch {
      setIsLoading(false);
      setError("Network error while connecting to server. Please try again.");
    }
  };

  return (
    <main className="min-h-screen w-full bg-black flex items-center justify-center p-0 sm:p-4 text-[#e7e9ea] font-sans antialiased">
      {/* Modal Container */}
      <div className="w-full sm:max-w-[600px] h-screen sm:h-[650px] bg-black text-[#e7e9ea] sm:rounded-2xl border-0 sm:border border-[#2f3336] flex flex-col relative overflow-hidden shadow-2xl">
        {/* Top Header */}
        <header className="h-[53px] px-4 flex items-center justify-between border-b border-transparent shrink-0">
          <Link
            href="/"
            className="w-[36px] h-[36px] rounded-full flex items-center justify-center hover:bg-[#181818] transition text-[#eff3f4]"
            aria-label="Close"
          >
            <svg className="w-5 h-5 fill-current" viewBox="0 0 24 24">
              <path d="M10.59 12L4.54 5.96l1.42-1.42L12 10.59l6.04-6.05 1.42 1.42L13.41 12l6.05 6.04-1.42 1.42L12 13.41 5.96 19.46 4.54 18.04 10.59 12z" />
            </svg>
          </Link>

          {/* Centered X Logo */}
          <div className="flex items-center justify-center">
            <svg viewBox="0 0 24 24" aria-hidden="true" className="h-[22px] w-[22px] fill-[#eff3f4]">
              <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z" />
            </svg>
          </div>

          {/* Balanced empty spacing */}
          <div className="w-[36px]" />
        </header>

        {/* Content Body */}
        <div className="flex-1 px-6 sm:px-20 pt-4 pb-8 flex flex-col justify-between overflow-y-auto">
          {isSuccess ? (
            <div className="flex-1 flex flex-col items-center justify-center text-center py-12 animate-fade-in">
              <div className="w-16 h-16 rounded-full bg-[#00ba7c]/20 border border-[#00ba7c] flex items-center justify-center mb-6">
                <svg className="w-8 h-8 text-[#00ba7c]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2.5" d="M5 13l4 4L19 7" />
                </svg>
              </div>
              <h2 className="text-2xl font-bold text-[#eff3f4] mb-2">Welcome back!</h2>
              <p className="text-[#71767b] text-[15px] mb-6">Signing you into your account...</p>
              <div className="text-sm text-[#1d9bf0] font-medium animate-pulse">Redirecting to your timeline...</div>
            </div>
          ) : (
            <form onSubmit={handleLoginSubmit} className="flex-1 flex flex-col justify-between">
              <div>
                <h1 className="text-[26px] sm:text-[31px] font-bold text-[#eff3f4] leading-tight mb-8">
                  Sign in to X
                </h1>

                {/* Error Banner */}
                {error && (
                  <div className="mb-6 p-3 bg-[#f4212e]/10 border border-[#f4212e]/40 rounded-xl text-[#f4212e] text-sm flex items-center gap-2">
                    <svg className="w-4 h-4 shrink-0 fill-current" viewBox="0 0 24 24">
                      <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z" />
                    </svg>
                    <span>{error}</span>
                  </div>
                )}

                {/* Email Input */}
                <div
                  className={`relative mb-6 rounded-lg border transition duration-200 px-3 pt-4 pb-2 ${
                    focusedField === "email"
                      ? "border-[#1d9bf0] ring-1 ring-[#1d9bf0]"
                      : "border-[#333639] hover:border-[#536471]"
                  }`}
                >
                  <label
                    htmlFor="login-email"
                    className={`absolute left-3 transition-all duration-150 pointer-events-none ${
                      focusedField === "email" || email
                        ? "top-1 text-[12px] text-[#1d9bf0]"
                        : "top-4 text-[16px] text-[#71767b]"
                    }`}
                  >
                    Email
                  </label>
                  <input
                    id="login-email"
                    type="email"
                    value={email}
                    onChange={(e) => {
                      setEmail(e.target.value);
                      if (error) setError(null);
                    }}
                    onFocus={() => setFocusedField("email")}
                    onBlur={() => setFocusedField(null)}
                    className="w-full bg-transparent text-[#e7e9ea] text-[16px] outline-none mt-1"
                    autoFocus
                    autoComplete="email"
                    required
                  />
                </div>

                {/* Password Input */}
                <div
                  className={`relative mb-8 rounded-lg border transition duration-200 px-3 pt-4 pb-2 ${
                    focusedField === "password"
                      ? "border-[#1d9bf0] ring-1 ring-[#1d9bf0]"
                      : "border-[#333639] hover:border-[#536471]"
                  }`}
                >
                  <label
                    htmlFor="login-password"
                    className={`absolute left-3 transition-all duration-150 pointer-events-none ${
                      focusedField === "password" || password
                        ? "top-1 text-[12px] text-[#1d9bf0]"
                        : "top-4 text-[16px] text-[#71767b]"
                    }`}
                  >
                    Password
                  </label>
                  <input
                    id="login-password"
                    type={showPassword ? "text" : "password"}
                    value={password}
                    onChange={(e) => {
                      setPassword(e.target.value);
                      if (error) setError(null);
                    }}
                    onFocus={() => setFocusedField("password")}
                    onBlur={() => setFocusedField(null)}
                    className="w-full bg-transparent text-[#e7e9ea] text-[16px] outline-none mt-1 pr-10"
                    autoComplete="current-password"
                    required
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-3 top-4 text-[#71767b] hover:text-[#eff3f4] transition"
                    aria-label="Toggle password visibility"
                  >
                    {showPassword ? (
                      <svg className="w-5 h-5 fill-current" viewBox="0 0 24 24">
                        <path d="M12 4.5C7 4.5 2.73 7.61 1 12c1.73 4.39 6 7.5 11 7.5s9.27-3.11 11-7.5c-1.73-4.39-6-7.5-11-7.5zM12 17c-2.76 0-5-2.24-5-5s2.24-5 5-5 5 2.24 5 5-2.24 5-5 5zm0-8c-1.66 0-3 1.34-3 3s1.34 3 3 3 3-1.34 3-3-1.34-3-3-3z" />
                      </svg>
                    ) : (
                      <svg className="w-5 h-5 fill-current" viewBox="0 0 24 24">
                        <path d="M12 7c2.76 0 5 2.24 5 5 0 .65-.13 1.26-.36 1.83l2.92 2.92c1.51-1.26 2.7-2.89 3.43-4.75-1.73-4.39-6-7.5-11-7.5-1.4 0-2.74.25-3.98.7l2.16 2.16C10.74 7.13 11.35 7 12 7zM2 4.27l2.28 2.28.46.46C3.08 8.3 1.78 10.02 1 12c1.73 4.39 6 7.5 11 7.5 1.55 0 3.03-.3 4.38-.84l.42.42L19.73 22 21 20.73 3.27 3 2 4.27zM7.53 9.8l1.55 1.55c-.05.21-.08.43-.08.65 0 1.66 1.34 3 3 3 .22 0 .44-.03.65-.08l1.55 1.55c-.67.33-1.41.53-2.2.53-2.76 0-5-2.24-5-5 0-.79.2-1.53.53-2.2zm4.31-.78l3.15 3.15.02-.16c0-1.66-1.34-3-3-3l-.17.01z" />
                      </svg>
                    )}
                  </button>
                </div>
              </div>

              {/* Action Buttons & Links */}
              <div className="flex flex-col gap-4 pt-4">
                <button
                  type="submit"
                  disabled={!canSubmit || isLoading}
                  className="w-full h-[52px] rounded-full bg-[#eff3f4] text-[#0f1419] font-bold text-[16px] hover:bg-[#d7dbdc] active:bg-[#b9bcbd] disabled:opacity-50 disabled:cursor-not-allowed transition duration-200 flex items-center justify-center cursor-pointer shadow-md"
                >
                  {isLoading ? (
                    <div className="w-5 h-5 border-2 border-[#0f1419] border-t-transparent rounded-full animate-spin" />
                  ) : (
                    "Log in"
                  )}
                </button>

                <div className="text-center pt-2">
                  <p className="text-[15px] text-[#71767b]">
                    Don&apos;t have an account?{" "}
                    <Link href="/signup" className="text-[#1d9bf0] hover:underline font-medium">
                      Sign up
                    </Link>
                  </p>
                </div>
              </div>
            </form>
          )}
        </div>
      </div>
    </main>
  );
}
