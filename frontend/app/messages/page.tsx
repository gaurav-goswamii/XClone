"use client";

import React, { useCallback, useEffect, useRef, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import UserAvatar from "../../components/ui/UserAvatar";

// The `|| ""` matters: without it an unset NEXT_PUBLIC_API_URL makes every
// request URL start with the literal string "undefined". Every other page in
// the app already has this fallback.
const API_BASE = process.env.NEXT_PUBLIC_API_URL || "";

// How often an open thread re-asks the server. A thread you are staring at
// should feel near-live; this is the one place to change that if 3s is wrong.
const POLL_INTERVAL_MS = 3000;
// Mirrors MaxMessageRunes in backend/service/message.go, so the usual mistake
// costs no round trip
const MESSAGE_LIMIT = 2000;

interface CurrentUser {
  id: string;
  name: string;
  username: string;
  avatar_url?: string;
}

interface Chat {
  id: string;
  user1_id: string;
  user2_id: string;
  created_at: string;
}

interface Message {
  id: string;
  chat_id: string;
  sender_id: string;
  content: string;
  seq: number;
  created_at: string;
}

// errorFrom pulls the server's message out of the house envelope. The API
// always answers {status, message}, so a failure has something to show.
async function errorFrom(res: Response, fallback: string): Promise<string> {
  const data = await res.json().catch(() => null);
  return data?.message || fallback;
}

function formatTime(iso: string): string {
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return "";
  return date.toLocaleTimeString([], { hour: "numeric", minute: "2-digit" });
}

export default function MessagesPage() {
  const router = useRouter();

  const [user, setUser] = useState<CurrentUser | null>(null);
  const [isLoading, setIsLoading] = useState<boolean>(true);

  // Which conversation is open, and with whom. The handle is kept alongside
  // the chat because the chat row only carries user IDs.
  const [chat, setChat] = useState<Chat | null>(null);
  const [partnerUsername, setPartnerUsername] = useState<string>("");

  const [usernameInput, setUsernameInput] = useState<string>("");
  const [isOpeningChat, setIsOpeningChat] = useState<boolean>(false);
  const [openChatError, setOpenChatError] = useState<string | null>(null);

  const [messages, setMessages] = useState<Message[]>([]);
  const [isLoadingMessages, setIsLoadingMessages] = useState<boolean>(false);
  const [messagesError, setMessagesError] = useState<string | null>(null);

  const [draft, setDraft] = useState<string>("");
  const [isSending, setIsSending] = useState<boolean>(false);
  const [sendError, setSendError] = useState<string | null>(null);

  const bottomRef = useRef<HTMLDivElement>(null);

  // Session check, same shape as every other page in the app
  useEffect(() => {
    let isCancelled = false;

    fetch(`${API_BASE}/api/auth/me`, { credentials: "include" })
      .then((res) => (res.ok ? res.json() : null))
      .then((data) => {
        if (isCancelled) return;
        if (data?.status === "success" && data.data?.user) {
          setUser(data.data.user);
          setIsLoading(false);
        } else {
          router.replace("/login");
        }
      })
      .catch(() => {
        if (!isCancelled) router.replace("/login");
      });

    return () => {
      isCancelled = true;
    };
  }, [router]);

  // Open (or reuse) a chat with the handle in the box. The endpoint is
  // idempotent, so submitting the same name twice is harmless.
  const handleOpenChat = async (event: React.FormEvent) => {
    event.preventDefault();

    const username = usernameInput.trim().replace(/^@/, "").toLowerCase();
    if (!username || isOpeningChat) return;

    setIsOpeningChat(true);
    setOpenChatError(null);

    try {
      const res = await fetch(`${API_BASE}/api/chats`, {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username }),
      });

      if (!res.ok) {
        setOpenChatError(await errorFrom(res, "Could not open that chat."));
        return;
      }

      const data = await res.json();
      // Switching threads clears the old one's messages and errors, or the
      // previous conversation's bubbles would show under the new name
      setChat(data.data.chat);
      setPartnerUsername(username);
      setMessages([]);
      setMessagesError(null);
      setSendError(null);
      setDraft("");
      setIsLoadingMessages(true);
      setUsernameInput("");
    } catch {
      setOpenChatError("Network error. Please try again.");
    } finally {
      setIsOpeningChat(false);
    }
  };

  // loadMessages is deliberately a plain fetch returning the thread, with no
  // state writing of its own: the poll decides what to do with a late answer.
  const loadMessages = useCallback(async (chatID: string): Promise<Message[] | null> => {
    try {
      const res = await fetch(`${API_BASE}/api/chats/${chatID}/messages`, {
        credentials: "include",
      });
      if (!res.ok) return null;
      const data = await res.json();
      return Array.isArray(data?.data?.messages) ? data.data.messages : [];
    } catch {
      return null;
    }
  }, []);

  // The polling loop. Four guards, each preventing a different real failure —
  // this is the part of the feature most likely to go wrong.
  useEffect(() => {
    const chatID = chat?.id;
    if (!chatID) return;

    let cancelled = false;
    let inFlight = false;

    const tick = async () => {
      // 1. unmounted, or the open thread changed
      if (cancelled) return;
      // 2. a response slower than the interval must not let requests stack up
      if (inFlight) return;
      // 3. a tab nobody is looking at should not poll all day
      if (document.visibilityState !== "visible") return;

      inFlight = true;
      const fetched = await loadMessages(chatID);
      inFlight = false;

      // 4. the answer that lands *after* switching away. Guard 1 cannot catch
      // this one: the await had already started when the thread changed.
      if (cancelled) return;

      setIsLoadingMessages(false);

      if (fetched === null) {
        setMessagesError("Could not load messages.");
        return;
      }

      setMessagesError(null);
      // Replacing only when something actually changed keeps React from
      // re-rendering the whole thread every 3 seconds. The server returns the
      // messages ordered by seq, so the last one's seq identifies the state.
      setMessages((prev) => {
        const prevLast = prev.length > 0 ? prev[prev.length - 1].seq : 0;
        const nextLast = fetched.length > 0 ? fetched[fetched.length - 1].seq : 0;
        if (prev.length === fetched.length && prevLast === nextLast) {
          return prev;
        }
        return fetched;
      });
    };

    const intervalID = setInterval(tick, POLL_INTERVAL_MS);
    // Fire once immediately, so opening a chat is not a 3s wait
    tick();

    return () => {
      // Without this the thread you just left keeps polling forever, and ten
      // opened chats become ten requests a second writing into the same state
      cancelled = true;
      clearInterval(intervalID);
    };
  }, [chat?.id, loadMessages]);

  // Keep the newest message in view as the thread grows
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ block: "end" });
  }, [messages.length]);

  // No optimistic bubble on purpose: wait for the server, then append. An
  // optimistic one needs a temporary ID, reconciliation against the real
  // seq, and care that the poll does not deliver the same message twice.
  const handleSend = async (event: React.FormEvent) => {
    event.preventDefault();

    const content = draft.trim();
    if (!chat || !content || isSending) return;

    setIsSending(true);
    setSendError(null);

    try {
      const res = await fetch(`${API_BASE}/api/chats/${chat.id}/messages`, {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ content }),
      });

      if (!res.ok) {
        setSendError(await errorFrom(res, "Could not send that message."));
        return;
      }

      const data = await res.json();
      const sent: Message = data.data.message;

      // The poll would pick this up within 3s anyway; appending here is what
      // makes your own message appear the instant it is stored. Guarding on
      // seq keeps it from being added twice if a poll landed in between.
      setMessages((prev) =>
        prev.some((message) => message.seq === sent.seq) ? prev : [...prev, sent]
      );
      setDraft("");
    } catch {
      setSendError("Network error. Please try again.");
    } finally {
      setIsSending(false);
    }
  };

  if (isLoading || !user) {
    return (
      <div className="min-h-screen bg-black flex items-center justify-center">
        <p className="text-[#71767b]">Loading…</p>
      </div>
    );
  }

  const remaining = MESSAGE_LIMIT - draft.length;

  return (
    <div className="min-h-screen bg-black text-[#e7e9ea] flex justify-center">
      <div className="w-full max-w-[600px] border-x border-[#2f3336] flex flex-col h-screen">
        {/* Header */}
        <header className="sticky top-0 bg-black/80 backdrop-blur-md border-b border-[#2f3336] z-10">
          <div className="flex items-center gap-4 px-4 py-3">
            <Link
              href="/dashboard"
              className="w-9 h-9 rounded-full flex items-center justify-center hover:bg-[#181818] transition text-[#eff3f4]"
              aria-label="Back to home"
            >
              ←
            </Link>
            <div>
              <h1 className="text-xl font-bold text-[#eff3f4]">Messages</h1>
              {partnerUsername && (
                <p className="text-sm text-[#71767b]">@{partnerUsername}</p>
              )}
            </div>
          </div>
        </header>

        {/* Username box — the only way into a conversation for now */}
        <form onSubmit={handleOpenChat} className="border-b border-[#2f3336] p-4">
          <label htmlFor="chat-username" className="block text-sm text-[#71767b] mb-2">
            Message someone by username
          </label>
          <div className="flex gap-2">
            <input
              id="chat-username"
              type="text"
              value={usernameInput}
              onChange={(event) => setUsernameInput(event.target.value)}
              placeholder="@username"
              autoComplete="off"
              className="flex-1 bg-transparent border border-[#2f3336] rounded-full px-4 py-2 text-[#e7e9ea] placeholder-[#71767b] focus:outline-none focus:border-[#1d9bf0]"
            />
            <button
              type="submit"
              disabled={isOpeningChat || usernameInput.trim() === ""}
              className="px-4 py-2 rounded-full bg-[#1d9bf0] text-white font-bold disabled:opacity-50 disabled:cursor-not-allowed hover:bg-[#1a8cd8] transition"
            >
              {isOpeningChat ? "Opening…" : "Open chat"}
            </button>
          </div>
          {openChatError && (
            <p className="mt-2 text-sm text-[#f4212e]">{openChatError}</p>
          )}
        </form>

        {/* Message list */}
        <div className="flex-1 overflow-y-auto p-4 space-y-3">
          {!chat && (
            <p className="text-[#71767b] text-center py-10">
              Enter a username above to start a conversation.
            </p>
          )}

          {chat && isLoadingMessages && (
            <p className="text-[#71767b] text-center py-10">Loading messages…</p>
          )}

          {chat && messagesError && (
            <p className="text-[#f4212e] text-center py-4">{messagesError}</p>
          )}

          {chat && !isLoadingMessages && !messagesError && messages.length === 0 && (
            <p className="text-[#71767b] text-center py-10">
              No messages yet. Say hello to @{partnerUsername}.
            </p>
          )}

          {messages.map((message) => {
            const isOwn = message.sender_id === user.id;
            return (
              <div
                key={message.id}
                className={`flex items-end gap-2 ${isOwn ? "justify-end" : "justify-start"}`}
              >
                {!isOwn && <UserAvatar name={partnerUsername} size="sm" />}
                <div
                  className={`max-w-[75%] px-4 py-2 rounded-2xl ${
                    isOwn
                      ? "bg-[#1d9bf0] text-white rounded-br-sm"
                      : "bg-[#202327] text-[#e7e9ea] rounded-bl-sm"
                  }`}
                >
                  <p className="whitespace-pre-wrap break-words">{message.content}</p>
                  <p
                    className={`text-[11px] mt-1 ${
                      isOwn ? "text-white/70" : "text-[#71767b]"
                    }`}
                  >
                    {formatTime(message.created_at)}
                  </p>
                </div>
              </div>
            );
          })}

          <div ref={bottomRef} />
        </div>

        {/* Composer */}
        <form onSubmit={handleSend} className="border-t border-[#2f3336] p-4">
          {sendError && <p className="mb-2 text-sm text-[#f4212e]">{sendError}</p>}
          <div className="flex items-end gap-2">
            <textarea
              value={draft}
              onChange={(event) => setDraft(event.target.value)}
              onKeyDown={(event) => {
                // Enter sends, Shift+Enter makes a new line — the convention
                // every chat app shares
                if (event.key === "Enter" && !event.shiftKey) {
                  event.preventDefault();
                  handleSend(event);
                }
              }}
              disabled={!chat || isSending}
              rows={1}
              maxLength={MESSAGE_LIMIT}
              placeholder={chat ? "Start a new message" : "Open a chat first"}
              className="flex-1 bg-transparent resize-none text-[#e7e9ea] placeholder-[#71767b] text-lg focus:outline-none disabled:cursor-not-allowed"
            />
            <div className="flex items-center gap-3">
              {draft.length > 0 && (
                <span
                  className={`text-sm ${
                    remaining < 0 ? "text-[#f4212e]" : "text-[#71767b]"
                  }`}
                >
                  {remaining}
                </span>
              )}
              <button
                type="submit"
                disabled={!chat || isSending || draft.trim() === ""}
                className="px-4 py-2 rounded-full bg-[#1d9bf0] text-white font-bold disabled:opacity-50 disabled:cursor-not-allowed hover:bg-[#1a8cd8] transition"
              >
                {isSending ? "Sending…" : "Send"}
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>
  );
}
