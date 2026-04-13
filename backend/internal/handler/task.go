package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"go.uber.org/zap"

	"github.com/taskflow/backend/internal/helpers"
	"github.com/taskflow/backend/internal/middleware"
	"github.com/taskflow/backend/internal/model"
	"github.com/taskflow/backend/internal/service"
)

type TaskHandler struct {
	taskService *service.TaskService
	logger      *zap.Logger
}

func NewTaskHandler(taskService *service.TaskService, logger *zap.Logger) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
		logger:      logger,
	}
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")
	status := r.URL.Query().Get("status")
	assignee := r.URL.Query().Get("assignee")

	var statusPtr, assigneePtr *string
	if status != "" {
		if !helpers.ValidStatus(&status) {
			respondError(w, http.StatusBadRequest, "invalid status filter")
			return
		}
		statusPtr = &status
	}
	if assignee != "" {
		assigneePtr = &assignee
	}

	page := 1
	limit := 10

	if p := r.URL.Query().Get("page"); p != "" {
		parsed, err := strconv.Atoi(p)
		if err != nil || parsed < 1 {
			respondError(w, http.StatusBadRequest, "page must be a positive integer")
			return
		}
		page = parsed
	}

	if l := r.URL.Query().Get("limit"); l != "" {
		parsed, err := strconv.Atoi(l)
		if err != nil || parsed < 1 || parsed > 100 {
			respondError(w, http.StatusBadRequest, "limit must be between 1 and 100")
			return
		}
		limit = parsed
	}

	tasks, total, err := h.taskService.ListByProject(r.Context(), projectID, statusPtr, assigneePtr, page, limit)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			respondError(w, http.StatusNotFound, "not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respondJSON(w, http.StatusOK, &model.PaginatedResponse{
		Data:  tasks,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	projectID := r.PathValue("id")

	var req model.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	fields := make(map[string]string)
	if req.Title == "" {
		fields["title"] = "required"
	}
	if !helpers.ValidStatus(req.Status) {
		fields["status"] = "must be todo, in_progress, or done"
	}
	if !helpers.ValidPriority(req.Priority) {
		fields["priority"] = "must be low, medium, or high"
	}
	if len(fields) > 0 {
		respondValidation(w, fields)
		return
	}

	var keyPtr *string
	if key := r.Header.Get("X-Idempotency-Key"); key != "" {
		keyPtr = &key
	}

	task, err := h.taskService.Create(r.Context(), userID, projectID, keyPtr, &req)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			respondError(w, http.StatusNotFound, "not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respondJSON(w, http.StatusCreated, task)
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	taskID := r.PathValue("id")

	var req model.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	fields := make(map[string]string)
	if !helpers.ValidStatus(req.Status) {
		fields["status"] = "must be todo, in_progress, or done"
	}
	if !helpers.ValidPriority(req.Priority) {
		fields["priority"] = "must be low, medium, or high"
	}
	if len(fields) > 0 {
		respondValidation(w, fields)
		return
	}

	task, err := h.taskService.Update(r.Context(), userID, taskID, &req)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			respondError(w, http.StatusNotFound, "not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respondJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	taskID := r.PathValue("id")

	err := h.taskService.Delete(r.Context(), userID, taskID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			respondError(w, http.StatusNotFound, "not found")
			return
		}
		if errors.Is(err, service.ErrForbidden) {
			respondError(w, http.StatusForbidden, "forbidden")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
