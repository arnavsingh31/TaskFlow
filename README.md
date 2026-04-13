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

### Why Standard Library Over Gin/Chi

Go 1.22 introduced method-based routing and path parameters in `net/http.ServeMux` — features that previously required a third-party router. Since this project has ~12 routes and straightforward middleware needs, the standard library covers everything without adding framework overhead. The tradeoff is slightly more manual middleware wiring, but the result is zero routing dependencies and code that uses Go's native HTTP types (`http.Request`, `http.ResponseWriter`) throughout.

### Backend Structure

The backend follows a layered architecture where each layer has a single responsibility:

```
HTTP Request → Middleware → Handler → Service → Repository → PostgreSQL
```

- **Handlers** own the HTTP contract — parse requests, validate input, map errors to status codes, write responses. They never touch SQL or make authorization decisions.
- **Services** own business logic — authorization checks (is this user the project owner?), transaction management, and orchestrating multiple repository calls within a single transaction.
- **Repositories** own data access — raw SQL queries with parameterized inputs. They return Go structs and don't know about HTTP or business rules.

Dependencies flow one direction (handler → service → repository) and are injected via constructors in `main.go`. No global state, no init functions, no service locator patterns.

This separation was chosen over a simpler flat structure because the project has enough complexity (soft deletes, row locking, idempotency, authorization) that mixing these concerns in handlers would make the code harder to reason about and test.

### Data Integrity

**Transactions:** Every write operation (create, update, delete) runs inside a `WithTx()` wrapper that handles commit/rollback automatically. This ensures multi-step operations (e.g., soft-deleting a project and all its tasks) either fully succeed or fully roll back.

**Row Locking:** Update and delete operations use `SELECT ... FOR UPDATE` to lock the target row before modifying it. This prevents race conditions where two concurrent requests could read stale data and overwrite each other's changes.

**RETURNING clause:** All INSERT and UPDATE queries use `RETURNING *` to get the actual database state back, rather than trusting what was sent. This avoids discrepancies between application state and DB state (e.g., default values, triggers).

### Soft Deletes

All tables use a `deleted_at TIMESTAMP` column instead of hard deletes. DELETE endpoints set `deleted_at = NOW()` and all SELECT queries include `WHERE deleted_at IS NULL`. Partial indexes (e.g., `CREATE INDEX ... WHERE deleted_at IS NULL`) ensure only active rows are scanned, so deleted rows don't degrade query performance.

The tradeoff is slightly more complex queries (every SELECT needs the filter), but the benefit is data recovery capability and a complete audit trail of when records were removed.

### Idempotency

POST endpoints for creating projects and tasks accept an `X-Idempotency-Key` header. The key is stored in a table with a composite primary key `(key, user_id)` — scoped per user so two different users can safely use the same key independently. On duplicate requests, the backend returns the originally created resource instead of creating a duplicate.

This was added because double-form-submissions and network retries are common in real usage. The frontend also disables submit buttons during requests as a first layer of defense. In a production system, idempotency keys would ideally live in Redis for faster lookups and built-in TTL expiry, but for the current scope that would add unnecessary infrastructure complexity — a PostgreSQL table with periodic cleanup serves the same purpose.

### Frontend

Next.js was considered but would have added SSR complexity that isn't needed for a dashboard app behind authentication — there are no public pages that benefit from server-side rendering or SEO. Instead, React with TypeScript (Vite) was chosen for a simpler, faster build with client-side rendering. **Shadcn/ui** is used as the component library.

### What Was Intentionally Left Out

- **ORM (GORM, sqlx):** Raw SQL with `database/sql` was chosen for full control over queries, especially for dynamic PATCH updates and complex JOINs. The tradeoff is more boilerplate for scanning rows, but every query is explicit and easy to reason about.

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
- **Project membership** system with invite/join flow — currently users see projects they own or have tasks in, but a formal member role with permissions would be more robust
- **Email notifications** when a task is assigned to you or a due date is approaching
- **User profile page** — view and update name, email, and profile picture
- **Account/password recovery** — forgot password flow with email-based reset link and token verification
