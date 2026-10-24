# Claude Instructions

@AGENTS.md

## Core Rules
- Inspect existing code and project structure before making changes.
- Prefer simple, maintainable solutions. Avoid unnecessary complexity and rewrites.
- Preserve existing functionality and user changes. Never modify unrelated files.
- Never expose secrets or perform destructive actions without approval.
- Keep the architecture: Controller → Service → Repository → PostgreSQL
- Never introduce mock or in-memory storage as a production fallback.


## Workflow
- For complex tasks, investigate and propose a short plan before editing.
- Automate repetitive work; don't make me do manually what you can safely do.
- Run relevant tests and checks. Never claim unverified success.
- Report changes, test results, and unresolved issues concisely.

## Help Me Learn
- I am learning to become an independent developer.
- Explain important decisions and unfamiliar concepts in simple words.
- Use small examples when they help understanding.
- Keep explanations concise; don't explain every line unless I ask.
- Prefer teaching through real project work rather than lengthy theory.
- When appropriate, explain the underlying problem and why the solution works.

## Project Conventions
- Database schema lives in `backend/database/postgres.go` → `Migrate()`: raw idempotent SQL re-run on every boot. There is no migration tool — do not add migration files.
- Auth is enforced per handler (read the `x_auth_token` cookie, then `utils.ValidateJWT`). `backend/middleware/` is CORS + logging only — do not move auth there.
- `frontend/proxy.ts` is Next.js 16's renamed `middleware.ts`. It decodes the JWT `exp` only and does not verify the signature: a deliberate UX gate, not a security boundary. Do not rename or "fix" it.
- Post photos live in `post_media`, keyed by `(post_id, ordinal)` with `ordinal` 0-3 — the composite primary key plus a `CHECK` are what cap a post at 4 photos, so do not rely on Go validation alone. `posts.media_url` is a legacy mirror of the first photo; `media[]` is authoritative.
- Photo uploads are two-step: `POST /api/upload/post-media` returns a `public_id`, and `POST /api/posts` accepts **only that `public_id` and builds the URL server-side**. Never let the client supply a media URL — that check in `service/post_media.go` is what stops a post pointing at a host we do not control, or at another user's upload.

## Checks
- Backend: `cd backend && go build ./... && go vet ./...`
- Frontend: `cd frontend && npm run lint`
- Frontend tests: `cd frontend && npx vitest run` — one-shot. `npm test` is bare `vitest`, i.e. watch mode, so it never exits; do not use it for an automated check.
- Test stack is Vitest + jsdom + React Testing Library, configured in `frontend/vitest.config.ts`.
- Coverage is partial: `frontend/components/post/PostCard.test.tsx` is the only test file so far. There is no backend test suite — say so instead of inventing one.
