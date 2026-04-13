package main

import (
	"fmt"
	"net/http"

	"github.com/taskflow/backend/internal/handler"
	"github.com/taskflow/backend/internal/middleware"
)

func newRouter(
	jwtSecret string,
	authHandler *handler.AuthHandler,
	userHandler *handler.UserHandler,
	projectHandler *handler.ProjectHandler,
	taskHandler *handler.TaskHandler,
) http.Handler {
	authMW := func(next http.HandlerFunc) http.HandlerFunc {
		return middleware.Auth(jwtSecret, next)
	}

	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)

	// Protected routes
	mux.HandleFunc("GET /projects", authMW(projectHandler.List))
	mux.HandleFunc("POST /projects", authMW(projectHandler.Create))
	mux.HandleFunc("GET /projects/{id}", authMW(projectHandler.Get))
	mux.HandleFunc("PATCH /projects/{id}", authMW(projectHandler.Update))
	mux.HandleFunc("DELETE /projects/{id}", authMW(projectHandler.Delete))
	mux.HandleFunc("GET /projects/{id}/stats", authMW(projectHandler.Stats))

	mux.HandleFunc("GET /users/search", authMW(userHandler.Search))

	mux.HandleFunc("GET /projects/{id}/tasks", authMW(taskHandler.List))
	mux.HandleFunc("POST /projects/{id}/tasks", authMW(taskHandler.Create))
	mux.HandleFunc("PATCH /tasks/{id}", authMW(taskHandler.Update))
	mux.HandleFunc("DELETE /tasks/{id}", authMW(taskHandler.Delete))

	// Catch-all 404
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"error":"not found"}`)
	})

	return middleware.CORS(mux)
}
