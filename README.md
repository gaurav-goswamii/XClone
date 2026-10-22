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
│   │   │   └── PostCard.tsx                # Tweet card with relative timestamps & actions
│   │   ├── profile/                        # User profile components
│   │   │   └── EditProfileModal.tsx        # Profile editing modal (Avatar, Banner, Bio)
│   │   └── ui/                             # Shared atomic components
│   │       └── UserAvatar.tsx              # Fallback-aware avatar component
│   ├── proxy.ts                            # Next.js route protection & auth proxy
│   ├── eslint.config.mjs                   # ESLint configuration (React 19 rules)
│   ├── next.config.ts                      # Next.js config with remote image domains
│   ├── package.json                        # Dependencies & build scripts
│   ├── postcss.config.mjs                  # PostCSS configuration
│   └── tsconfig.json                       # TypeScript compiler configuration
│
├── backend/                                # Go (Golang) REST API Server
│   ├── config/                             # Environment parser & server configuration
│   │   └── config.go
│   ├── controllers/                        # HTTP controllers & request handlers
│   │   ├── auth_controller.go              # Register, Login, Me, Logout, Profile handlers
│   │   ├── post_controller.go              # Tweet creation & feed handlers
│   │   ├── upload_controller.go            # Media & Cloudinary image upload handlers
│   │   └── controller.go                   # Health check & base controllers
│   ├── database/                           # Database connection & schema setup
│   │   └── postgres.go                     # PostgreSQL pooling & table migrations
│   ├── middleware/                         # HTTP middleware stack
│   │   └── middleware.go                   # CORS, Structured Logger, JWT Authentication
│   ├── models/                             # Domain models & entities
│   │   └── model.go                        # User & Post data structures
│   ├── repository/                         # Persistence abstraction layer
│   │   ├── repository.go                   # UserRepository & PostRepository interface definitions
│   │   └── postgres_repo.go                # PostgreSQL SQL implementation
│   ├── routes/                             # Route definitions & router setup
│   │   └── routes.go                       # Route registration & dependency injection
│   ├── service/                            # Core business logic layer
│   │   ├── auth_service.go                 # Registration, login, validation & hashing
│   │   └── post_service.go                 # Post creation & timeline retrieval logic
│   ├── utils/                              # Cryptographic & utility helpers
│   │   ├── auth.go                         # JWT token creation, parsing & validation
│   │   └── utils.go                        # Utility helpers
│   ├── go.mod                              # Go module specifications
│   ├── go.sum                              # Go dependency checksums
│   └── main.go                             # Backend server initialization & entrypoint
│
├── .gitignore                              # Root git ignore rules
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
- **State & Route Guards**: Next.js Proxy/Middleware with JWT cookie evaluation

### Backend
- **Language**: [Go (Golang 1.23+)](https://go.dev/)
- **Architecture**: Clean / Layered Architecture (`Controller` -> `Service` -> `Repository`)
- **Database**: PostgreSQL 16 with automated schema migration, indexed queries, and Docker Compose orchestration
- **Authentication**: JWT (`golang-jwt/jwt/v5`) stored in HTTP-only secure cookies (`x_auth_token`)
- **Password Hashing**: `golang.org/x/crypto/bcrypt`
- **Cloud Storage**: Cloudinary Go SDK for media uploads
- **Router**: Go standard library `http.ServeMux` with modular middleware chaining

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
- ✍️ **Post & Tweet Creation**:
  - Real-time tweet creation with character counter (280 characters limit).
  - Authenticated post submissions linked to user identity via secure JWT HTTP cookies.
  - Live posting feedback with loading indicators and client-side error handling.
- 📰 **Chronological Feed & Dynamic Timelines**:
  - Global chronological feed (`/api/posts/feed`) joining post records with author profile data (`name`, `username`, `avatar_url`).
  - User-specific post timeline (`/api/posts/user?username=...`) on profile pages.
  - Reusable `PostCard` component featuring relative timestamps ("now", "5m", "2h", "3d"), action button counters (replies, retweets, likes), and direct profile navigation.
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
| `PUT` / `PATCH` | `/api/users/profile` | Update profile (name, bio, avatar) | Yes |

### Posts & Timelines
| Method | Endpoint | Description | Auth Required |
| :--- | :--- | :--- | :---: |
| `POST` | `/api/posts` | Create a post, with up to 4 optional photos | Yes |
| `GET` | `/api/posts/feed` | Retrieve global chronological timeline feed (`?limit=20&offset=0`) | No |
| `GET` | `/api/posts/user` | Retrieve posts by a specific username (`?username=&limit=20&offset=0`) | No |

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

### Backend Configuration (`backend/.env`)
Create a `.env` file in the `backend/` folder:

```env
PORT=8080
ENV=development
DATABASE_URL=postgres://postgres:postgres@localhost:5432/xclone?sslmode=disable
JWT_SECRET=your-secure-jwt-secret-key
CLOUDINARY_CLOUD_NAME=your_cloud_name
CLOUDINARY_API_KEY=your_api_key
CLOUDINARY_API_SECRET=your_api_secret
CLOUDINARY_UPLOAD_FOLDER=xclone
```

### Frontend Configuration (`frontend/.env.local`)
Create a `.env.local` file in the `frontend/` folder:

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
```

---

## 🚀 Getting Started

### 1. Prerequisites
- **Node.js**: v18.0.0 or higher
- **npm**: v9.0.0 or higher
- **Go**: v1.23.0 or higher
- **Docker & Docker Compose**: (Recommended for PostgreSQL) or local PostgreSQL v14+

---

### 2. Database Setup

Start the PostgreSQL 16 container using the provided `compose.yaml`:

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

Before submitting changes, run the test suites and linters:

```bash
# Verify backend Go code
cd backend
go vet ./...
go build -o /dev/null ./...

# Verify frontend Next.js code
cd ../frontend
npm run lint
npm run build
```

---

## 📄 License
This project is open source and available under the [MIT License](LICENSE).
