# XClone

A modern full-stack social media application (Twitter / X clone) built with a Next.js frontend and a Go (Golang) backend following clean, layered architectural principles.

---

## 📁 Project Architecture & Directory Structure

The repository is organized into distinct `frontend` and `backend` workspaces:

```
XClone/
├── frontend/                     # Next.js 16 Client Application
│   ├── app/                      # Next.js App Router (pages, layouts, globals)
│   ├── public/                   # Static assets (SVGs, icons)
│   ├── .gitignore                # Frontend-specific ignore rules
│   ├── eslint.config.mjs         # ESLint configuration
│   ├── next.config.ts            # Next.js configuration
│   ├── package.json              # Frontend npm dependencies & scripts
│   ├── postcss.config.mjs        # PostCSS configuration
│   └── tsconfig.json             # TypeScript compiler configuration
│
├── backend/                      # Go (Golang) API Server
│   ├── config/                   # Configuration loader & environment variables
│   │   └── config.go
│   ├── controllers/              # HTTP request handlers & API response formatters
│   │   └── controller.go
│   ├── middleware/               # HTTP middlewares (CORS, request logging, auth)
│   │   └── middleware.go
│   ├── models/                   # Core domain data models (User, Post)
│   │   └── model.go
│   ├── repository/               # Data access layer & database interfaces
│   │   └── repository.go
│   ├── routes/                   # HTTP routing setup & route registration
│   │   └── routes.go
│   ├── service/                  # Business logic layer
│   │   └── service.go
│   ├── utils/                    # Helper utilities (token generation, hashing)
│   │   └── utils.go
│   ├── go.mod                    # Go module definition
│   └── main.go                   # Backend server entrypoint
│
├── .gitignore                    # Root git ignore rules
└── README.md                     # Project documentation
```

---

## 🛠 Tech Stack

### Frontend
- **Framework**: [Next.js 16](https://nextjs.org/) (App Router, Turbopack)
- **UI Library**: [React 19](https://react.dev/)
- **Styling**: [Tailwind CSS v4](https://tailwindcss.com/)
- **Language**: [TypeScript](https://www.typescriptlang.org/)
- **Linting**: [ESLint](https://eslint.org/)

### Backend
- **Language**: [Go (Golang 1.23+)](https://go.dev/)
- **Architecture**: Layered / Clean Architecture
  - **`config`**: Environment parsing with fallback defaults
  - **`controllers`**: Parses HTTP inputs and writes JSON responses
  - **`middleware`**: Reusable HTTP handlers (CORS, Logger, JWT Auth)
  - **`models`**: Data structures defining the domain (Users, Tweets/Posts)
  - **`repository`**: Database interaction interfaces
  - **`routes`**: Routing configuration mapping URLs to controllers
  - **`service`**: Core application logic and validation
  - **`utils`**: Cryptographic tokens and utility helpers
  - **`main.go`**: Service initialization, config loading, and server listening

---

## 🚀 Getting Started

### Prerequisites
- **Node.js**: v18.0.0 or higher
- **npm**: v9.0.0 or higher
- **Go**: v1.23.0 or higher

---

### 1. Frontend Setup

Navigate to the `frontend` folder, install dependencies, and start the development server:

```bash
cd frontend
npm install
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) in your browser to view the application.

#### Frontend Scripts
- `npm run dev` - Start the development server with Turbopack
- `npm run build` - Create an optimized production build
- `npm run start` - Start the production server
- `npm run lint` - Run ESLint checks

---

### 2. Backend Setup

Navigate to the `backend` folder and start the API server:

```bash
cd backend
go run main.go
```

The server will start on [http://localhost:8080](http://localhost:8080).

#### Test the Backend Health Check
```bash
curl http://localhost:8080/api/health
```

Expected response:
```json
{"status":"ok","service":"xclone-backend"}
```

---

## ⚙️ Environment Configuration

### Backend (`backend/.env`)
| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | Port for the Go backend server |
| `ENV` | `development` | Runtime environment (`development` / `production`) |
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/xclone?sslmode=disable` | Database connection string |
| `JWT_SECRET` | `xclone-super-secret-key` | Secret key used for signing JWT tokens |

### Frontend (`frontend/.env.local`)
| Variable | Default | Description |
|---|---|---|
| `NEXT_PUBLIC_API_URL` | `http://localhost:8080` | Base URL pointing to the Go backend API |

---

## 📄 License
This project is open source and available under the [MIT License](LICENSE).
