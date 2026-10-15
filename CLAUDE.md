# XClone — Claude Instructions

## Stack
- Frontend: Next.js App Router, TypeScript, Tailwind CSS v4.
- Backend: Go REST API.
- Database: PostgreSQL.

## Rules
- Inspect existing code before changing anything. Reuse existing patterns; avoid unnecessary rewrites.
- Keep the architecture: Controller → Service → Repository → PostgreSQL.
- Never introduce mock or in-memory storage as a production fallback.
- Preserve existing APIs, authentication, and functionality unless explicitly asked to change them.
- Never expose secrets or perform destructive database/Git operations without approval.
- Keep changes minimal. Do not modify unrelated files.

## Workflow
- For complex tasks: investigate, propose a short plan, and wait for approval before editing.
- For simple tasks: implement directly, then verify.
- Run relevant tests and checks. Never claim unrun tests passed.
- Report what changed, what was tested, and anything unresolved.

## Teach Me
- I am learning to become an independent developer.
- Prefer the simplest maintainable solution; avoid unnecessary abstractions.
- Briefly explain why a change is needed and important design decisions.
- Explain unfamiliar concepts in simple words, using small examples when useful.
- Do not overwhelm me with explanations of code I already understand.
- Help me learn without making me do manually what you can safely automate.