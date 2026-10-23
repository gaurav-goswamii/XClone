# Plan — render the photo grid in PostCard

Date: 2026-10-23
Scope: frontend only. No backend change, no database change, no new dependency.

---

## Step 0 — What already works (do not rebuild it)

Check these off before writing anything, so you know the only missing piece is display:

- `backend/service/post_media.go` validates photo references and builds the Cloudinary URL.
- `models.PostResponse.Media` is a `[]PostMedia` with `json:"media"` — no `omitempty`, so the
  API always sends `media`, as `[]` when there are none.
- Each photo carries `ordinal` (0-3), `url`, `public_id`, and optional `width` / `height`.
- `frontend/app/dashboard/page.tsx` already uploads photos and sends `public_id`s on create.
- `frontend/components/post/PostCard.tsx:11-17` **already declares the `PostMedia` interface**
  and `media?: PostMedia[]` on `PostResponse` (line 33).

So the gap is exactly one thing: `PostCard` renders `post.content` and never looks at
`post.media`. A photo-only post renders as a blank card today.

One change pays off three times — `PostCard` is used by `app/dashboard/page.tsx`,
`app/[username]/page.tsx`, and `app/post/[id]/page.tsx`.

---

## Step 1 — Decide the layout rules first (the only real design work)

Everything after this is typing. X uses a different shape per photo count:

| Photos | Shape |
|---|---|
| 1 | one image, its own aspect ratio, height capped |
| 2 | two columns, side by side, equal halves |
| 3 | one tall image on the left, two stacked on the right |
| 4 | 2 x 2 |

In Tailwind that is one grid container whose classes change with the count:

- container, always: `grid gap-0.5 rounded-2xl overflow-hidden border border-[#2f3336]`
  (`gap-0.5` is the thin dark seam between photos; `overflow-hidden` is what makes the
  rounded corners actually clip the images)
- 2 photos: `grid-cols-2`
- 3 photos: `grid-cols-2 grid-rows-2`, and the **first** child gets `row-span-2`
- 4 photos: `grid-cols-2 grid-rows-2`
- every image: `w-full h-full object-cover`

**Why `object-cover` matters:** photos arrive in all shapes. `object-cover` fills the box and
crops the overflow, so the grid stays a tidy rectangle. Without it the images stretch and
faces look wrong. This is the single class doing the most work in the whole feature.

**Height.** For 2, 3, and 4 photos give the container a fixed `aspect-video` (16/9) so the
grid is the same height for every post — that is what makes a feed feel calm. For a single
photo, use its real ratio when the API sent `width` and `height`:
`style={{ aspectRatio: `${w} / ${h}` }}`, plus a `max-h-[510px]` cap so a very tall photo
cannot take over the screen. Fall back to `aspect-video` when the dimensions are missing
(the API omits them when unknown — `omitempty` on `Width`/`Height`).

**Why bother with the ratio?** Reserving the right space before the image loads stops the
text below from jumping down when it arrives. That jump is called layout shift.

---

## Step 2 — Create one new component

New file: `frontend/components/post/PostPhotoGrid.tsx`

- Start it with `"use client";` to match every other component in the folder.
- Import the existing type instead of redeclaring it:
  `import { PostMedia } from "./PostCard";`
- Props: `{ media: PostMedia[]; authorName: string }` — `authorName` only feeds the `alt`
  text in Step 6.
- No `useState`, no `useEffect`. It takes data and returns markup. Keeping it stateless is
  why it is safe to render three times per screen.

**Why a separate file and not inline in `PostCard`?** `PostCard` is already ~290 lines. The
grid has its own rules (counts, ratios, cropping) that have nothing to do with the author
header or the action buttons. Separate file = you can read either one without the other.

---

## Step 3 — Normalize the input before rendering it

At the top of the component, turn the untrusted-ish array into a clean one, in this order:

1. `if (!media || media.length === 0) return null;` — render nothing, not an empty box.
   This is also what protects you from a response cached by an older build that has no
   `media` key at all.
2. Drop rows with a missing or blank `url`.
3. Sort by `ordinal` ascending — display order is the server's decision, and you should not
   depend on the array arriving pre-sorted.
4. `.slice(0, 4)` — the database caps a post at 4 photos via the composite primary key on
   `(post_id, ordinal)`, so this is belt-and-braces, not the real limit.
5. If the list is now empty, return `null`.

Then read `const count = photos.length;` once and pick your classes from it.

**Why defend against data your own server produced?** Because the component cannot tell a
fresh response from a stale cached one, and `null` is a much better failure than a broken
layout.

---

## Step 4 — Pick the classes from the count

Keep it boring and readable — a small lookup, not clever math:

```
const gridClass =
  count === 1 ? "" :
  count === 2 ? "grid-cols-2" :
                "grid-cols-2 grid-rows-2";
```

and for the 3-photo case, the first child needs `row-span-2`. Compute that per item:

```
const spanClass = count === 3 && index === 0 ? "row-span-2" : "";
```

Use `key={photo.public_id}` on each item — stable and unique per photo. (`ordinal` works too,
but `public_id` reads better and is what the backend considers the photo's identity.)

Use a plain `<img>`, not `next/image`. That matches `UserAvatar.tsx` and the composer
preview grid, and it means you do **not** have to add Cloudinary to `remotePatterns` in
`next.config.ts`. Add `loading="lazy"` like `UserAvatar` does, so photos further down the
feed only download when scrolled near.

Keep the ESLint comment above each `<img>`:
`{/* eslint-disable-next-line @next/next/no-img-element */}` — the repo's lint config warns
on raw `<img>`, and this is how the two existing cases silence it.

---

## Step 5 — Wire it into PostCard (small, and one related tidy-up)

In `PostCard.tsx`, the body currently goes:

    author header  ->  <p>{post.content}</p>  ->  action buttons bar

Insert the grid **between the text and the action bar**, wrapped in `<div className="mt-3">`.

While you are there, one honest defect you will see immediately: for a photo-only post,
`post.content` is an empty string, so that `<p>` renders an empty element with `mt-1`
spacing. Guard it — render the paragraph only when `post.content` is non-empty. The server
already allows photo-only posts (`EditPostModal.tsx:36` relies on exactly that), so this
case is real, not hypothetical.

**Optional decision — legacy posts.** `posts.media_url` is a legacy mirror of the first
photo. If any post predates the `post_media` table, it will have `media_url` set and an
empty `media[]`. Recommendation: handle it with a tiny fallback in `PostCard` — if `media`
is empty but `media_url` is set, pass `[{ ordinal: 0, url: post.media_url, public_id:
post.media_url }]`. Two lines, and it means no old post silently loses its photo. Skip it if
you check the table and no such rows exist.

---

## Step 6 — Decide alt text now, not later

There is no `alt_text` column yet (BUGS.md #10c), and that bug explicitly says to decide
this alongside the grid, before there is markup to retrofit.

Do this now: build the alt string in **one place** in the grid component —
`alt={`Photo ${index + 1} of ${count} posted by ${authorName}`}` — mirroring the composer's
existing `Selected photo ${index + 1} of ${draftPhotos.length}` wording.

A generic alt is not great accessibility, but it is honest and it costs one line to improve
later: when `alt_text` ships, that single expression becomes
`photo.alt_text || <the generic fallback>`. The point of doing it now is that there is
exactly one place to change.

---

## Step 7 — Checks and manual verification

Automated:

- `cd frontend && npm run lint`
- Backend is untouched, so no Go build is needed. There is no test suite in this repo, so
  there are no tests to run — say that rather than inventing one.

By hand (start the app, then post one of each):

- 1, 2, 3, and 4 photos — confirm each count gets the shape from the Step 1 table
- a photo-only post (no text) — no empty gap above the grid
- a text-only post — no empty box, no stray border
- a very tall photo and a very wide photo — neither breaks the card
- the same post on all three pages: dashboard, `/<username>`, `/post/<id>`
- narrow the window to phone width — the grid should shrink, not overflow
- open the edit modal on a photo post and save a text change — photos should survive

---

## Step 8 — Break it in three places (your loop, step 6)

Write your prediction first, then run it, then `git checkout .`:

1. Remove `object-cover` from the images -> predict what 4 mixed-shape photos look like
2. Remove the `ordinal` sort -> predict whether order visibly changes, and why it might not
3. Remove `overflow-hidden` from the container -> predict what happens to the rounded corners

A good fourth: delete the `key` prop and read the console warning carefully.

---

## Step 9 — Deliberately not in this change

- **Lightbox / click-to-expand.** Needs state, keyboard handling, and focus trapping. It is
  a feature of its own, not part of rendering.
- **`alt_text` column** (BUGS.md #10c) — migration + composer input + backend plumbing.
- **Per-photo broken-image fallback.** `UserAvatar` does this with `useState` + `onError`;
  doing it per photo means tracking which ordinals failed. Leave the grid stateless for now.
- **Cloudinary transformations** (`w_`, `h_` in the URL for right-sized downloads). The
  delivery URL already carries `f_auto,q_auto`, which is the big win. Sizing is a later
  performance pass.
