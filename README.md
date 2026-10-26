# XClone

A modern, high-performance full-stack social media platform (Twitter / X clone) built with a **Next.js 16** App Router frontend and a modular, clean-architecture **Go (Golang)** backend backed by **PostgreSQL**.

---

## 📁 Project Architecture & Directory Structure

The repository is structured as a monorepo separating the Next.js client from the Go API server:

```
XClone/
├── compose.yaml                            # Docker Compose configuration for PostgreSQL 16
├── frontend/                               # Next.js 16 Client Application
│   ├── app/                                # Next.js App Router
│   │   ├── [username]/                     # Dynamic user profile page & user post timeline
│   │   │   └── page.tsx
│   │   ├── dashboard/                      # Main feed and tweet timeline
│   │   │   └── page.tsx
│   │   ├── login/                          # User login page
│   │   │   └── page.tsx
│   │   ├── post/                           # Single post permalink
│   │   │   └── [id]/
│   │   │       └── page.tsx
│   │   ├── signup/                         # User onboarding and registration
│   │   │   └── page.tsx
│   │   ├── favicon.ico                     # Application favicon
│   │   ├── globals.css                     # Global styles & design tokens
│   │   ├── layout.tsx                      # Root layout wrapper
│   │   └── page.tsx                        # Landing / welcome page
│   ├── components/                         # Modular UI Components
│   │   ├── navigation/                     # Navigation bars & account popovers
│   │   │   ├── AccountPill.tsx             # Floating user badge in sidebar
│   │   │   └── AccountPopover.tsx          # Logout & account switch dropdown
│   │   ├── post/                           # Post feed & timeline components
│   │   │   ├── EditPostModal.tsx           # Edit & delete modal for a user's own post
│   │   │   ├── PostCard.tsx                # Tweet card: relative timestamps, photo grid, actions
│   │   │   └── PostCard.test.tsx           # Vitest + React Testing Library unit tests
│   │   ├── profile/                        # User profile components
│   │   │   └── EditProfileModal.tsx        # Profile editing modal (Avatar, Banner, Bio)
│   │   └── ui/                             # Shared atomic components
│   │       └── UserAvatar.tsx              # Fallback-aware avatar component
│   ├── lib/                                # Shared frontend logic
│   │   └── postPhotos.ts                   # Photo limits, upload call & DraftPhoto type
│   ├── proxy.ts                            # Next.js route protection & auth proxy
│   ├── eslint.config.mjs                   # ESLint configuration (React 19 rules)
│   ├── next.config.ts                      # Next.js config with remote image domains
│   ├── package.json                        # Dependencies & build scripts
│   ├── postcss.config.mjs                  # PostCSS configuration
│   ├── tsconfig.json                       # TypeScript compiler configuration
│   └── vitest.config.ts                    # Vitest + jsdom test configuration
│
├── backend/                                # Go (Golang) REST API Server
│   ├── .env.example                        # Template for backend/.env — copy and fill in
│   ├── config/                             # Environment parser & server configuration
│   │   └── config.go
│   ├── controllers/                        # HTTP controllers & request handlers
│   │   ├── auth_controller.go              # Register, Login, Me, Logout, Profile handlers
│   │   ├── post_controller.go              # Post CRUD & feed handlers
│   │   ├── upload_controller.go            # Media & Cloudinary image upload handlers
│   │   └── controller.go                   # Health check & base controllers
│   ├── database/                           # Database connection & schema setup
│   │   └── postgres.go                     # PostgreSQL pooling & table migrations
│   ├── middleware/                         # HTTP middleware stack
│   │   └── middleware.go                   # CORS & structured request logging only
│   ├── models/                             # Domain models & entities
│   │   └── model.go                        # User & Post data structures
│   ├── repository/                         # Persistence abstraction layer
│   │   ├── repository.go                   # UserRepository & PostRepository interface definitions
│   │   └── postgres_repo.go                # PostgreSQL SQL implementation
│   ├── routes/                             # Route definitions & router setup
│   │   └── routes.go                       # Route registration & dependency injection
│   ├── service/                            # Core business logic layer
│   │   ├── auth_service.go                 # Registration, login, validation & hashing
│   │   ├── cloudinary_deleter.go           # Destroys stored photos a post no longer references
│   │   ├── post_media.go                   # Photo validation & public_id ownership checks
│   │   ├── post_service.go                 # Post CRUD, photo resolution & timeline logic
│   │   └── post_service_test.go            # Go unit tests for the service layer
│   ├── utils/                              # Cryptographic & utility helpers
│   │   ├── auth.go                         # JWT token creation, parsing & validation
│   │   └── utils.go                        # Utility helpers
│   ├── go.mod                              # Go module specifications
│   ├── go.sum                              # Go dependency checksums
│   └── main.go                             # Backend server initialization & entrypoint
│
├── .gitignore                              # Root git ignore rules
├── LICENSE                                 # MIT license
└── README.md                               # Project documentation
```

---

## 🛠 Tech Stack

### Frontend
- **Framework**: [Next.js 16](https://nextjs.org/) (App Router, Turbopack)
- **UI Library**: [React 19](https://react.dev/)
- **Styling**: [Tailwind CSS v4](https://tailwindcss.com/)
- **Icons**: [Lucide React](https://lucide.dev/)
- **Language**: [TypeScript](https://www.typescriptlang.org/)
- **State & Route Guards**: `proxy.ts` reads the session cookie's expiry to gate routes (a UX
  redirect, not a security boundary — every protected endpoint re-validates the JWT server-side)

### Backend
- **Language**: [Go (Golang 1.26+)](https://go.dev/)
- **Architecture**: Clean / Layered Architecture (`Controller` -> `Service` -> `Repository`)
- **Database**: PostgreSQL 16 with automated schema migration, indexed queries, and Docker Compose orchestration
- **Authentication**: JWT (`golang-jwt/jwt/v5`) stored in HTTP-only secure cookies (`x_auth_token`)
- **Password Hashing**: `golang.org/x/crypto/bcrypt`
- **Cloud Storage**: Cloudinary Go SDK for media uploads
- **Router**: Go standard library `http.ServeMux` with modular middleware chaining
- **Middleware**: CORS and structured request logging. Authentication is enforced per handler
  (read the `x_auth_token` cookie, then validate the JWT), not in the middleware chain.

---

## ✨ Features

- 🔐 **Robust Authentication**:
  - Full registration workflow validating name, username, email, password strength, and birthdate.
  - Secure login issuing signed JWT tokens inside `HttpOnly`, `SameSite=Lax` cookies (`x_auth_token`).
  - Active session endpoint (`/api/auth/me`) with automatic validation on window focus.
  - Safe logout clearing session cookies.
- 🛡️ **Edge Route Protection**:
  - Next.js Proxy middleware automatically redirects unauthenticated users trying to visit `/dashboard` to `/login`.
  - Automatically redirects authenticated users away from `/login` or `/signup` directly to `/dashboard`.
- ✍️ **Post Creation & Editing**:
  - Real-time post creation with character counter (280 characters limit, counted in Unicode code
    points so an emoji costs one character).
  - Up to **4 photos per post**, uploaded as you pick them, with per-tile progress and retry.
    A photo-only post (no caption) is valid.
  - Authenticated post submissions linked to user identity via secure JWT HTTP cookies.
  - Live posting feedback with loading indicators and client-side error handling.
- 🖊️ **Full Post CRUD with ownership checks**:
  - Edit a post's caption *and* its photos — add, replace, reorder, or remove them — from the edit
    modal on the feed, a profile, or the post's own page.
  - Delete a post; `ON DELETE CASCADE` clears its photo rows.
  - Every write re-validates ownership server-side: editing or deleting someone else's post is
    `403`, and photo references are re-resolved through the same validation a new post goes through.
  - Photos dropped by an edit or removed with a post are destroyed in Cloudinary.
    ⚠️ **Known limitation:** this path does not count how many posts still reference a photo
    before destroying it, and `post_media.public_id` is deliberately not unique. If the same
    asset is attached to two posts, editing or deleting either one breaks the other's image.
  - Single-post permalink page at `/post/[id]`.
- 📰 **Chronological Feed & Dynamic Timelines**:
  - Global chronological feed (`/api/posts/feed`) joining post records with author profile data (`name`, `username`, `avatar_url`).
  - User-specific post timeline (`/api/posts/user?username=...`) on profile pages.
  - Reusable `PostCard` component featuring relative timestamps ("now", "5m", "2h", "3d"), an
    attached-photo grid, action button counters (replies, retweets, likes), and direct profile
    navigation.
- 👤 **User Profiles & Customization**:
  - Dynamic user profile routes (`/[username]`).
  - Full profile editing modal supporting custom display names, bios, avatars, and banner images.
  - Fallback-safe avatar rendering with first-letter fallback badges.
  - Interactive profile tabs: Posts, Replies, Highlights, Media, and Likes.
- 📱 **Interactive Dashboard Layout**:
  - Responsive 3-column Twitter/X layout (Left navigation, Main tweet timeline, Right widgets).
  - Trending topics and "Who to follow" sidebar widgets.
- 🗄️ **Production PostgreSQL Persistence**:
  - Relational schema with foreign key constraints, cascading deletions, and indexed lookups (`idx_users_username_lower`, `idx_posts_created_at`, `idx_posts_user_id`).
  - Automated database migrations run at server boot.
  - Seamless local orchestration with Docker Compose.

---

## 📡 API Endpoints

### Authentication & Users
| Method | Endpoint | Description | Auth Required |
| :--- | :--- | :--- | :---: |
| `POST` | `/api/auth/signup` | Register a new user account | No |
| `POST` | `/api/auth/login` | Authenticate user & issue session cookie | No |
| `POST` | `/api/auth/logout` | Invalidate session & clear auth cookie | Yes |
| `GET` | `/api/auth/check-username` | Check availability of a username (`?username=`) | No |
| `GET` | `/api/auth/me` | Fetch currently authenticated user session via cookie | Yes |
| `GET` | `/api/users/profile` | Retrieve public user profile (`?username=`) | No |
| `PUT` | `/api/users/profile` | Update profile (name, bio, avatar) | Yes |

The profile handler also accepts `PATCH`, but `PATCH` is missing from
`Access-Control-Allow-Methods`, so the browser's preflight blocks it before it is sent. Use `PUT`.

### Posts & Timelines
| Method | Endpoint | Description | Auth Required |
| :--- | :--- | :--- | :---: |
| `POST` | `/api/posts` | Create a post, with up to 4 optional photos | Yes |
| `GET` | `/api/posts/feed` | Retrieve global chronological timeline feed (`?limit=20&offset=0`) | No |
| `GET` | `/api/posts/user` | Retrieve posts by a specific username (`?username=&limit=20&offset=0`) | No |
| `GET` | `/api/posts/{id}` | Retrieve a single post by ID | No |
| `PUT` | `/api/posts/{id}` | Update a post you own: its caption, its photos, or both | Yes |
| `DELETE` | `/api/posts/{id}` | Delete a post you own | Yes |

Note that `limit` is **not** clamped: a value above 100 or below 1 silently falls back to the
default of 20 rather than erroring, so a client cannot distinguish "you asked for too many"
from "there were only 20".

#### Creating a post with photos

Photos are attached in two steps. Upload each file to `POST /api/upload/post-media` first,
then pass the `public_id` values it returns to `POST /api/posts`:

```json
{
  "content": "optional caption",
  "media": [
    { "public_id": "xclone/posts/usr_abc/9f2c4e7a1b3d5f08", "width": 2048, "height": 1365 }
  ]
}
```

- **Up to 4 photos** per post.
- **`content` is optional when at least one photo is attached** — a photo-only post is
  valid. A post with neither text nor photos is rejected with `400`.
- **Only `public_id` is accepted; the server builds the URL.** A `url` supplied by the
  client is ignored, and a `public_id` that does not belong to the caller's own uploads is
  rejected with `400`.
- `width` and `height` are optional hints used to reserve layout space.

Every post in a response carries a `media` array, ordered by `ordinal` (0-3), which is
always present and never `null`:

```json
"media": [
  { "ordinal": 0, "url": "https://res.cloudinary.com/...", "public_id": "xclone/posts/...",
    "media_type": "image", "width": 2048, "height": 1365 }
]
```

The older `media_url` field is **legacy**: it mirrors the first photo's URL for backwards
compatibility. `media[]` is authoritative.

| Error | Status |
| :--- | :---: |
| No text and no photos | `400` |
| More than 4 photos | `400` |
| `public_id` not one of the caller's own uploads | `400` |
| Request body over 64KB | `413` |
| Cloudinary not configured (only when photos are attached) | `503` |

#### Editing a post

`PUT /api/posts/{id}` rewrites a post you own. It takes the same body as create, and the `media`
field is **tri-state** — this is the part worth reading carefully:

| `media` in the request body | Result |
| :--- | :--- |
| Field **omitted** (or `null`) | Photos are left exactly as they are; only the caption changes |
| `[]` (empty array) | **All photos are removed** |
| `[{ "public_id": ... }, ...]` | The post ends up with precisely these photos, in this order |

```json
{
  "content": "updated caption",
  "media": [
    { "public_id": "xclone/posts/usr_abc/9f2c4e7a1b3d5f08", "width": 2048, "height": 1365 }
  ]
}
```

- Photos are **re-resolved, not copied**: every `public_id` goes through the same ownership and
  shape validation as a brand new upload, so an edit cannot smuggle in a reference that a create
  would have rejected.
- Reordering is free — the array's order becomes the new `ordinal`, and photos are matched by
  `public_id`, so moving a photo does not delete and re-upload it.
- A photo **dropped** by the edit is destroyed in Cloudinary once the database write commits.
- The post must still carry something afterwards: clearing the text *and* the photos is `400`.
- Because an omitted `media` key means "keep the photos", a **misspelled** key (`medias`) silently
  keeps them instead of removing them: the decoder does not reject unknown fields. Spell the key
  correctly and send `[]` to clear a post's photos.

`DELETE /api/posts/{id}` removes the post and, by cascade, its `post_media` rows, then destroys
the photos in Cloudinary. It responds `200` with the deleted `id`.

Shared by both:

| Error | Status |
| :--- | :---: |
| Missing or invalid auth cookie | `401` |
| Post does not exist | `404` |
| Post belongs to another user | `403` |
| Caption over 280 characters, over 4 photos, or an invalid `public_id` | `400` |
| Nothing left after the edit (no text and no photos) | `400` |
| Request body over 64KB | `413` |

### Media & System
| Method | Endpoint | Description | Auth Required |
| :--- | :--- | :--- | :---: |
| `POST` | `/api/upload/avatar` | Upload an avatar image to Cloudinary (multipart, field `avatar`) | Yes |
| `POST` | `/api/upload/post-media` | Upload one post photo to Cloudinary (multipart, field `photo`) | Yes |
| `GET` | `/api/health` | Health check probe endpoint | No |

#### Upload endpoint details

Both upload endpoints take a single file as `multipart/form-data` and talk to Cloudinary
directly; neither writes to the database. The returned reference is persisted by a separate
request (`PUT /api/users/profile` for an avatar, `POST /api/posts` for a photo).

| | `/api/upload/avatar` | `/api/upload/post-media` |
| :--- | :--- | :--- |
| Form field | `avatar` | `photo` |
| Max size | 5MB | 5MB |
| Accepted types | JPEG, PNG, WEBP, GIF | JPEG, PNG, WEBP, GIF |
| Transformation | `c_thumb,g_face,w_400,h_400` (square face crop) | `c_limit,w_2048,h_2048` (aspect ratio preserved) |
| Returns | `avatar_url` | `public_id`, `width`, `height`, `url` |

The type is determined by **sniffing the file's first bytes**, not its extension or the
client's `Content-Type` — so a text file renamed `.jpg` is rejected. HEIC/HEIF is not
recognised and is rejected; iOS normally converts to JPEG on upload.

Post photos are re-encoded at upload time and capped at 2048px on the long edge. Because
Cloudinary re-encodes, the stored image has its EXIF orientation applied and its metadata
(including GPS coordinates) stripped.

| Error | Status |
| :--- | :---: |
| Missing or invalid auth cookie | `401` |
| Wrong HTTP method | `405` |
| Missing file, or an unaccepted image type | `400` |
| File larger than 5MB | `413` |
| Cloudinary credentials not configured | `503` |
| Cloudinary rejected the upload | `502` |

---

## ⚙️ Environment Configuration

No real credentials are committed to this repository. All three files below are gitignored;
`backend/.env.example` is the committed template.

### Backend (`backend/.env`)

Copy the template and fill in your own values:

```bash
cp backend/.env.example backend/.env
```

| Variable | Required | Notes |
| :--- | :---: | :--- |
| `DATABASE_URL` | **Yes** | The server calls `log.Fatal` at boot if unset. Must match the `POSTGRES_*` values below. |
| `JWT_SECRET` | **Yes** | Signs session JWTs. The server calls `log.Fatal` at boot if unset. Generate with `openssl rand -base64 32`. |
| `PORT` | **Yes** | Not defaulted in code: if unset, the server binds an arbitrary free port and the frontend cannot reach it. |
| `ENV` | No | `development` or `production`; logged at boot. |
| `CLOUDINARY_CLOUD_NAME` | No | All three Cloudinary values are needed for uploads. |
| `CLOUDINARY_API_KEY` | No | Without them the app still runs, but avatar and photo uploads answer `503` and a warning is logged at boot. |
| `CLOUDINARY_API_SECRET` | No | |
| `CLOUDINARY_UPLOAD_FOLDER` | No | Folder prefix shared by **both** upload endpoints. Not purely cosmetic: it is baked into the `public_id` that the posts API validates against, so changing it changes which references are accepted. Existing images keep resolving (their full URLs are stored), but a photo uploaded just before a change and posted just after is rejected. |

### Database (`.env` at the repository root)

`compose.yaml` reads these three variables to initialise the PostgreSQL container, so Docker
Compose needs a `.env` in the repository root (the same directory as `compose.yaml`):

```env
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=xclone
```

These must agree with the credentials in `DATABASE_URL`.

### Frontend (`frontend/.env.local`)

Create a `.env.local` file in the `frontend/` folder. `NEXT_PUBLIC_API_URL` must point at the
backend's `PORT`:

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
```

---

## 🚀 Getting Started

### 1. Prerequisites
- **Node.js**: v20.9.0 or higher (required by Next.js 16)
- **npm**: v9.0.0 or higher
- **Go**: v1.26.0 or higher (set by `backend/go.mod`)
- **Docker & Docker Compose**: (Recommended for PostgreSQL) or local PostgreSQL v14+

---

### 2. Database Setup

Create the root `.env` described above first — `compose.yaml` reads `POSTGRES_USER`,
`POSTGRES_PASSWORD`, and `POSTGRES_DB` from it, and the container will not initialise
correctly if they are empty. Then start the PostgreSQL 16 container:

```bash
docker compose up -d
```

To verify the database container status:
```bash
docker compose ps
```

---

### 3. Backend Setup

```bash
cd backend

# Verify and download dependencies
go mod download

# Start the Go API server
go run main.go
```

The server listens on `http://localhost:8080`.

Verify health check:
```bash
curl http://localhost:8080/api/health
```

---

### 4. Frontend Setup

```bash
cd frontend

# Install dependencies
npm install

# Start the Next.js development server
npm run dev
```

The frontend application runs on `http://localhost:3000`.

---

## 🧪 Quality Assurance & Pre-commit Verification

Before submitting changes, run the linters and tests:

```bash
# Backend: compile, vet, and test
cd backend
go build ./...
go vet ./...
go test ./...

# Frontend: lint, test, and production build
cd ../frontend
npm run lint
npx vitest run
npm run build
```

Use `npx vitest run` for a one-shot run. `npm test` is bare `vitest`, which starts an
interactive watch process and never exits — it is not suitable for CI or scripted checks.

**Test coverage is thin.** There are two test files:

| File | Covers |
| :--- | :--- |
| `backend/service/post_service_test.go` | `droppedPublicIDs` — which photos an edit removes |
| `frontend/components/post/PostCard.test.tsx` | Whether a post renders an attached photo or falls back to text-only |

No controller, repository, or end-to-end test exists for the create/update/delete paths, and the
Cloudinary release path has no test at all. The repository-layer interfaces in
`backend/repository/repository.go` exist partly to make that straightforward to add.

`gofmt` is **not** currently clean (`gofmt -l backend/` lists eleven files) and `npm run lint`
emits one warning (`@next/next/no-img-element`, from the photo grid). Neither is wired into the
checks above, so both are expected output rather than a regression you introduced.

---

## 📄 License
This project is open source and available under the [MIT License](LICENSE).
