# XClone

XClone is a lightweight social-media clone built with a Go API and a Next.js frontend. The app includes user auth, profile pages, post creation/editing, media uploads, comments, and 1:1 chat.

## Tech stack

- Frontend: Next.js 16, React 19, TypeScript
- Backend: Go, `net/http`, PostgreSQL, JWT cookies
- Media storage: Cloudinary
- Database orchestration: Docker Compose

## Repository layout

```text
XClone/
├── backend/                # Go HTTP API and business logic
│   ├── config/             # environment loading
│   ├── controllers/        # request handlers
│   ├── database/           # Postgres connection + schema init
│   ├── middleware/         # CORS + request logging
│   ├── models/             # domain models
│   ├── repository/         # Postgres repository implementations
│   ├── routes/             # HTTP route registration
│   ├── service/            # auth, posts, chat, media rules
│   ├── utils/              # JWT helpers and small utilities
│   ├── .env.example        # backend env template
│   ├── go.mod              # Go module
│   └── main.go             # server entrypoint
├── frontend/              # Next.js app
│   ├── app/                # App Router pages and route groups
│   ├── components/         # reusable UI blocks
│   ├── lib/                # auth helpers and shared frontend logic
│   ├── proxy.ts            # route-level redirect gate (UX behavior only)
│   ├── package.json        # frontend scripts and deps
│   └── vitest.config.ts   # test setup
├── compose.yaml            # PostgreSQL container definition
├── BUGS.md                 # current known issues
├── README.md               # project overview and setup
├── AGENTS.md               # project-specific instructions
├── LICENSE                 # MIT license
└── recovery.md            # project notes / recovery notes
```

## What is in the app

- Account signup/login/logout with JWTs stored in an HTTP-only cookie
- Profile reading and profile updates
- Feed and profile timelines for posts
- Post creation, editing, and deletion with ownership checks
- Photo uploads with Cloudinary, limited to a small number per post
- Comments on posts
- 1:1 direct messaging between users
- Basic protected/unprotected route separation in the frontend

## Environment setup

Create the repo-level Postgres env for Docker Compose:

```bash
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=xclone
```

Copy the backend template and fill in the required values:

```bash
cp backend/.env.example backend/.env
```

Required backend env keys are:

- `DATABASE_URL`
- `JWT_SECRET`
- `PORT`

Optional but used for media uploads:

- `CLOUDINARY_CLOUD_NAME`
- `CLOUDINARY_API_KEY`
- `CLOUDINARY_API_SECRET`
- `CLOUDINARY_UPLOAD_FOLDER`

## Local development

Start PostgreSQL:

```bash
docker compose up -d
```

Start the backend:

```bash
cd backend
go mod download
go run main.go
```

Start the frontend:

```bash
cd frontend
npm install
npm run dev
```

Default local ports:

- backend: `http://localhost:8080`
- frontend: `http://localhost:3000`

## Validation commands

Backend:

```bash
cd backend
go build ./...
go vet ./...
go test ./...
```

Frontend:

```bash
cd frontend
npm run lint
npx vitest run
```

Use `npx vitest run` instead of `npm test` because the default test command is watch mode and does not exit cleanly.

## Notes

- The backend enforces auth at the handler level, not in the middleware stack.
- The frontend uses route groups like `(public)` and `(protected)` only for organization; they do not change URL structure.
- The project is intentionally layered: controller -> service -> repository -> Postgres.
- Some paths still have known bugs or rough edges; see [BUGS.md](BUGS.md).

## License

This project is released under the MIT License.

