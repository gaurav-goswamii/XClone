"use client";

import React, { useCallback, useEffect, useRef, useState } from "react";
import Link from "next/link";
import UserAvatar from "@/components/ui/UserAvatar";
import type { CurrentUser } from "@/lib/user";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "";
const MESSAGE_LIMIT = 2000;


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

async function errorFrom(res: Response, fallback: string): Promise<string> {
  const data = await res.json().catch(() => null);
  return data?.message || fallback;
}

function formatTime(iso: string): string {
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return "";
  return date.toLocaleTimeString([], { hour: "numeric", minute: "2-digit" });
}

function makeWebSocketURL(chatID: string): string {
  const base = API_BASE || window.location.origin;
  const url = new URL(base)

  url.protocol = url.protocol === "https:" ? "wss:" : "ws:"
  url.pathname = `/api/chats/${encodeURIComponent(chatID)}/ws`
  url.search = ""
  url.hash = ""

  return url.toString()
}

function mergeMessages(...groups: Message[][]): Message[] {
  const byID = new Map<string, Message>()

  for (const group of groups) {
    for (const message of group) {
      byID.set(message.id, message)
    }
  }

  return [...byID.values()].sort((a, b) => a.seq - b.seq)
}

interface MessagesViewProps {
  user: CurrentUser;
}

export default function MessagesView({ user }: MessagesViewProps) {
  const [isConnected, setIsConnected] = useState(false)
const socketRef = useRef<WebSocket | null>(null)

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

  useEffect(() => {
    const chatID = chat?.id;
    if (!chatID) return;

    let cancelled = false;
    const socket = new WebSocket(makeWebSocketURL(chatID))
    socketRef.current = socket

    setIsLoadingMessages(true)
    setMessagesError(null)
    setIsConnected(false)

    void loadMessages(chatID).then((fetched) => {
      if (cancelled) return

      setIsLoadingMessages(false)

      if (fetched === null) {
        setMessagesError("Could not load messages.")
        return
      }

      setMessagesError(null)
      setMessages((current) => mergeMessages(current, fetched))
    })

    socket.onopen = () => {
      if (cancelled || socketRef.current !== socket) return
      setIsConnected(true)
    }

    socket.onmessage = (event) => {
      if (cancelled || socketRef.current !== socket) return
      if (typeof event.data !== "string") return

      try {
        const message = JSON.parse(event.data) as Message
        if (!message.id || message.chat_id !== chatID) return

        setMessages((current) => mergeMessages(current, [message]))
      } catch {
        setSendError("Could not read a message from the chat connection.")
      }
    }

    socket.onerror = () => {
      if (cancelled || socketRef.current !== socket) return
      setIsConnected(false)
      setSendError("Chat connection failed.")
    }

    socket.onclose = () => {
      if (cancelled || socketRef.current !== socket) return
      setIsConnected(false)
      setSendError("Chat connection closed. Reopen the chat to reconnect.")
    }

    return () => {
      cancelled = true;
      if (socketRef.current === socket) {
        socketRef.current = null
      }
      socket.close()
    };
  }, [chat?.id, loadMessages]);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ block: "end" });
  }, [messages.length]);

  const handleSend = async (event: React.FormEvent) => {
    event.preventDefault();

    const content = draft.trim();
    if (!chat || !content || isSending) return;

    const socket = socketRef.current
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      setSendError("Chat is not connected. Reopen the chat and try again.")
      return
    }

    setIsSending(true);
    setSendError(null);

    try {
       
      socket.send(JSON.stringify({ content }))
      setDraft("");
    } catch {
      setSendError("Could not send the message. Please try again.");
    } finally {
      setIsSending(false);
    }
  };

  const remaining = MESSAGE_LIMIT - draft.length;

  return (
    <div className="min-h-screen bg-black text-[#e7e9ea] flex justify-center">
      <div className="w-full max-w-[600px] border-x border-[#2f3336] flex flex-col h-screen">
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
                <>
                <p className="text-sm text-[#71767b]">@{partnerUsername}</p>
                
                <p className="text-sm text-[#71767b]">
                  {isConnected? "Connected" : "Connecting or disconnected"}
                </p>
                </>
              )}
            </div>
          </div>
        </header>

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

        <form onSubmit={handleSend} className="border-t border-[#2f3336] p-4">
          {sendError && <p className="mb-2 text-sm text-[#f4212e]">{sendError}</p>}
          <div className="flex items-end gap-2">
            <textarea
              value={draft}
              onChange={(event) => setDraft(event.target.value)}
              onKeyDown={(event) => {
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
