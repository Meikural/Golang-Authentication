## 🦫 [Collaboration] Go Auth + RBAC + Policy Microservice — Fiber + JWT + PostgreSQL

Hey everyone 👋

I’m building an **open-source authentication and RBAC microservice** written in **Go (Fiber)** with **PostgreSQL**, designed to be **self-hosted, reusable, and Docker-ready**.

The goal is a **plug-and-play Auth module** for any internal or distributed system —
JWT-only, no external auth providers, no ORM — just pure Go + SQL + Docker.

---

### 💡 Core Idea

A standalone **Auth + RBAC + Policy service** that runs as its own container and exposes REST APIs under `/api/v1`.
Applications simply talk to it via HTTP or Docker network.

**Core features:**

- 🔐 JWT-based Authentication (Access + Refresh Tokens)
- 🧩 Role-Based Access Control (Super Admin → Admin → User / Service)
- ⚙️ Policy Engine (control registration, verification, etc. dynamically)
- 🧱 Super Admin auto-seeding from environment
- 🗄️ PostgreSQL with plain SQL (no ORM)

**Perfect for:** internal tools, self-hosted dashboards, microservice backends.

---

### 🏗️ Architecture

```
+---------------------------------------------+
|           AUTH SERVICE (Go + Fiber)         |
|---------------------------------------------|
| JWT Access + Refresh Tokens                 |
| RBAC System (Super Admin → Admin → User)    |
| Policy Manager (runtime config)             |
| PostgreSQL + Plain SQL (pgx / database/sql) |
+---------------------------------------------+
             ↑
             |
             ↓
+---------------------------------------------+
|  Any Application (Go, Node, Django, etc.)   |
|  - Uses REST APIs via Docker Compose        |
|  - Verifies JWT from Auth Service           |
+---------------------------------------------+
```

---

### 🔐 Key Endpoints

**Base URL:** `/api/v1`

| Category     | Method | Path                      | Access            | Description                |
| ------------ | ------ | ------------------------- | ----------------- | -------------------------- |
| **Auth**     | POST   | `/register`               | depends_on_policy | Register new user          |
|              | POST   | `/login`                  | public            | Authenticate and issue JWT |
|              | POST   | `/refresh`                | public            | Refresh token              |
|              | GET    | `/me`                     | authenticated     | Get current user           |
|              | POST   | `/logout`                 | authenticated     | Invalidate refresh token   |
| **Users**    | GET    | `/admin/users`            | admin/super_admin | List all users             |
|              | POST   | `/admin/users`            | depends_on_policy | Create user manually       |
|              | PATCH  | `/admin/users/:id/status` | admin/super_admin | Activate/deactivate        |
|              | DELETE | `/admin/users/:id`        | super_admin       | Delete user                |
| **Roles**    | GET    | `/admin/roles`            | super_admin       | List roles                 |
|              | POST   | `/admin/roles`            | super_admin       | Create role                |
|              | POST   | `/admin/assign-role`      | super_admin       | Assign roles               |
| **Policies** | GET    | `/superadmin/policies`    | super_admin       | List or update policies    |
| **System**   | GET    | `/health`                 | public            | Health check               |
|              | GET    | `/version`                | public            | Version info               |

---

### ⚙️ Example `.env`

```env
DATABASE_URL=postgres://user:pass@postgres:5432/authdb?sslmode=disable
JWT_SECRET=supersecretkey
SUPERADMIN_EMAIL=superadmin@internal.local
SUPERADMIN_PASSWORD=change_me_now
```

---

### 🧭 Highlights

- Super Admin seeded automatically on startup
- Policy table controls runtime behavior (registration, email verification, etc.)
- Roles: Super Admin, Admin, User, Service
- All tokens are JWTs — easily verifiable by other services
- Can be run via:

  ```bash
  docker-compose up -d auth
  ```

---

### 🧰 Tech Stack

- **Language:** Go
- **Framework:** Fiber
- **Database:** PostgreSQL
- **Auth:** JWT (Access + Refresh)
- **Queries:** Plain SQL (no ORM)
- **Deployment:** Docker + Docker Compose

---

### 🚀 Vision

To create a **reusable, self-hosted Auth microservice** that you can spin up in seconds.
Think of it as a minimal, internal **Auth0-style** service — written in Go, open-source, and easy to extend.

Future ideas:

- CLI (`authctl`) for superadmin ops
- Audit logs + caching layer
- Optional OAuth or MFA modules

---

### 🤝 Looking For Collaborators

I’m looking for Go developers interested in:

- Microservice architecture
- JWT auth, Fiber, and PostgreSQL
- Policy-based config systems
- Security + Docker setup reviews

Repo: [github.com/Meikural/Golang-Authentication](https://github.com/Meikural/Golang-Authentication)

If you love clean Go code, minimal dependencies, and well-defined architecture —
let’s collaborate 🚀
