# Known Bugs & Issues

Running list of defects and rough edges in XClone. Each entry gives the location, the problem,
why it matters, and a suggested fix. Tick the box when fixed.

Last updated: 2026-10-22

---

## High

### 3. CORS reflects any `Origin` while allowing credentials
- [ ] **Location:** `backend/middleware/middleware.go:12-18`
- **Problem:** The middleware echoes whatever `Origin` the caller sent and pairs it with
  `Access-Control-Allow-Credentials: true`:
  ```go
  if origin != "" { w.Header().Set("Access-Control-Allow-Origin", origin) }
  w.Header().Set("Access-Control-Allow-Credentials", "true")
  ```
  That is an allow-list of *every site on the internet*.
- **Why it matters:** The intended guard is a fixed allow-list. Today the only thing preventing a
  malicious page from making authenticated calls on a logged-in victim's behalf is the cookie's
  `SameSite=Lax` setting, which happens to withhold the cookie on cross-site `fetch`. The CORS
  policy itself provides no protection — so changing the cookie to `SameSite=None` (needed for some
  deployment setups) would immediately open the API to any origin.
- **Fix:** Keep an explicit allow-list (`http://localhost:3000` plus the production domain, ideally
  from an env var) and reflect the origin only when it appears in that list.

---

## Medium

### 4. Cookie `Secure: false` is hardcoded
- [ ] **Location:** `backend/controllers/auth_controller.go:55, 67`
- **Problem:** `Secure: false` is a literal with a comment explaining it is for local HTTP. There is
  no environment switch, even though `cfg.Environment` exists and is already logged at startup.
- **Why it matters:** Deployed over HTTPS, the auth cookie would travel in cleartext on any
  accidental HTTP request. Easy to forget at deploy time precisely because nothing breaks locally.
- **Fix:** Derive it: `Secure: cfg.Environment == "production"`. The `AuthController` already holds
  config-derived values, so threading the environment in is small.

### 5. Edited post lingers on screen after a 404 ("ghost post")
- [ ] **Location:** `frontend/components/post/EditPostModal.tsx` (error branch), and the pages that
  hold the lists: `app/dashboard/page.tsx`, `app/[username]/page.tsx`
- **Problem:** If the post was deleted elsewhere (another tab, another device) and you then save an
  edit, the server correctly answers **404**. The modal shows "post not found" — but the page's list
  still contains the post, so after closing the modal you keep looking at something that no longer
  exists until a manual reload.
- **Why it matters:** The screen contradicts the server. Delete already handles this case
  (`if (res.ok || res.status === 404)` removes the row); edit does not.
- **Fix:** Give the modal an `onMissing` callback (or pass the status to `onUpdated`) so a 404 from
  the edit path removes the post from the list as well, then closes.

### 5b. `EditProfileModal` leaks every preview URL it creates
- [ ] **Location:** `frontend/components/profile/EditProfileModal.tsx` (`handleFileChange`)
- **Problem:** `URL.createObjectURL(file)` is called for the avatar preview and
  `URL.revokeObjectURL` is never called, so each selection pins its blob for the lifetime of the
  page. A second, related defect in the same handler: the file input's `value` is never reset, so
  re-selecting the *same* file fires no `onChange` and appears to do nothing.
- **Why it matters:** The leak is small per selection but unbounded; the input reset is a visible
  dead end for a user retrying the same file after an error.
- **Fix:** Both are already done correctly in `app/dashboard/page.tsx`'s photo composer — copy
  that pattern: revoke on removal, after a successful save, and in a `useEffect` cleanup, and set
  `e.target.value = ""` before any early return.

### 6. The composer still counts characters in UTF-16 units (partly fixed)
- [ ] **Location:** `frontend/app/dashboard/page.tsx` — the `<textarea maxLength={CHAR_LIMIT}>`
  attribute and the countdown `{CHAR_LIMIT - tweetText.length}`
- **Problem:** The submit guard and the Post button's `disabled` expression were switched to
  `[...tweetText].length` (code points, matching the backend's `len([]rune(...))`) while the photo
  composer was built. Two UTF-16 counters remain in the same file:
  - the `maxLength` attribute, which hard-stops typing at 280 **UTF-16 units**
  - the countdown number, which is wrong for any text containing emoji
- **Why it matters:** `maxLength` means a legal 280-emoji post still cannot be typed into the box,
  even though the guards and the server would both accept it. It fails safe (never too permissive),
  which is why it is Medium and not High.
- **Fix:** Drop the `maxLength` attribute (it counts UTF-16 units and cannot be made Unicode-aware)
  and use `[...tweetText].length` for the countdown. The `disabled` guard already blocks over-limit
  text, so nothing is lost by removing the hard stop.

### 7. The number `280` and its message are duplicated
- [ ] **Location:** `backend/service/post_service.go:14, 40`,
  `backend/controllers/post_controller.go:56`, `frontend/app/dashboard/page.tsx:22`,
  `frontend/components/post/EditPostModal.tsx:7`
- **Problem:** The limit is *enforced* in exactly one place (good), but the string
  `"post content exceeds 280 characters"` is written twice in Go, and `CHAR_LIMIT = 280` twice in
  TypeScript.
- **Why it matters:** Change the limit to 500 and the backend behaves correctly while telling users
  "exceeds 280 characters", and both frontend screens keep blocking at 280.
- **Fix:** Backend: `const MaxPostRunes = 280` **now exists** in `service/post_service.go`, so
  this is down to building the two message strings with `fmt.Sprintf`/`fmt.Errorf` from it — one in
  `post_service.go` (`ErrContentTooLong`) and one in `post_controller.go` (`writePostError`).
  Frontend: still one shared constant imported by both files. Note `MaxPostPhotos` in
  `service/post_media.go` is already done this way, and is the pattern to copy.

---

## Low

### 8. Out-of-range `limit` is silently rewritten
- [ ] **Location:** `backend/repository/postgres_repo.go:255, 303`
- **Problem:** `if limit <= 0 || limit > 100 { limit = 20 }`. A request for `?limit=500` returns 20
  rows with a `200 OK` and no indication that the parameter was ignored.
- **Why it matters:** A client paginating by `limit` cannot tell the difference between "you asked
  for too many" and "there were only 20". Clamping to the maximum (100) would be more predictable
  than resetting to the default.
- **Fix:** Clamp to 100 rather than 20, or reject with a 400. Also consider moving the policy to the
  service layer, where the other business rules live.

### 9. `PATCH` is accepted by a handler the browser can never reach
- [ ] **Location:** `backend/controllers/auth_controller.go:224` vs
  `backend/middleware/middleware.go:19`
- **Problem:** `UpdateProfileHandler` accepts `PUT` **or** `PATCH`, but `Access-Control-Allow-Methods`
  omits `PATCH`, so the browser's preflight blocks any cross-origin `PATCH` before it is sent.
- **Why it matters:** Dead branch, and a trap: a future `PATCH` call would fail with an opaque CORS
  console error, no server log line, and no status code to debug.
- **Fix:** Either add `PATCH` to the allow-list or drop it from the handler. Pick one.

### 10. No `.env.example`
- [ ] **Location:** repo root
- **Problem:** `.gitignore` already whitelists `!.env.example`, but the file does not exist. A new
  contributor has no way to know which variables are required.
- **Fix:** Commit `.env.example` listing `PORT`, `ENV`, `DATABASE_URL`, `JWT_SECRET`,
  `CLOUDINARY_*` with placeholder values. Pairs with issues 1 and 2.

### 10b. Abandoned composer uploads leak Cloudinary assets
- [ ] **Location:** `backend/controllers/upload_controller.go` (`UploadPostMediaHandler`), and the
  same pattern in `UploadAvatarHandler`
- **Problem:** Photos upload to Cloudinary *before* the post exists. Pick four photos, close the
  tab, and four assets remain with no database row referencing them. Deleting a post has the same
  shape: `ON DELETE CASCADE` removes the `post_media` rows, but the Cloudinary files stay.
- **Why it matters:** Unbounded storage growth, and there is **no rate limiting anywhere in the
  app**, so a single authenticated user can loop the upload endpoint and burn the account quota.
- **Fix:** Every post photo is tagged `xclone_post_media`, so a reaper can list by tag through the
  Admin API and destroy anything older than a day with no matching row. **Important:** `public_id`
  is intentionally *not* unique — the same asset may be attached to two posts — so any delete path
  must count references (`SELECT count(*) FROM post_media WHERE public_id = $1`) before calling
  `uploader.Destroy`, or deleting one post will break another post's image.

### 10c. Photos have no alt text
- [ ] **Location:** `backend/database/postgres.go` (`post_media`), and the composer in
  `frontend/app/dashboard/page.tsx`
- **Problem:** There is no `alt_text` column and no field in the composer, so a screen reader has
  nothing describing a post's photos.
- **Why it matters:** Accessibility. It is cheapest to decide alongside the rendering grid, before
  there is markup to retrofit.
- **Fix:** Add `alt_text TEXT NOT NULL DEFAULT ''` to `post_media`, a per-photo input in the
  composer, and use it for each `<img alt>` when the grid is built.

### 10d. A misspelled key in `POST /api/posts` fails silently
- [ ] **Location:** `backend/controllers/post_controller.go` (`CreatePostHandler`)
- **Problem:** The JSON decoder does not call `DisallowUnknownFields()`, so a client that sends
  `medias` instead of `media` (or `url` instead of `public_id`) gets **`201 Created` with a
  text-only post** and no indication that its photos were dropped.
- **Why it matters:** Silent data loss is worse than an error. It is also the one remaining soft
  spot in the create contract now that media references are otherwise strictly validated.
- **Fix:** `decoder := json.NewDecoder(r.Body); decoder.DisallowUnknownFields()` in
  `CreatePostHandler` only — it is a behaviour change on an existing endpoint, so it was
  deliberately held back rather than slipped in with the photo work.

### 10e. Post photos are stored under a folder named for profile pictures
- [ ] **Location:** `backend/.env` → `CLOUDINARY_UPLOAD_FOLDER` (currently
  `Xclone/ProfilePicture`), consumed by `backend/controllers/upload_controller.go:192` and
  `backend/service/post_media.go:64-65`; documented in `README.md:250`
- **Problem:** A single `CLOUDINARY_UPLOAD_FOLDER` is shared by **both** upload endpoints, and it
  is set to `Xclone/ProfilePicture`. So post photos land at
  `Xclone/ProfilePicture/posts/<userID>/<token>` — nested inside a folder named for avatars —
  while avatars land at `Xclone/ProfilePicture/avatar_<userID>_<unix>` beside them.
  Two smaller problems travel with it:
  - `README.md:250` documents the value as `xclone`, which does not match the real setting, so a
    new contributor following the README gets a different folder layout than the one in use.
  - The folder is baked into the `public_id` that `service/post_media.go` validates, so the var is
    not purely cosmetic configuration — changing it changes what counts as a valid reference.
- **Why it matters:** Nothing breaks: the paths are literal strings, `regexp.QuoteMeta` handles the
  embedded `/`, and this layout is exercised by the live checks. It is a readability and
  maintenance trap — anyone browsing the Cloudinary media library, or writing the orphan reaper
  from issue 10b, will reasonably assume `ProfilePicture/` holds profile pictures.
- **Fix:** Set `CLOUDINARY_UPLOAD_FOLDER=Xclone`. Avatars then sit at `Xclone/avatar_...` and post
  photos at `Xclone/posts/<userID>/<token>`, separated by their own prefixes, and update
  `README.md:250` to match. Safe to do: every already-stored URL lives in `users.avatar_url` and
  `post_media.url`, so existing images keep resolving — only **new** uploads move. The one edge
  case is a photo uploaded just before the change and posted just after, which would be rejected as
  an invalid reference; harmless, and the user can re-pick the file.
  If avatars and post photos should live in genuinely separate trees, the alternative is to split
  the setting into `CLOUDINARY_AVATAR_FOLDER` and `CLOUDINARY_POST_FOLDER` and thread each to its
  own consumer — more explicit, but two more env vars to keep in sync.

### 11. Backend is not `gofmt`-clean
- [ ] **Location:** `models/model.go`, `utils/utils.go`, `controllers/post_controller.go`,
  `service/post_service.go`, `repository/*.go`
- **Problem:** `gofmt -l .` lists six files. The cause is import grouping (`backend/...` placed
  after stdlib instead of sorted together) and a stray trailing tab at `postgres_repo.go:16`.
- **Why it matters:** Not a behavioral bug, but it means `gofmt`/`goimports` on save produces noisy
  diffs in files unrelated to whatever you are working on.
- **Fix:** One `gofmt -w .` pass, committed on its own so it never mixes with a feature diff.

### 12. Manual URL path parsing where a wildcard would do
- [ ] **Location:** `backend/controllers/auth_controller.go:280-287`
- **Problem:** `GetUserProfileHandler` extracts the username with
  `strings.TrimPrefix` + `strings.Trim`, including a special case for the literal `"profile"`.
  `/api/posts/{id}` now uses Go 1.22+ wildcards and `r.PathValue("id")` instead.
- **Why it matters:** Two routing styles in one codebase, and the hand-rolled one carries the
  `profile` special case as a latent trap.
- **Fix:** Register `/api/users/{username}` and read `r.PathValue("username")`. Literal patterns
  still take precedence, so `/api/users/profile` keeps working.

### 13. Unreachable guard in `PostByIDHandler`
- [ ] **Location:** `backend/controllers/post_controller.go` (top of `PostByIDHandler`)
- **Problem:** `if postID == ""` can never fire: a `{id}` wildcard does not match an empty segment,
  and `/api/posts/` returns the router's own 404 before the handler runs. Verified with a standalone
  `ServeMux` test.
- **Why it matters:** Harmless, but it is dead code that implies a case that cannot happen.
  *(Introduced while adding post CRUD — my own, not pre-existing.)*
- **Fix:** Delete the check, or keep it deliberately as a guard against future route changes.

---

## Known gaps (not bugs — unbuilt features)

These are visible in the UI but intentionally unimplemented. Listed so they are not mistaken for
regressions.

- **Like / repost / reply do nothing.** `PostCard` renders all three buttons and accepts
  `onLike`/`onRetweet`/`onReply` props, but **no page passes them**, and no endpoint exists to change
  the counters. The `likes`, `retweets`, and `replies` columns are real but permanently `0`. Clicking
  gives no feedback at all — the most user-visible gap in the app.
- **The "Following" tab** (`dashboard/page.tsx:271`) is inert. The feed is global and
  reverse-chronological; there is no follows table.
- **"What's happening" is hardcoded.** `dashboard/page.tsx:371-387` — `#NextJS16 · 24.8K posts` is
  typed into the file.
- **Photos are stored but not yet displayed.** Posts accept up to 4 photos, they upload to
  Cloudinary, and they persist in `post_media` and come back in the API's `media` array — but
  `PostCard` does not render them, so a photo post looks text-only (and a photo-only post looks
  like an empty card) in the feed, on a profile, and on the post page. **This is expected, not a
  regression**; the grid lands with the photo-CRUD work below.
- **Photo CRUD on an existing post is unimplemented.** `UpdatePost` edits `content` only. Photos
  are fixed at creation: they cannot be added, replaced, reordered, or removed afterwards. The
  caption of a photo post *can* be edited or cleared.
- **No test suite.** No `*_test.go`, no frontend test runner. Verification today is
  `go build ./... && go vet ./...`, `npm run lint`, and manual/curl checks.
