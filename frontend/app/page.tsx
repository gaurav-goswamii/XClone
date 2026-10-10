"use client";

import Link from "next/link";

export default function HomePage() {
  return (
    <div className="min-h-screen bg-black text-[#e7e9ea] flex flex-col justify-between selection:bg-[#1d9bf0]/30 selection:text-white">
      {/* Main Content Area */}
      <main className="flex-1 flex flex-col lg:flex-row items-center justify-center lg:justify-between px-6 sm:px-12 md:px-24 py-8 max-w-7xl mx-auto w-full">
        {/* Left Side: Large X Logo */}
        <div className="w-full lg:w-1/2 flex items-center justify-center py-8 lg:py-16">
          <svg
            viewBox="0 0 24 24"
            aria-hidden="true"
            className="w-20 h-20 sm:w-48 sm:h-48 lg:w-80 lg:h-80 fill-[#eff3f4]"
          >
            <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z" />
          </svg>
        </div>

        {/* Right Side: Headlines & Auth Actions */}
        <div className="w-full lg:w-1/2 flex flex-col items-start lg:pl-12">
          <h1 className="text-[42px] sm:text-[58px] lg:text-[68px] font-extrabold tracking-tight text-[#eff3f4] leading-tight mb-8">
            Happening now
          </h1>

          <h2 className="text-[24px] sm:text-[32px] font-bold text-[#eff3f4] mb-8">
            Join today.
          </h2>

          <div className="w-full max-w-[320px] flex flex-col gap-3">
            {/* Create Account Button */}
            <Link
              href="/signup"
              className="w-full h-[44px] rounded-full bg-[#1d9bf0] hover:bg-[#1a8cd8] active:bg-[#177cc0] text-white font-bold text-[15px] flex items-center justify-center transition duration-200 shadow-md"
            >
              Create account
            </Link>

            <p className="text-[11px] text-[#71767b] leading-3.5 mb-6">
              By signing up, you agree to the{" "}
              <span className="text-[#1d9bf0] hover:underline cursor-pointer">Terms of Service</span> and{" "}
              <span className="text-[#1d9bf0] hover:underline cursor-pointer">Privacy Policy</span>, including{" "}
              <span className="text-[#1d9bf0] hover:underline cursor-pointer">Cookie Use</span>.
            </p>

            <div className="mt-4">
              <h3 className="text-[17px] font-bold text-[#eff3f4] mb-3">
                Already have an account?
              </h3>
              <Link
                href="/login"
                className="w-full h-[44px] rounded-full border border-[#536471] hover:bg-[#eff3f4]/10 active:bg-[#eff3f4]/20 text-[#1d9bf0] font-bold text-[15px] flex items-center justify-center transition duration-200"
              >
                Sign in
              </Link>
            </div>
          </div>
        </div>
      </main>

      {/* Footer Nav Links */}
      <footer className="py-4 px-6 text-center text-[12px] text-[#71767b] flex flex-wrap justify-center gap-x-4 gap-y-2 border-t border-[#2f3336]">
        <span className="hover:underline cursor-pointer">About</span>
        <span className="hover:underline cursor-pointer">Download the X app</span>
        <span className="hover:underline cursor-pointer">Help Center</span>
        <span className="hover:underline cursor-pointer">Terms of Service</span>
        <span className="hover:underline cursor-pointer">Privacy Policy</span>
        <span className="hover:underline cursor-pointer">Cookie Policy</span>
        <span className="hover:underline cursor-pointer">Accessibility</span>
        <span className="hover:underline cursor-pointer">Ads info</span>
        <span className="hover:underline cursor-pointer">Blog</span>
        <span className="hover:underline cursor-pointer">Careers</span>
        <span className="hover:underline cursor-pointer">Brand Resources</span>
        <span className="hover:underline cursor-pointer">Advertising</span>
        <span className="hover:underline cursor-pointer">Marketing</span>
        <span className="hover:underline cursor-pointer">X for Business</span>
        <span className="hover:underline cursor-pointer">Developers</span>
        <span>© 2026 X Corp.</span>
      </footer>
    </div>
  );
}
