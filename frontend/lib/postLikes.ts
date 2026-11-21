const API_BASE = process.env.NEXT_PUBLIC_API_URL || ""

export interface PostLikeResult {
    liked: boolean
    likes: number
}

export async function setPostLike (
    postId: string,
    liked: boolean
): Promise<PostLikeResult> {
    const response = await fetch (
        `${API_BASE}/api/posts/${encodeURIComponent(postId)}/like`,
        {
            method: liked? "PUT" : "DELETE",
            credentials: "include"
        }
    )

    const data = await response.json().catch(() => null)

if (!response.ok || data?.status !== "success" || typeof data.data?.liked !== "boolean" ||
    typeof data.data?.likes !== "number"
) {
    throw new Error(data?.message || "Failed to update post like.")
}

return {
    liked: data.data.liked,
    likes: data.data.likes
}

}