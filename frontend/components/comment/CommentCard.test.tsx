import {describe, it, expect} from "vitest"
import {render} from "@testing-library/react"
import CommentCard from "./CommentCard"
import "@testing-library/jest-dom/vitest"

it("renders the comment date in the time title", () => {
    const createdAt = "2026-10-24T10:00:00Z"

    const { container } = render(
        <CommentCard
        comment={{
            id:"1",
            post_id:"post-1",
            user_id:"user-1",
            author_name:"Test User",
            author_username:"testuser",
            content: "Test comment",
            created_at: createdAt,
        }}
        />
    )

    expect(container.querySelector("time")).toHaveAttribute(
        "title",
        new Date(createdAt).toLocaleString()
    )
})