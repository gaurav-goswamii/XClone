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

## Checks
- Backend: `cd backend && go build ./... && go vet ./...`
- Frontend: `cd frontend && npm run lint`
- There is no test suite yet, so there are no tests to run. Say so instead of inventing one.
