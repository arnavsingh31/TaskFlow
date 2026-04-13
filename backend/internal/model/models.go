package model

import "time"

// --- Domain Models ---

type User struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	Password  string     `json:"-"`
	DeletedAt *time.Time `json:"-"`
	CreatedAt time.Time  `json:"created_at"`
}

type Project struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description *string    `json:"description,omitempty"`
	OwnerID     string     `json:"owner_id"`
	DeletedAt   *time.Time `json:"-"`
	CreatedAt   time.Time  `json:"created_at"`
}

type Task struct {
	ID            string     `json:"id"`
	Title         string     `json:"title"`
	Description   *string    `json:"description,omitempty"`
	Status        string     `json:"status"`
	Priority      string     `json:"priority"`
	ProjectID     string     `json:"project_id"`
	AssigneeID    *string    `json:"assignee_id,omitempty"`
	AssigneeName  *string    `json:"assignee_name,omitempty"`
	AssigneeEmail *string    `json:"assignee_email,omitempty"`
	CreatedBy     string     `json:"created_by"`
	DueDate       *string    `json:"due_date,omitempty"`
	DeletedAt     *time.Time `json:"-"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// --- Request Types ---
// Pointers here because PATCH needs to distinguish "not sent" (nil) from "sent empty" ("")

type RegisterRequest struct {
	Name     *string `json:"name"`
	Email    *string `json:"email"`
	Password *string `json:"password"`
}

type LoginRequest struct {
	Email    *string `json:"email"`
	Password *string `json:"password"`
}

type CreateProjectRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type UpdateProjectRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type CreateTaskRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	Priority    *string `json:"priority"`
	AssigneeID  *string `json:"assignee_id"`
	DueDate     *string `json:"due_date"`
}

type UpdateTaskRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	Priority    *string `json:"priority"`
	AssigneeID  *string `json:"assignee_id"`
	DueDate     *string `json:"due_date"`
}

// --- Response Types ---

type AuthResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

type ProjectDetailResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	OwnerID     string    `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	Tasks       []*Task   `json:"tasks"`
}

type ListResponse struct {
	Data any `json:"data"`
}

type PaginatedResponse struct {
	Data  any `json:"data"`
	Total int `json:"total"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

// Raw types returned by repository (no business/display logic)

type StatusCount struct {
	Status string
	Count  int
}

type AssigneeCount struct {
	AssigneeID *string
	Name       *string
	Count      int
}

// API response types (shaped by service layer)

type AssigneeStat struct {
	AssigneeID *string `json:"assignee_id"`
	Name       string  `json:"assignee_name"`
	Count      int     `json:"count"`
}

type ProjectStatsResponse struct {
	Todo       int             `json:"todo"`
	InProgress int             `json:"in_progress"`
	Done       int             `json:"done"`
	ByAssignee []*AssigneeStat `json:"by_assignee"`
}
