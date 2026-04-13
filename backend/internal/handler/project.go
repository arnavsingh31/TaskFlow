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

type ProjectHandler struct {
	projectService *service.ProjectService
	logger         *zap.Logger
}

func NewProjectHandler(projectService *service.ProjectService, logger *zap.Logger) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
		logger:         logger,
	}
}

func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	// Parse and validate pagination params
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

	projects, total, err := h.projectService.List(r.Context(), userID, page, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respondJSON(w, http.StatusOK, &model.PaginatedResponse{
		Data:  projects,
		Total: total,
		Page:  page,
		Limit: limit,
	})
}

func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")

	project, err := h.projectService.GetByID(r.Context(), projectID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			respondError(w, http.StatusNotFound, "not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respondJSON(w, http.StatusOK, project)
}

func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var req model.CreateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if req.Name == "" {
		respondValidation(w, map[string]string{"name": "required"})
		return
	}

	// Reject HTML in input
	if helpers.ContainsHTML(req.Name) || helpers.ContainsHTMLPtr(req.Description) {
		respondError(w, http.StatusBadRequest, "input contains invalid characters")
		return
	}

	// Trim whitespace
	req.Name = helpers.TrimString(req.Name)
	req.Description = helpers.TrimPtr(req.Description)

	var keyPtr *string
	if key := r.Header.Get("X-Idempotency-Key"); key != "" {
		keyPtr = &key
	}

	project, err := h.projectService.Create(r.Context(), userID, keyPtr, &req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respondJSON(w, http.StatusCreated, project)
}

func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	projectID := r.PathValue("id")

	var req model.UpdateProjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// Reject HTML in input
	if helpers.ContainsHTMLPtr(req.Name) || helpers.ContainsHTMLPtr(req.Description) {
		respondError(w, http.StatusBadRequest, "input contains invalid characters")
		return
	}

	// Trim whitespace
	req.Name = helpers.TrimPtr(req.Name)
	req.Description = helpers.TrimPtr(req.Description)

	project, err := h.projectService.Update(r.Context(), userID, projectID, &req)
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

	respondJSON(w, http.StatusOK, project)
}

func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	projectID := r.PathValue("id")

	err := h.projectService.Delete(r.Context(), userID, projectID)
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

func (h *ProjectHandler) Stats(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")

	stats, err := h.projectService.Stats(r.Context(), projectID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			respondError(w, http.StatusNotFound, "not found")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respondJSON(w, http.StatusOK, stats)
}
