"use client";

import React, { useState, useMemo, useRef, useCallback } from "react";
import { useRouter } from "next/navigation";

// Months configuration
const MONTHS = [
  { name: "January", value: "01" },
  { name: "February", value: "02" },
  { name: "March", value: "03" },
  { name: "April", value: "04" },
  { name: "May", value: "05" },
  { name: "June", value: "06" },
  { name: "July", value: "07" },
  { name: "August", value: "08" },
  { name: "September", value: "09" },
  { name: "October", value: "10" },
  { name: "November", value: "11" },
  { name: "December", value: "12" },
];

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "";

export default function SignupPage() {
  const router = useRouter();

  // Multi-step state (1: Info, 2: Password, 3: Username)
  const [step, setStep] = useState<number>(1);

  // Form Fields
  const [name, setName] = useState<string>("");
  const [email, setEmail] = useState<string>("");
  const [birthMonth, setBirthMonth] = useState<string>("");
  const [birthDay, setBirthDay] = useState<string>("");
  const [birthYear, setBirthYear] = useState<string>("");

  const [password, setPassword] = useState<string>("");
  const [showPassword, setShowPassword] = useState<boolean>(false);

  const [username, setUsername] = useState<string>("");
  const [isCheckingUsername, setIsCheckingUsername] = useState<boolean>(false);
  const [usernameStatus, setUsernameStatus] = useState<{
    checked: boolean;
    available: boolean;
    message: string;
  }>({ checked: false, available: false, message: "" });

  // UI & Loading States
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [generalError, setGeneralError] = useState<string | null>(null);
  const [focusedField, setFocusedField] = useState<string | null>(null);
  const [signupComplete, setSignupComplete] = useState<boolean>(false);

  // Suggested handles generator (pure deterministic)
  const suggestedHandles = useMemo(() => {
    if (!name.trim()) return [];
    const base = name.toLowerCase().replace(/[^a-z0-9]/g, "").slice(0, 10);
    if (!base) return [];
    return [`${base}_`, `${base}26`, `the_${base}`];
  }, [name]);

  // Days list based on selected month & year
  const daysInMonth = useMemo(() => {
    if (!birthMonth) return 31;
    const m = parseInt(birthMonth, 10);
    if ([4, 6, 9, 11].includes(m)) return 30;
    if (m === 2) {
      const y = parseInt(birthYear, 10);
      if (y && ((y % 4 === 0 && y % 100 !== 0) || y % 400 === 0)) return 29;
      return 28;
    }
    return 31;
  }, [birthMonth, birthYear]);

  // Years list (current year down to 1910)
  const years = useMemo(() => {
    const currentYear = new Date().getFullYear();
    const result = [];
    for (let y = currentYear; y >= currentYear - 110; y--) {
      result.push(y);
    }
    return result;
  }, []);

  // Debounce handle check
  const debounceRef = useRef<NodeJS.Timeout | null>(null);
  const triggerUsernameCheck = useCallback((handleVal: string) => {
    const clean = handleVal.trim().replace(/^@/, "");
    if (!clean) {
      setUsernameStatus({ checked: false, available: false, message: "" });
      return;
    }

    if (clean.length < 3 || clean.length > 15) {
      setUsernameStatus({
        checked: true,
        available: false,
        message: "Your username must be between 3 and 15 characters.",
      });
      return;
    }

    if (!/^[a-zA-Z0-9_]+$/.test(clean)) {
      setUsernameStatus({
        checked: true,
        available: false,
        message: "Your username can only contain letters, numbers and '_' with no spaces.",
      });
      return;
    }

    if (debounceRef.current) clearTimeout(debounceRef.current);
    setIsCheckingUsername(true);

    debounceRef.current = setTimeout(async () => {
      try {
        const res = await fetch(
          `${API_BASE}/api/auth/check-username?username=${encodeURIComponent(clean)}`
        );
        const data = await res.json();
        setIsCheckingUsername(false);
        if (data.available) {
          setUsernameStatus({
            checked: true,
            available: true,
            message: "Username is available!",
          });
        } else {
          setUsernameStatus({
            checked: true,
            available: false,
            message: data.message || "That username has been taken. Please choose another.",
          });
        }
      } catch {
        setIsCheckingUsername(false);
        setUsernameStatus({
          checked: true,
          available: false,
          message: "Unable to verify username right now. Please try again.",
        });
      }
    }, 300);
  }, []);

  const handleUsernameInputChange = (newVal: string) => {
    const clean = newVal.replace(/[^a-zA-Z0-9_]/g, "");
    setUsername(clean);
    triggerUsernameCheck(clean);
  };

  const handleSelectSuggestion = (handle: string) => {
    setUsername(handle);
    triggerUsernameCheck(handle);
  };

  // Validation helpers
  const formattedBirthDate = useMemo(() => {
    if (!birthYear || !birthMonth || !birthDay) return "";
    return `${birthYear}-${birthMonth}-${birthDay.padStart(2, "0")}`;
  }, [birthYear, birthMonth, birthDay]);

  const isStep1Valid = useMemo(() => {
    const isEmailValid = /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.trim());
    const isNameValid = name.trim().length > 0 && name.trim().length <= 50;
    const isDateValid = Boolean(birthMonth && birthDay && birthYear);
    return isEmailValid && isNameValid && isDateValid;
  }, [name, email, birthMonth, birthDay, birthYear]);

  const passwordValidation = useMemo(() => {
    const minLength = password.length >= 8;
    const hasLetter = /[a-zA-Z]/.test(password);
    const hasNumber = /[0-9]/.test(password);
    return {
      minLength,
      hasLetter,
      hasNumber,
      isValid: minLength && hasLetter && hasNumber,
    };
  }, [password]);

  // Navigation handlers
  const handleBack = () => {
    setGeneralError(null);
    if (step > 1) {
      setStep((prev) => prev - 1);
    } else {
      router.push("/");
    }
  };

  // Step 1: Info submitted -> Go to Password
  const handleStep1Next = (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    if (!isStep1Valid) return;
    setGeneralError(null);
    setStep(2);
  };

  // Step 2: Password submitted -> Go to Username
  const handlePasswordSubmit = (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    if (!passwordValidation.isValid) return;
    setGeneralError(null);

    // If username is not set, initialize with default suggestion and verify
    if (!username && suggestedHandles.length > 0) {
      setUsername(suggestedHandles[0]);
      triggerUsernameCheck(suggestedHandles[0]);
    }

    setStep(3);
  };

  // Step 3: Single Signup Call (POST /api/auth/signup)
  const handleSignupSubmit = async (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    const cleanHandle = username.trim().replace(/^@/, "");
    if (!usernameStatus.available || isLoading) return;

    setIsLoading(true);
    setGeneralError(null);

    try {
      const res = await fetch(`${API_BASE}/api/auth/signup`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({
          name: name.trim(),
          username: cleanHandle,
          email: email.trim().toLowerCase(),
          password: password,
          birth_date: formattedBirthDate,
        }),
      });

      const data = await res.json();
      setIsLoading(false);

      if (res.ok && data.status === "success") {
        setSignupComplete(true);
        setTimeout(() => {
          router.push("/login");
        }, 1200);
      } else {
        setGeneralError(data.message || "Registration failed. Please check your details.");
      }
    } catch {
      setIsLoading(false);
      setGeneralError("Network error while creating account. Please check your backend.");
    }
  };

  return (
    <main className="min-h-screen w-full bg-black flex items-center justify-center p-0 sm:p-4 text-[#e7e9ea] font-sans antialiased">
      {/* Modal Container */}
      <div className="w-full sm:max-w-[600px] h-screen sm:h-[650px] bg-black text-[#e7e9ea] sm:rounded-2xl border-0 sm:border border-[#2f3336] flex flex-col relative overflow-hidden shadow-2xl">
        
        {/* Top Header */}
        <header className="h-[53px] px-4 flex items-center justify-between border-b border-transparent shrink-0">
          {/* Back or Close button */}
          <button
            type="button"
            onClick={handleBack}
            className="w-[36px] h-[36px] rounded-full flex items-center justify-center hover:bg-[#181818] transition text-[#eff3f4]"
            aria-label={step > 1 ? "Go back" : "Close"}
          >
            {step > 1 ? (
              <svg className="w-5 h-5 fill-current" viewBox="0 0 24 24">
                <path d="M7.414 13l5.043 5.04-1.414 1.42L3.586 12l7.457-7.46 1.414 1.42L7.414 11H21v2H7.414z" />
              </svg>
            ) : (
              <svg className="w-5 h-5 fill-current" viewBox="0 0 24 24">
                <path d="M10.59 12L4.54 5.96l1.42-1.42L12 10.59l6.04-6.05 1.42 1.42L13.41 12l6.05 6.04-1.42 1.42L12 13.41 5.96 19.46 4.54 18.04 10.59 12z" />
              </svg>
            )}
          </button>

          {/* Centered X Logo */}
          <div className="flex items-center justify-center">
            <svg viewBox="0 0 24 24" aria-hidden="true" className="h-[22px] w-[22px] fill-[#eff3f4]">
              <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z" />
            </svg>
          </div>

          {/* Step Indicator */}
          <div className="text-[14px] font-medium text-[#71767b] min-w-[70px] text-right">
            Step {step} of 3
          </div>
        </header>

        {/* Content Body */}
        <div className="flex-1 px-6 sm:px-16 pt-2 pb-6 flex flex-col justify-between overflow-y-auto">
          {/* Success Banner */}
          {signupComplete ? (
            <div className="flex-1 flex flex-col items-center justify-center text-center py-12 animate-fade-in">
              <div className="w-16 h-16 rounded-full bg-[#00ba7c]/20 border border-[#00ba7c] flex items-center justify-center mb-6">
                <svg className="w-8 h-8 text-[#00ba7c]" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2.5" d="M5 13l4 4L19 7" />
                </svg>
              </div>
              <h2 className="text-2xl font-bold text-[#eff3f4] mb-2">Welcome to X!</h2>
              <p className="text-[#71767b] text-[15px] mb-6">Your account @{username} has been created successfully.</p>
              <div className="text-sm text-[#1d9bf0] font-medium animate-pulse">Routing to login...</div>
            </div>
          ) : (
            <>
              {/* STEP 1: Account Information */}
              {step === 1 && (
                <form onSubmit={handleStep1Next} className="flex-1 flex flex-col justify-between">
                  <div>
                    <h1 className="text-[26px] sm:text-[31px] font-bold text-[#eff3f4] leading-tight mb-8">
                      Create your account
                    </h1>

                    {generalError && (
                      <div className="mb-4 p-3 bg-[#f4212e]/10 border border-[#f4212e]/40 rounded-xl text-[#f4212e] text-sm flex items-center gap-2">
                        <svg className="w-4 h-4 shrink-0 fill-current" viewBox="0 0 24 24">
                          <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z" />
                        </svg>
                        <span>{generalError}</span>
                      </div>
                    )}

                    {/* Name Floating Input */}
                    <div
                      className={`relative mb-6 rounded-lg border transition duration-200 px-3 pt-4 pb-2 ${
                        focusedField === "name"
                          ? "border-[#1d9bf0] ring-1 ring-[#1d9bf0]"
                          : "border-[#333639] hover:border-[#536471]"
                      }`}
                    >
                      <label
                        htmlFor="signup-name"
                        className={`absolute left-3 transition-all duration-150 pointer-events-none ${
                          focusedField === "name" || name
                            ? "top-1 text-[12px] text-[#1d9bf0]"
                            : "top-4 text-[16px] text-[#71767b]"
                        }`}
                      >
                        Name
                      </label>
                      <input
                        id="signup-name"
                        type="text"
                        maxLength={50}
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                        onFocus={() => setFocusedField("name")}
                        onBlur={() => setFocusedField(null)}
                        className="w-full bg-transparent text-[#e7e9ea] text-[16px] outline-none mt-1"
                        autoFocus
                      />
                      <span className="absolute right-3 top-2 text-[12px] text-[#71767b]">
                        {name.length} / 50
                      </span>
                    </div>

                    {/* Email Floating Input */}
                    <div
                      className={`relative mb-8 rounded-lg border transition duration-200 px-3 pt-4 pb-2 ${
                        focusedField === "email"
                          ? "border-[#1d9bf0] ring-1 ring-[#1d9bf0]"
                          : "border-[#333639] hover:border-[#536471]"
                      }`}
                    >
                      <label
                        htmlFor="signup-email"
                        className={`absolute left-3 transition-all duration-150 pointer-events-none ${
                          focusedField === "email" || email
                            ? "top-1 text-[12px] text-[#1d9bf0]"
                            : "top-4 text-[16px] text-[#71767b]"
                        }`}
                      >
                        Email
                      </label>
                      <input
                        id="signup-email"
                        type="email"
                        value={email}
                        onChange={(e) => setEmail(e.target.value)}
                        onFocus={() => setFocusedField("email")}
                        onBlur={() => setFocusedField(null)}
                        className="w-full bg-transparent text-[#e7e9ea] text-[16px] outline-none mt-1"
                      />
                    </div>

                    {/* Date of Birth Section */}
                    <div>
                      <h2 className="text-[15px] font-bold text-[#eff3f4] mb-1">Date of birth</h2>
                      <p className="text-[14px] text-[#71767b] leading-5 mb-4">
                        This will not be shown publicly. Confirm your own age, even if this account is for a business, a pet, or something else.
                      </p>

                      <div className="flex gap-3">
                        {/* Month */}
                        <div className="flex-[2] relative border border-[#333639] rounded-lg p-2 hover:border-[#536471] focus-within:border-[#1d9bf0] transition">
                          <label className="text-[11px] text-[#71767b] block">Month</label>
                          <select
                            id="signup-month"
                            value={birthMonth}
                            onChange={(e) => setBirthMonth(e.target.value)}
                            className="w-full bg-transparent text-[#e7e9ea] text-[15px] outline-none cursor-pointer appearance-none pr-4"
                          >
                            <option value="" disabled className="bg-black text-[#71767b]"></option>
                            {MONTHS.map((m) => (
                              <option key={m.value} value={m.value} className="bg-black text-[#e7e9ea]">
                                {m.name}
                              </option>
                            ))}
                          </select>
                          <span className="absolute right-3 bottom-3 pointer-events-none text-[#71767b] text-xs">▼</span>
                        </div>

                        {/* Day */}
                        <div className="flex-1 relative border border-[#333639] rounded-lg p-2 hover:border-[#536471] focus-within:border-[#1d9bf0] transition">
                          <label className="text-[11px] text-[#71767b] block">Day</label>
                          <select
                            id="signup-day"
                            value={birthDay}
                            onChange={(e) => setBirthDay(e.target.value)}
                            className="w-full bg-transparent text-[#e7e9ea] text-[15px] outline-none cursor-pointer appearance-none pr-4"
                          >
                            <option value="" disabled className="bg-black text-[#71767b]"></option>
                            {Array.from({ length: daysInMonth }, (_, i) => i + 1).map((d) => (
                              <option key={d} value={String(d)} className="bg-black text-[#e7e9ea]">
                                {d}
                              </option>
                            ))}
                          </select>
                          <span className="absolute right-3 bottom-3 pointer-events-none text-[#71767b] text-xs">▼</span>
                        </div>

                        {/* Year */}
                        <div className="flex-1 relative border border-[#333639] rounded-lg p-2 hover:border-[#536471] focus-within:border-[#1d9bf0] transition">
                          <label className="text-[11px] text-[#71767b] block">Year</label>
                          <select
                            id="signup-year"
                            value={birthYear}
                            onChange={(e) => setBirthYear(e.target.value)}
                            className="w-full bg-transparent text-[#e7e9ea] text-[15px] outline-none cursor-pointer appearance-none pr-4"
                          >
                            <option value="" disabled className="bg-black text-[#71767b]"></option>
                            {years.map((y) => (
                              <option key={y} value={String(y)} className="bg-black text-[#e7e9ea]">
                                {y}
                              </option>
                            ))}
                          </select>
                          <span className="absolute right-3 bottom-3 pointer-events-none text-[#71767b] text-xs">▼</span>
                        </div>
                      </div>
                    </div>
                  </div>

                  {/* Submit Button */}
                  <div className="pt-6">
                    <button
                      type="submit"
                      disabled={!isStep1Valid || isLoading}
                      className="w-full h-[52px] rounded-full bg-[#eff3f4] text-[#0f1419] font-bold text-[16px] hover:bg-[#d7dbdc] active:bg-[#b9bcbd] disabled:opacity-50 disabled:cursor-not-allowed transition duration-200 flex items-center justify-center cursor-pointer shadow-md"
                    >
                      {isLoading ? (
                        <div className="w-5 h-5 border-2 border-[#0f1419] border-t-transparent rounded-full animate-spin" />
                      ) : (
                        "Next"
                      )}
                    </button>
                  </div>
                </form>
              )}

              {/* STEP 2: Password */}
              {step === 2 && (
                <form onSubmit={handlePasswordSubmit} className="flex-1 flex flex-col justify-between">
                  <div>
                    <h1 className="text-[26px] sm:text-[31px] font-bold text-[#eff3f4] leading-tight mb-2">
                      You&apos;ll need a password
                    </h1>
                    <p className="text-[15px] text-[#71767b] mb-8">
                      Make sure it’s 8 characters or more and includes letters & numbers.
                    </p>

                    {generalError && (
                      <div className="mb-4 p-3 bg-[#f4212e]/10 border border-[#f4212e]/40 rounded-xl text-[#f4212e] text-sm flex items-center gap-2">
                        <svg className="w-4 h-4 shrink-0 fill-current" viewBox="0 0 24 24">
                          <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z" />
                        </svg>
                        <span>{generalError}</span>
                      </div>
                    )}

                    {/* Password Floating Input with Toggle */}
                    <div
                      className={`relative mb-6 rounded-lg border transition duration-200 px-3 pt-4 pb-2 ${
                        focusedField === "password"
                          ? "border-[#1d9bf0] ring-1 ring-[#1d9bf0]"
                          : "border-[#333639] hover:border-[#536471]"
                      }`}
                    >
                      <label
                        htmlFor="signup-password"
                        className={`absolute left-3 transition-all duration-150 pointer-events-none ${
                          focusedField === "password" || password
                            ? "top-1 text-[12px] text-[#1d9bf0]"
                            : "top-4 text-[16px] text-[#71767b]"
                        }`}
                      >
                        Password
                      </label>
                      <input
                        id="signup-password"
                        type={showPassword ? "text" : "password"}
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        onFocus={() => setFocusedField("password")}
                        onBlur={() => setFocusedField(null)}
                        className="w-full bg-transparent text-[#e7e9ea] text-[16px] outline-none mt-1 pr-10"
                        autoFocus
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

                    {/* Requirements Checklist */}
                    <div className="space-y-2 text-[14px]">
                      <div className="flex items-center gap-2">
                        <span className={`text-base ${passwordValidation.minLength ? "text-[#00ba7c]" : "text-[#71767b]"}`}>
                          {passwordValidation.minLength ? "✓" : "○"}
                        </span>
                        <span className={passwordValidation.minLength ? "text-[#e7e9ea]" : "text-[#71767b]"}>
                          At least 8 characters
                        </span>
                      </div>
                      <div className="flex items-center gap-2">
                        <span className={`text-base ${passwordValidation.hasLetter && passwordValidation.hasNumber ? "text-[#00ba7c]" : "text-[#71767b]"}`}>
                          {passwordValidation.hasLetter && passwordValidation.hasNumber ? "✓" : "○"}
                        </span>
                        <span className={passwordValidation.hasLetter && passwordValidation.hasNumber ? "text-[#e7e9ea]" : "text-[#71767b]"}>
                          Contains both letters and numbers
                        </span>
                      </div>
                    </div>
                  </div>

                  <div className="pt-6">
                    <button
                      type="submit"
                      disabled={!passwordValidation.isValid}
                      className="w-full h-[52px] rounded-full bg-[#eff3f4] text-[#0f1419] font-bold text-[16px] hover:bg-[#d7dbdc] active:bg-[#b9bcbd] disabled:opacity-50 disabled:cursor-not-allowed transition duration-200 flex items-center justify-center cursor-pointer shadow-md"
                    >
                      Next
                    </button>
                  </div>
                </form>
              )}

              {/* STEP 3: What Should We Call You? */}
              {step === 3 && (
                <form onSubmit={handleSignupSubmit} className="flex-1 flex flex-col justify-between">
                  <div>
                    <h1 className="text-[26px] sm:text-[31px] font-bold text-[#eff3f4] leading-tight mb-2">
                      What should we call you?
                    </h1>
                    <p className="text-[15px] text-[#71767b] mb-6">
                      Your @username is unique. You can always change it later.
                    </p>

                    {generalError && (
                      <div className="mb-4 p-3 bg-[#f4212e]/10 border border-[#f4212e]/40 rounded-xl text-[#f4212e] text-sm flex items-center gap-2">
                        <svg className="w-4 h-4 shrink-0 fill-current" viewBox="0 0 24 24">
                          <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z" />
                        </svg>
                        <span>{generalError}</span>
                      </div>
                    )}

                    {/* Username Input with @ Prefix */}
                    <div
                      className={`relative mb-2 rounded-lg border transition duration-200 px-3 pt-4 pb-2 ${
                        focusedField === "username"
                          ? "border-[#1d9bf0] ring-1 ring-[#1d9bf0]"
                          : usernameStatus.checked && !usernameStatus.available
                          ? "border-[#f4212e]"
                          : usernameStatus.checked && usernameStatus.available
                          ? "border-[#00ba7c]"
                          : "border-[#333639] hover:border-[#536471]"
                      }`}
                    >
                      <label
                        htmlFor="signup-username"
                        className={`absolute left-3 transition-all duration-150 pointer-events-none ${
                          focusedField === "username" || username
                            ? "top-1 text-[12px] text-[#1d9bf0]"
                            : "top-4 text-[16px] text-[#71767b]"
                        }`}
                      >
                        Username
                      </label>
                      <div className="flex items-center mt-1">
                        <span className="text-[#71767b] text-[16px] select-none mr-0.5">@</span>
                        <input
                          id="signup-username"
                          type="text"
                          maxLength={15}
                          value={username.replace(/^@/, "")}
                          onChange={(e) => handleUsernameInputChange(e.target.value)}
                          onFocus={() => setFocusedField("username")}
                          onBlur={() => setFocusedField(null)}
                          className="w-full bg-transparent text-[#e7e9ea] text-[16px] outline-none"
                          autoFocus
                        />
                        {isCheckingUsername ? (
                          <div className="w-4 h-4 border-2 border-[#1d9bf0] border-t-transparent rounded-full animate-spin shrink-0" />
                        ) : usernameStatus.checked ? (
                          usernameStatus.available ? (
                            <span className="text-[#00ba7c] font-bold text-lg shrink-0">✓</span>
                          ) : (
                            <span className="text-[#f4212e] font-bold text-lg shrink-0">✕</span>
                          )
                        ) : null}
                      </div>
                    </div>

                    {/* Username feedback message */}
                    {usernameStatus.message && (
                      <p
                        className={`text-[13px] mb-4 ${
                          usernameStatus.available ? "text-[#00ba7c]" : "text-[#f4212e]"
                        }`}
                      >
                        {usernameStatus.message}
                      </p>
                    )}

                    {/* Handle suggestions */}
                    {suggestedHandles.length > 0 && (
                      <div className="mt-4">
                        <span className="text-[13px] text-[#71767b] block mb-2">Suggestions:</span>
                        <div className="flex flex-wrap gap-2">
                          {suggestedHandles.map((handle) => (
                            <button
                              key={handle}
                              type="button"
                              onClick={() => handleSelectSuggestion(handle)}
                              className={`text-[13px] px-3 py-1.5 rounded-full border transition cursor-pointer ${
                                username === handle
                                  ? "border-[#1d9bf0] text-[#1d9bf0] bg-[#1d9bf0]/10"
                                  : "border-[#333639] text-[#71767b] hover:border-[#1d9bf0] hover:text-[#1d9bf0]"
                              }`}
                            >
                              @{handle}
                            </button>
                          ))}
                        </div>
                      </div>
                    )}
                  </div>

                  <div className="pt-6">
                    <button
                      type="submit"
                      disabled={!usernameStatus.available || isLoading}
                      className="w-full h-[52px] rounded-full bg-[#eff3f4] text-[#0f1419] font-bold text-[16px] hover:bg-[#d7dbdc] active:bg-[#b9bcbd] disabled:opacity-50 disabled:cursor-not-allowed transition duration-200 flex items-center justify-center cursor-pointer shadow-md"
                    >
                      {isLoading ? (
                        <div className="w-5 h-5 border-2 border-[#0f1419] border-t-transparent rounded-full animate-spin" />
                      ) : (
                        "Sign up"
                      )}
                    </button>
                  </div>
                </form>
              )}
            </>
          )}
        </div>
      </div>
    </main>
  );
}
