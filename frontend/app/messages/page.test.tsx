import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { render, screen, cleanup, fireEvent, waitFor, act } from "@testing-library/react";
import "@testing-library/jest-dom/vitest";
import MessagesPage from "./page";

(globalThis as { IS_REACT_ACT_ENVIRONMENT?: boolean }).IS_REACT_ACT_ENVIRONMENT = true;

const replace = vi.fn();
vi.mock("next/navigation", () => ({
  useRouter: () => ({ replace }),
}));

const me = { id: "usr_me", name: "Me", username: "me" };
const chat = {
  id: "chat_1",
  user1_id: "usr_me",
  user2_id: "usr_them",
  created_at: "2026-10-28T10:00:00Z",
};

function message(seq: number, senderID: string, content: string) {
  return {
    id: `message_${seq}`,
    chat_id: chat.id,
    sender_id: senderID,
    content,
    seq,
    created_at: "2026-10-28T10:00:00Z",
  };
}

function ok(body: unknown) {
  return { ok: true, json: async () => ({ status: "success", data: body }) };
}

function fail(status: number, msg: string) {
  return { ok: false, status, json: async () => ({ status: "error", message: msg }) };
}

let thread: ReturnType<typeof message>[] = [];
let openChatResponse: unknown = null;

const fetchMock = () => fetch as unknown as ReturnType<typeof vi.fn>;

beforeEach(() => {
  replace.mockClear();
  thread = [];
  openChatResponse = null;

  Element.prototype.scrollIntoView = vi.fn();

  vi.stubGlobal(
    "fetch",
    vi.fn(async (url: string, init?: RequestInit) => {
      const method = init?.method ?? "GET";

      if (url.endsWith("/api/auth/me")) return ok({ user: me });

      if (url.endsWith("/api/chats") && method === "POST") {
        return openChatResponse ?? ok({ chat });
      }

      if (url.includes("/messages") && method === "GET") {
        return ok({ messages: thread });
      }

      if (url.includes("/messages") && method === "POST") {
        const body = JSON.parse(String(init?.body));
        const sent = message(thread.length + 1, me.id, body.content);
        thread = [...thread, sent];
        return ok({ message: sent });
      }

      throw new Error(`unexpected request: ${method} ${url}`);
    })
  );
});

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

function openChatWith(username: string) {
  const box = screen.getByLabelText("Message someone by username");
  fireEvent.change(box, { target: { value: username } });
  fireEvent.click(screen.getByRole("button", { name: "Open chat" }));
}

async function tick(ms: number) {
  await act(async () => {
    await vi.advanceTimersByTimeAsync(ms);
  });
}

describe("MessagesPage", () => {
  it("opens a chat by username and shows its messages", async () => {
    thread = [message(1, "usr_them", "hey, you there?")];

    render(<MessagesPage />);
    await screen.findByLabelText("Message someone by username");
    openChatWith("them");

    expect(await screen.findByText("hey, you there?")).toBeInTheDocument();
    expect(screen.getByText("@them")).toBeInTheDocument();
  });

  it("shows the server's message when the username does not exist", async () => {
    openChatResponse = fail(404, "user not found");

    render(<MessagesPage />);
    await screen.findByLabelText("Message someone by username");
    openChatWith("nobody");

    expect(await screen.findByText("user not found")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Send" })).toBeDisabled();
  });

  it("sends a message and appends it to the thread", async () => {
    render(<MessagesPage />);
    await screen.findByLabelText("Message someone by username");
    openChatWith("them");
    await screen.findByText(/No messages yet/);

    const composer = screen.getByPlaceholderText("Start a new message");
    fireEvent.change(composer, { target: { value: "hello there" } });
    fireEvent.click(screen.getByRole("button", { name: "Send" }));

    expect(await screen.findByText("hello there")).toBeInTheDocument();
    await waitFor(() => expect(composer).toHaveValue(""));
  });

  it("picks up a reply that arrives between polls, without duplicating what is there", async () => {
    vi.useFakeTimers();
    try {
      thread = [message(1, me.id, "first")];

      render(<MessagesPage />);
      await tick(0);
      openChatWith("them");
      await tick(0);

      expect(screen.getByText("first")).toBeInTheDocument();

      thread = [...thread, message(2, "usr_them", "second")];
      await tick(3000);

      expect(screen.getByText("second")).toBeInTheDocument();
      expect(screen.getAllByText("first")).toHaveLength(1);

      await tick(9000);
      expect(screen.getAllByText("second")).toHaveLength(1);
      expect(screen.getAllByText("first")).toHaveLength(1);
    } finally {
      vi.useRealTimers();
    }
  });

  it("stops polling once the page unmounts", async () => {
    vi.useFakeTimers();
    try {
      render(<MessagesPage />);
      await tick(0);
      openChatWith("them");
      await tick(3000);

      const before = fetchMock().mock.calls.length;
      cleanup();

      expect(vi.getTimerCount()).toBe(0);

      await tick(30000);
      expect(fetchMock().mock.calls.length).toBe(before);
    } finally {
      vi.useRealTimers();
    }
  });
});
