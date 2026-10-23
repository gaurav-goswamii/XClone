# XClone — Gap-Closing Plan (recovery notes)

Date: 2026-10-23

Two things settled first:
- **Do not rewrite the project.** The code is good. Retyping working code teaches far less
  than changing it — and CLAUDE.md already says avoid pointless rewrites.
- **Lower the bar to the correct one.** You don't need to write every line from scratch.
  The bar is: *"Could I change this safely, and would I notice if I broke it?"*

---

## Step 1 — Triage. Not all lines deserve equal effort.

Repo is ~6,000 lines. The parts where being wrong is **silent and serious** are ~370.

### Tier 1 — must own completely (~370 lines)

| File | Lines | Why it's Tier 1 |
|---|---|---|
| `service/post_media.go` | 140 | Trust boundary. One weakened regex = strangers' images in your feed. |
| `database/postgres.go` -> `Migrate()` | 90 | Every constraint protecting your data. Errors here corrupt data permanently. |
| `utils/auth.go` | 76 | Hashing + JWT. The whole identity system. |
| `middleware/middleware.go` | 38 | CORS — has a live **High** bug right now. |
| `auth_controller.go` cookie block | ~30 | `HttpOnly`, `SameSite`, `Secure`. |

370 lines = one focused week. That's the whole security surface of the app.

### Tier 2 — must be able to *modify* confidently
Not all of it — **one of each pattern**, known deeply:
- one controller handler, start to finish
- one service method, start to finish
- one repository method, start to finish
- one page's data flow (`useState` -> `useEffect` -> `fetch` -> render)

You have 11 handlers. You need **one** so thoroughly that the other ten are obvious.
Patterns, not inventory.

### Tier 3 — reading is genuinely fine
`frontend/` is 3,350 lines and most of it is Tailwind markup and SVG paths. Landing-page
footer, signup wizard layout, hardcoded trending widget — skimming is correct. There is no
hidden depth in `className="flex items-center gap-4"`.

> Spend effort where being wrong is expensive: ~400 lines of Go, not 3,350 lines of JSX.

---

## Step 2 — Earn Tier 1 through production, not reading

### A. The break-it sweep (one evening)
Start here: it's fast and shows where you actually stand. For each — **write your prediction
first**, then run it, then `git checkout .`

1. Delete `AND user_id = $2` from the delete query -> try deleting someone else's post
2. Change `len([]rune(x))` -> `len(x)` -> post 100 emoji
3. Remove the `^` and `$` from the media regex -> try attaching a foreign public_id
4. Remove `defer tx.Rollback()` -> force an error on photo #3, inspect the DB
5. Remove `credentials: "include"` from the `/me` fetch -> watch which page breaks, how badly

Every wrong prediction is a Tier 1 gap you now know by name.

### B. The BUGS.md campaign (your best asset — you wrote it)
13 real bugs in your own code, pre-triaged, with reasoning and suggested fixes already written.
Fixing a bug is production work: understand the surroundings, change it, prove it.

| Order | Bug | What it teaches | Size |
|---|---|---|---|
| 1 | **#4** cookie `Secure: false` | Threading config into a controller | 1 line |
| 2 | **#6** UTF-16 counter in composer | Bytes vs characters — the thing you missed, in a 2nd language | ~3 lines |
| 3 | **#10c** photos have no alt text | Accessibility; needed by the photo grid anyway | small |
| 4 | **#7** `280` duplicated | Single source of truth — the other thing you half-missed | small |
| 5 | **#5b** EditProfileModal leaks preview URLs | Browser memory, by copying a correct pattern from your own dashboard | medium |
| 6 | **#3** CORS reflects any origin | **High.** Real security. Do it with momentum behind you. | medium |
| 7 | **#5** ghost post after 404 | State shared across components — hardest of the batch | hard |

Items 2 and 4 are the exact concepts fumbled in the teaching session. Your bug list and your
gap list are the same list.

**Rule: no AI writes the fix.** AI may explain surrounding code if you're stuck. You type the
fix. After each one run:
- `cd backend && go build ./... && go vet ./...`
- `cd frontend && npm run lint`

### C. Write your first tests (the thing that most changes you)
Repo has zero tests. A test forces you to state what *correct* means — pure production.

Pure functions first (no DB, no HTTP):
`service.ValidatePassword`, `service.ValidateAge`, `service.ValidateUsername`,
`service.normalizeContent`, `service.clampDimension`

Then the one that matters most:
`mediaResolver.resolve` — table-driven, feed it every attack: foreign host, another user's ID,
`../` escapes, 5 photos, duplicate public_id, empty folder.

That test file becomes the written specification of your security model. You cannot write it
without articulating every attack — which is how you come to own `post_media.go`.

---

## Step 3 — Build the half-finished feature, using the new loop

Don't pause building to study. Same motion. First feature:

> **Render the photo grid in `PostCard`.**

Why it's ideal:
- Backend is **already done and correct** — photos upload, persist, arrive in `media[]`
- Purely frontend -> **you cannot corrupt data while learning**
- `PostCard` is used in **three** places (dashboard, profile, post page) — one change, three payoffs
- Most visible gap in the app; photo posts currently render as blank cards
- BUGS.md already predicted it lands with this work

Next natural feature: **likes** — full-stack (migration -> repository -> service -> controller ->
frontend), so it walks all three layers in one go. That's Tier 2 earned.

---

## A realistic shape (1 hour/day)

| | Focus |
|---|---|
| **Week 1** | Break-it sweep. Bugs #4, #6, #10c, #7. |
| **Week 2** | First tests (pure functions). Bug #5b. Photo grid with the full loop. |
| **Week 3** | The `resolve` test with every attack case. Then bug #3 (CORS). |
| **Week 4** | Bug #5 (ghost post). Then likes, full-stack, new loop throughout. |

End state: no High bugs, a test suite that exists, two real features — and every Tier 1 file
changed by your own hands.

---

## The one honest check along the way

After each bug fix ask: *"Could I have found this myself?"*

Early on: no — you only found them because you (with AI) wrote them down. Fine. But by bug 7
the answer should start turning into *"yes, and here are two more that aren't on the list."*

> **The day you add an entry to BUGS.md that AI didn't suggest is the day the gap is closed.**
> That's the finish line — not finishing the list.

---

## The revised learning loop (for every new feature)

1. Plan with AI. *(unchanged — this is good)*
2. **Before reading the explanation: write your guess.**
3. Understand it. Note every place your guess was wrong.
4. **Type the first instance of any new pattern yourself, from blank, no AI.**
5. Let AI implement the rest.
6. **Break it in 3 places. Predict each symptom, then verify.**
7. **Write one test for the core rule.**
8. **Gate: explain it from memory, editor closed. Stalls -> your gap list.**
   *(replaces "I understand, move on")*
9. **In two days, re-implement one small piece from scratch.**

Steps 2, 6 and 8 are nearly free. Adopting only those three fixes most of the hole.

---

## Bonus fact established in session (not in BUGS.md)

**Logout is client-side only.** `clearAuthCookie` only tells the *browser* to forget the cookie.
A token copied out of DevTools keeps working for its full 30 days — there is no revocation
mechanism anywhere in the backend (grepped: no blocklist, no token_version, no sessions table).
That is the accepted price of stateless JWT auth. The standard fix when you need revocation:
short-lived access token (15 min) + a refresh token row in the DB that logout deletes.
