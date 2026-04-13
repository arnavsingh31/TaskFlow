# TaskFlow

A full-stack task management system where users register, log in, create projects, and manage tasks within projects.

## Tech Stack

- **Backend:** Go 1.22 (standard library `net/http` routing)
- **Database:** PostgreSQL 16
- **Frontend:** React 18 + TypeScript + Tailwind CSS + Shadcn/ui
- **Infrastructure:** Docker + docker-compose
- **Logging:** Uber Zap (structured JSON logging)
- **Auth:** JWT (HMAC-SHA256) + bcrypt (cost 12)
- **Migrations:** golang-migrate

## Architecture Decisions

### Standard Library Routing

Used Go 1.22+'s enhanced `net/http.ServeMux` with method-based routing (`GET /projects/{id}`) instead of a framework like Gin or Chi. This keeps external dependencies minimal and uses Go's native HTTP primitives directly.

### Component Library

Frontend uses **Shadcn/ui** (built on Tailwind CSS + Radix/Base UI primitives). Components are copied into the project — no heavy runtime dependency. Provides accessible, customizable UI with minimal bundle overhead.

### Layered Backend Architecture

```
HTTP Request → Middleware → Handler → Service → Repository → PostgreSQL
```

- **Handlers** parse requests, validate input, and return responses. They never touch SQL.
- **Services** contain business logic, authorization checks, and manage transactions.
- **Repositories** execute SQL queries. Dependencies are injected via struct constructors. No global variables.

### Transaction & Concurrency Safety

- All write operations use `WithTx()` wrapper for automatic commit/rollback
- `SELECT ... FOR UPDATE` locks rows before updates/deletes to prevent race conditions
- `RETURNING *` on all INSERT/UPDATE ensures the app sees the actual DB state

### Soft Deletes

All tables have a `deleted_at` timestamp. DELETE endpoints set `deleted_at = NOW()` instead of removing rows. Partial indexes (`WHERE deleted_at IS NULL`) ensure only active rows are scanned.

### Idempotency

POST endpoints accept an `X-Idempotency-Key` header. Keys are scoped per user via a composite primary key `(key, user_id)` to prevent duplicate resource creation from double-clicks or retries.

## Running Locally

```bash
git clone <repo-url>
cd taskflow
cp .env.example .env
docker compose up --build
```

Open `http://localhost:3000` in your browser.

- Backend API: `http://localhost:8080`
- Frontend: `http://localhost:3000`
- PostgreSQL: `localhost:5432`

## Running Migrations

Migrations run **automatically** when the backend container starts. No manual steps needed.

Migration files are in `backend/migrations/` using golang-migrate's numbered format with up/down pairs.

## Test Credentials

```
Email:    test@example.com
Password: password123
```

A seed migration creates this user along with 1 sample project and 3 tasks (todo, in_progress, done).

## API Reference

### Authentication

```bash
# Register
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name": "John", "email": "john@example.com", "password": "password123"}'
# Returns: {"token": "...", "user": {...}}

# Login
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "john@example.com", "password": "password123"}'
# Returns: {"token": "...", "user": {...}}
```

### Projects (all require `Authorization: Bearer <token>`)

```bash
# List projects (owned or assigned tasks in)
GET /projects

# Create project
POST /projects
Body: {"name": "My Project", "description": "Optional desc"}
Header: X-Idempotency-Key: <uuid>

# Get project detail (includes tasks)
GET /projects/:id

# Update project (owner only)
PATCH /projects/:id
Body: {"name": "New Name"}

# Delete project (owner only, soft delete)
DELETE /projects/:id
```

### Tasks (all require `Authorization: Bearer <token>`)

```bash
# List tasks with optional filters
GET /projects/:id/tasks?status=todo&assignee=<user-uuid>

# Create task
POST /projects/:id/tasks
Body: {"title": "Task", "priority": "high", "assignee_id": "<uuid>"}
Header: X-Idempotency-Key: <uuid>

# Update task
PATCH /tasks/:id
Body: {"status": "done"}

# Delete task (project owner or task creator only, soft delete)
DELETE /tasks/:id
```

### User Search

```bash
# Search user by exact email (for task assignment)
GET /users/search?email=john@example.com
# Returns: {"id": "...", "name": "...", "email": "..."}
```

### Error Responses

| Status | Body |
|--------|------|
| 400 | `{"error": "validation failed", "fields": {"name": "required"}}` |
| 401 | `{"error": "unauthorized"}` |
| 403 | `{"error": "forbidden"}` |
| 404 | `{"error": "not found"}` |
| 409 | `{"error": "email already exists"}` |
| 500 | `{"error": "internal server error"}` |

## What You'd Do With More Time

- **Real-time updates** via WebSocket or SSE so multiple users see changes instantly
- **Rate limiting** on auth endpoints to prevent brute-force attacks
- **Refresh tokens** with short-lived access tokens instead of single 24h JWTs
- **Project membership** system with invite/join flow — currently any authenticated user can view any project
- **Email notifications** when a task is assigned to you or a due date is approaching
