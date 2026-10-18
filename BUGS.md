# Known Bugs & Issues

Running list of defects and rough edges in XClone. Each entry gives the location, the problem,
why it matters, and a suggested fix. Tick the box when fixed.

Last updated: 2026-10-18

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

### 6. Three different ways of counting characters
- [ ] **Location:** `frontend/app/dashboard/page.tsx:295, 308-322` vs
  `frontend/components/post/EditPostModal.tsx:28` vs `backend/service/post_service.go:40`
- **Problem:** Three counters disagree about what "280" means:
  - backend: `len([]rune(content))` — Unicode code points
  - edit modal: `[...content].length` — code points (matches the backend)
  - dashboard composer: `tweetText.length` plus `maxLength={CHAR_LIMIT}` — **UTF-16 units**, so most
    emoji count as 2
- **Why it matters:** The composer is stricter than the server. A legal 200-emoji post is blocked
  before it is ever sent, and the countdown number shown to the user is wrong for any text
  containing emoji. It fails safe (never too permissive), which is why it is Medium and not High.
- **Fix:** Use `[...tweetText].length` in the composer, drop the `maxLength` attribute (it also
  counts UTF-16 units and cannot be made Unicode-aware), and rely on the existing `disabled` guard.

### 7. The number `280` and its message are duplicated
- [ ] **Location:** `backend/service/post_service.go:14, 40`,
  `backend/controllers/post_controller.go:56`, `frontend/app/dashboard/page.tsx:22`,
  `frontend/components/post/EditPostModal.tsx:7`
- **Problem:** The limit is *enforced* in exactly one place (good), but the string
  `"post content exceeds 280 characters"` is written twice in Go, and `CHAR_LIMIT = 280` twice in
  TypeScript.
- **Why it matters:** Change the limit to 500 and the backend behaves correctly while telling users
  "exceeds 280 characters", and both frontend screens keep blocking at 280.
- **Fix:** Backend: a `const MaxPostRunes = 280` with the message built via `fmt.Sprintf`. Frontend:
  one shared constant imported by both files.

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
- **Posts are text-only.** `media_url` exists on the table, model, and response, and `CreatePost`
  always writes `""`. The Cloudinary upload path is wired for avatars only.
- **No test suite.** No `*_test.go`, no frontend test runner. Verification today is
  `go build ./... && go vet ./...`, `npm run lint`, and manual/curl checks.
