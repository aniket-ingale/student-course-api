// Package service holds business logic, sitting between handlers and the
// repository layer.
package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/aniket/student-course-api/internal/apperr"
	"github.com/aniket/student-course-api/internal/model"
	"github.com/aniket/student-course-api/internal/repository"
)

// StudentService exposes student operations to the HTTP layer.
type StudentService struct {
	repo repository.StudentRepository
}

// NewStudentService constructs a StudentService over the given repository.
func NewStudentService(repo repository.StudentRepository) *StudentService {
	return &StudentService{repo: repo}
}

// GetByID validates the id, fetches the student, and translates a missing row
// into a not-found domain error.
func (s *StudentService) GetByID(ctx context.Context, id int) (model.Student, error) {
	if id <= 0 {
		return model.Student{}, apperr.Validation("student id must be a positive integer")
	}

	student, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return model.Student{}, apperr.NotFound("student not found", err)
		}
		return model.Student{}, fmt.Errorf("service: get student %d: %w", id, err)
	}
	return student, nil
}

// List returns all students.
func (s *StudentService) List(ctx context.Context) ([]model.Student, error) {
	students, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("service: list students: %w", err)
	}
	return students, nil
}
