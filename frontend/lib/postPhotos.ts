
const API_BASE = process.env.NEXT_PUBLIC_API_URL || "";

export const MAX_PHOTOS = 4;
export const MAX_PHOTO_BYTES = 5 * 1024 * 1024;
export const ACCEPTED_PHOTO_TYPES = [
  "image/jpeg",
  "image/png",
  "image/webp",
  "image/gif",
];

export interface DraftPhoto {
  key: string;
  previewUrl: string;
  status: "uploading" | "done" | "error";
  error?: string;
  public_id?: string;
  width?: number;
  height?: number;
}

export function rejectPhotoFile(file: File): string | null {
  if (!ACCEPTED_PHOTO_TYPES.includes(file.type)) {
    return "Photos must be JPEG, PNG, WEBP, or GIF.";
  }
  if (file.size > MAX_PHOTO_BYTES) {
    return "Each photo must be 5MB or smaller.";
  }
  return null;
}

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

export async function uploadPostPhoto(file: File): Promise<PhotoUploadResult> {
  try {
    const formData = new FormData();
    formData.append("photo", file);

    const res = await fetch(`${API_BASE}/api/upload/post-media`, {
      method: "POST",
      credentials: "include",
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
