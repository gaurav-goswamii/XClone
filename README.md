# XClone

A modern, high-performance full-stack social media platform (Twitter / X clone) built with a **Next.js 16** App Router frontend and a modular, clean-architecture **Go (Golang)** backend.

---

## 📁 Project Architecture & Directory Structure

The repository is structured as a monorepo separating the Next.js client from the Go API server:

```
XClone/
├── frontend/                               # Next.js 16 Client Application
│   ├── app/                                # Next.js App Router
│   │   ├── [username]/                     # Dynamic user profile page
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
│   │   ├── upload_controller.go            # Media & Cloudinary image upload handlers
│   │   └── controller.go                   # Health check & base controllers
│   ├── database/                           # Database connection & schema setup
│   │   └── postgres.go                     # PostgreSQL pooling & table migrations
│   ├── middleware/                         # HTTP middleware stack
│   │   └── middleware.go                   # CORS, Structured Logger, JWT Authentication
│   ├── models/                             # Domain models & entities
│   │   └── model.go                        # User & Post data structures
│   ├── repository/                         # Persistence abstraction layer
│   │   ├── repository.go                   # UserRepository interface definition
│   │   ├── postgres_repo.go                # PostgreSQL SQL implementation
│   │   └── memory_repo.go                  # Thread-safe in-memory fallback implementation
│   ├── routes/                             # Route definitions & router setup
│   │   └── routes.go                       # Route registration & dependency injection
│   ├── service/                            # Core business logic layer
│   │   ├── auth_service.go                 # Registration, login, validation & hashing
│   │   └── service.go                      # Base service interfaces
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
- **Database**: PostgreSQL (with automated schema migration) and In-Memory fallback
- **Authentication**: JWT (`golang-jwt/jwt/v5`) stored in HTTP-only secure cookies
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
- 👤 **User Profiles & Customization**:
  - Dynamic user profile routes (`/[username]`).
  - Full profile editing modal supporting custom display names, bios, avatars, and banner images.
  - Fallback-safe avatar rendering with first-letter fallback badges.
  - Interactive profile tabs: Posts, Replies, Highlights, Media, and Likes.
- 📱 **Interactive Dashboard & Timeline**:
  - Responsive 3-column Twitter/X layout (Left navigation, Main tweet timeline, Right widgets).
  - Tweet composer with character counter, media upload buttons, and post actions.
  - Trending topics and "Who to follow" sidebar widgets.
- 🗄️ **Dual Persistence Layer**:
  - Production-ready PostgreSQL storage with parameterized queries.
  - Zero-dependency in-memory repository fallback for fast testing and offline development.

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
| `GET` | `/api/users/:username` | Retrieve public user profile by username or `?username=` | No |
| `PUT` / `PATCH` | `/api/users/profile` | Update profile (name, bio, avatar) | Yes |

### Media & System
| Method | Endpoint | Description | Auth Required |
| :--- | :--- | :--- | :---: |
| `POST` | `/api/upload/avatar` | Upload an avatar image to Cloudinary (multipart form) | Yes |
| `GET` | `/api/health` | Health check probe endpoint | No |

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

> [!NOTE]
> If `DATABASE_URL` is omitted or PostgreSQL is unreachable, the backend gracefully falls back to the thread-safe `MemoryUserRepository` for local development.

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
- **PostgreSQL**: (Optional for local memory mode)

---

### 2. Backend Setup

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

### 3. Frontend Setup

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
