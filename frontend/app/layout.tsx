import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "X. It's what's happening / X",
  description: "From breaking news and entertainment to sports and politics, get the full story with all the live commentary.",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en" className={`${geistSans.variable} ${geistMono.variable} dark h-full bg-black`}>
      <body className="min-h-full bg-black text-[#e7e9ea] antialiased selection:bg-[#1d9bf0]/30 selection:text-white">
        {children}
      </body>
    </html>
  );
}
