// Shared rules and plumbing for the photos attached to a post.
//
// The composer (create) and the edit modal (update) both pick files, upload
// them, and send the resulting public IDs. Keeping the limits and the upload
// call in one place means the two screens cannot drift apart, and that they
// stay in step with the server's own limits.

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "";

export const MAX_PHOTOS = 4;
export const MAX_PHOTO_BYTES = 5 * 1024 * 1024;
// Mirrors the server's allow-list so the usual mistake costs no round trip.
// The narrow list also makes iOS transcode HEIC to JPEG on pick.
export const ACCEPTED_PHOTO_TYPES = [
  "image/jpeg",
  "image/png",
  "image/webp",
  "image/gif",
];

// A photo in an unsent post: either one already attached to it, or one the
// user has just picked. previewUrl is a blob: URL while the file is local and
// the delivery URL once the photo lives on the server; public_id is the only
// part the API cares about, and only exists once the upload has finished.
export interface DraftPhoto {
  key: string;
  previewUrl: string;
  status: "uploading" | "done" | "error";
  error?: string;
  public_id?: string;
  width?: number;
  height?: number;
}

// rejectPhotoFile returns the message to show for a file the server would
// refuse, or null when the file is fine.
export function rejectPhotoFile(file: File): string | null {
  if (!ACCEPTED_PHOTO_TYPES.includes(file.type)) {
    return "Photos must be JPEG, PNG, WEBP, or GIF.";
  }
  if (file.size > MAX_PHOTO_BYTES) {
    return "Each photo must be 5MB or smaller.";
  }
  return null;
}

// toMediaPayload turns the photos ready to send into the array the posts API
// expects. Photos still uploading or failed are dropped: the post carries only
// what actually reached the server.
export function toMediaPayload(photos: DraftPhoto[]) {
  return photos
    .filter((photo) => photo.status === "done" && photo.public_id)
    .map((photo) => ({
      public_id: photo.public_id,
      width: photo.width,
      height: photo.height,
    }));
}

export type PhotoUploadResult =
  | { ok: true; public_id: string; width?: number; height?: number }
  | { ok: false; error: string };

// uploadPostPhoto sends one file to the upload endpoint and returns the
// reference the posts API accepts. It never throws: a network failure is just
// another result the caller renders on that photo's tile.
export async function uploadPostPhoto(file: File): Promise<PhotoUploadResult> {
  try {
    const formData = new FormData();
    formData.append("photo", file);

    const res = await fetch(`${API_BASE}/api/upload/post-media`, {
      method: "POST",
      credentials: "include",
      // No Content-Type header: the browser has to set the multipart boundary
      body: formData,
    });

    const data = await res.json().catch(() => null);

    if (!res.ok || data?.status !== "success" || !data?.data?.public_id) {
      return {
        ok: false,
        error:
          res.status === 413
            ? "That photo is larger than 5MB."
            : data?.message || "Upload failed. Please try again.",
      };
    }

    return {
      ok: true,
      public_id: data.data.public_id,
      width: data.data.width,
      height: data.data.height,
    };
  } catch {
    return { ok: false, error: "Network error while uploading." };
  }
}
