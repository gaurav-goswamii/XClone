"use client"

import Link from "next/link"
import UserAvatar from "../ui/UserAvatar"

export interface CommentResponse {
    id: string;
    post_id: string;
    user_id: string;
    author_name: string;
    author_username: string;
    author_avatar?: string;
    content: string;
    created_at: string;
}

interface CommentCardProps {
    comment: CommentResponse;
    isOwner?:boolean;
    onDelete?:(commentId: string) => void;
}

function formatRelativeTime(dateString: string): string{
    const commentDate = new Date(dateString);
    if(Number.isNaN(commentDate.getTime())){
        return "";
    }

const diffSeconds = Math.max(
    0,
    Math.floor((Date.now() - commentDate.getTime()) / 1000),
);

if (diffSeconds < 60) return "now";

const diffMinutes = Math.floor(diffSeconds / 60);
if (diffMinutes < 60) return `${diffMinutes}m`;

const diffHours = Math.floor(diffMinutes / 60);
if (diffHours < 24) return `${diffHours}h`;

const diffDays = Math.floor(diffHours / 24);
if (diffDays < 7) return `${diffDays}d`;

return commentDate.toLocaleDateString("en-US",{
    month:"short",
    day:"numeric",
});
}

export default function CommentCard({
    comment,
    isOwner=false,
    onDelete,
}: CommentCardProps) {
    const timeFormatted = formatRelativeTime(comment.created_at);

    return (
        <article
        id={`comment-${comment.id}`}
        className="flex gap-3 border-b border-[#2f3336] p-4"
        >
            <Link 
            href={`/${comment.author_username}`}
            className="shrink-0"
            aria-label={`View ${comment.author_name}'s profile`}>
                <UserAvatar
                src={comment.author_avatar}
                name={comment.author_name}
                size="sm"
                />
            </Link>

            <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2">
                <div className="flex min-w-0 flex-1 items-center gap-1.5 
                truncate text-sm">
                    <Link 
                    href={`/${comment.author_username}`}
                    className="truncate font-bold text-[#eff3f4] hover:underline">
                        @{comment.author_username}
                    </Link>

                    {timeFormatted && (
                        <>
                        <span className="text-[#71767b]"></span>
                        <time
                        dateTime={comment.created_at}
                        title={new Date(comment.created_at).toLocaleString()}
                        className="shrink-0 text-[#71767b]">
                            {timeFormatted}
                        </time>
                        </>
                    )}
                </div>
                {isOwner && onDelete && (
                    <button
                    type="button"
                    onClick={() => onDelete(comment.id)}
                    aria-label="Delete comment"
                    className="shrink-0 rounded-full px-3 py-1 text-xs text-[#f4212e]
                    transition hover:bg-[#f4212e]/10">
                        Delete
                    </button>
                )}
                </div>

                <p className="mt-1 whitespace-pre-line break-words text-[15px]
                leading-relaxed text-[#e7e9ea]">
                    {comment.content}
                </p>
            </div>
        </article>
    );
}