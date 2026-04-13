package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/taskflow/backend/internal/helpers"
	"github.com/taskflow/backend/internal/model"
	"github.com/taskflow/backend/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
	logger      *zap.Logger
}

func NewAuthHandler(authService *service.AuthService, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	fields := make(map[string]string)
	if req.Name == nil || *req.Name == "" {
		fields["name"] = "required"
	}
	if !helpers.ValidEmail(req.Email) {
		fields["email"] = "valid email required"
	}
	if req.Password == nil || len(*req.Password) < 8 {
		fields["password"] = "minimum 8 characters required"
	}
	if len(fields) > 0 {
		respondValidation(w, fields)
		return
	}

	// Reject HTML in input
	if helpers.ContainsHTMLPtr(req.Name) {
		respondError(w, http.StatusBadRequest, "input contains invalid characters")
		return
	}

	// Trim whitespace
	req.Name = helpers.TrimPtr(req.Name)

	user, token, err := h.authService.Register(r.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrConflict) {
			respondError(w, http.StatusConflict, "email already exists")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respondJSON(w, http.StatusCreated, &model.AuthResponse{Token: token, User: user})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	fields := make(map[string]string)
	if req.Email == nil || *req.Email == "" {
		fields["email"] = "required"
	}
	if req.Password == nil || *req.Password == "" {
		fields["password"] = "required"
	}
	if len(fields) > 0 {
		respondValidation(w, fields)
		return
	}

	user, token, err := h.authService.Login(r.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrUnauthorized) {
			respondError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}
		respondError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	respondJSON(w, http.StatusOK, &model.AuthResponse{Token: token, User: user})
}
