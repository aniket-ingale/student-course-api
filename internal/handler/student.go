// Package handler contains the HTTP layer: request parsing, response writing,
// and mapping domain errors to status codes.
package handler

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/aniket/student-course-api/internal/apperr"
	"github.com/aniket/student-course-api/internal/model"
)

// studentService is the subset of the service layer the handler depends on.
// Defining it here keeps the handler decoupled and easy to fake in tests.
type studentService interface {
	GetByID(ctx context.Context, id int) (model.Student, error)
	List(ctx context.Context) ([]model.Student, error)
}

// StudentHandler serves the student endpoints.
type StudentHandler struct {
	svc studentService
}

// NewStudentHandler builds a StudentHandler over the given service.
func NewStudentHandler(svc studentService) *StudentHandler {
	return &StudentHandler{svc: svc}
}

// Get handles GET /students/{id}.
func (h *StudentHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "student id must be an integer")
		return
	}

	student, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, student)
}

// List handles GET /students.
func (h *StudentHandler) List(w http.ResponseWriter, r *http.Request) {
	students, err := h.svc.List(r.Context())
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, students)
}

// writeServiceError maps a service error to a status code and body, logging
// unexpected (5xx) failures with their full chain.
func (h *StudentHandler) writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
	status, msg := apperr.HTTPStatus(err)
	if status >= http.StatusInternalServerError {
		slog.ErrorContext(r.Context(), "request failed",
			"method", r.Method, "path", r.URL.Path, "error", err)
	}
	writeError(w, status, msg)
}
