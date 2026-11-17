"use client"

import { useState } from "react"
import type { FormEvent } from "react"
import type { CommentResponse } from "./CommentCard"

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "";

interface CommentComposerProps {
    postId: string;
    onPosted: (comment: CommentResponse) => void;
}

interface CreateCommentResponse {
    status?:string;
    message?:string;
    data?:{
        comment?: CommentResponse;
    }
}

export default function CommentComposer({
    postId,
    onPosted,
}: CommentComposerProps) {
    const [content, setContent] = useState("");
    const [isSubmitting, setIsSubmitting] = useState(false)
    const [error, setError] = useState<string | null>(null)

    const trimmedContent = content.trim();

    const handleSubmit = async (event: FormEvent<HTMLFormElement>)=> {
        event.preventDefault();
    
        if(!trimmedContent || isSubmitting) return;

        setIsSubmitting(true);
        setError(null)

        try{
            const response = await fetch(
                `${API_BASE}/api/posts/${postId}/comments`,
                {
                    method: "POST",
                    headers: {"Content-Type": "application/json"},
                    credentials: "include",
                    body: JSON.stringify({content: trimmedContent}),
                }
            )
         
            const data: CreateCommentResponse = await response.json()

            if(response.ok && data.status === "success" && data.data?.comment) {
                setContent("");
                onPosted(data.data.comment)
            }else {
                setError(data.message || "Failed to post comment");
            }
        } catch {
            setError("Network error while posting your comment")
        } finally {
            setIsSubmitting(false)
        }
    }

    return (
        <form
        onSubmit={handleSubmit}
        className="border-b border-[#2f3336] p-4">
            {error && (
                <p 
                role="alert"
                className="mb-3 text-sm text-[#f4212e]">
                    {error}
                </p>
            )}
        <textarea
        value={content}
        onChange={(event) => {
        setContent(event.target.value)
        if (error) setError(null)    
        }}
    disabled={isSubmitting}
    rows={3}
    aria-label="Write a comment"
    placeholder="Write a comment..."
    
        </form>
    )

}