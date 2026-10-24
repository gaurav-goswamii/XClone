import { describe, it, expect } from "vitest";
import { render, screen, cleanup } from "@testing-library/react";
import { afterEach } from "vitest";
import PostCard from "./PostCard";
import "@testing-library/jest-dom/vitest";

afterEach(() => {
  cleanup();
});

const post = {
  id: "1z2b",
  user_id: "3er4t5y56yy6",
  author_name: "DF",
  author_username: "fdsf",
  likes: 2,
  retweets: 1,
  replies: 0,
  created_at: "2026-10-24T10:00:00Z",
  author_avatar: "hero-pic.jpg",
  content: "In my life there is Light",
  media: [
  {
    ordinal: 0,
    url: "/test-image.jpg",
    public_id: "test-image",
    media_type: "image",
  },
],
};

const Textpost = {
  id: "1z2b",
  user_id: "3er4t5y56yy6",
  author_name: "DF",
  author_username: "fdsf",
  likes: 2,
  retweets: 1,
  replies: 0,
  created_at: "2026-10-24T10:00:00Z",
  author_avatar: "hero-pic.jpg",
  content: "In my life there is Light",
};

describe("PostCard", () => {
  it("displays an image when the post contains media", () => {
    render(<PostCard post={post} />);
    expect(
  screen.getByRole("img", { name: "Post Image" })
).toBeInTheDocument();
  });
});

describe("PostCard", ()=> {
  it("if media not present in post than text only post", () => {
     render(<PostCard post={Textpost} />);
     expect(
      screen.getByText("In my life there is Light")
     ).toBeInTheDocument();
  })
})