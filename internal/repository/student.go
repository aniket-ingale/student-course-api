// Package repository provides data access for domain models via GORM.
package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/aniket/student-course-api/internal/model"
)

// StudentRepository describes read access to student records.
type StudentRepository interface {
	GetByID(ctx context.Context, id int) (model.Student, error)
	List(ctx context.Context) ([]model.Student, error)
}

// studentRepo is the GORM-backed implementation of StudentRepository.
type studentRepo struct {
	db *gorm.DB
}

// NewStudentRepository builds a StudentRepository backed by the given GORM DB.
func NewStudentRepository(db *gorm.DB) StudentRepository {
	return &studentRepo{db: db}
}

// GetByID fetches a single student. It returns ErrNotFound when no row matches.
func (r *studentRepo) GetByID(ctx context.Context, id int) (model.Student, error) {
	var s model.Student
	err := r.db.WithContext(ctx).Where("student_id = ?", id).First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, fmt.Errorf("repository: get student %d: %w", id, err)
	}
	return s, nil
}

// List returns all students ordered by student_id. The slice is never nil.
func (r *studentRepo) List(ctx context.Context) ([]model.Student, error) {
	students := make([]model.Student, 0)
	if err := r.db.WithContext(ctx).Order("student_id").Find(&students).Error; err != nil {
		return nil, fmt.Errorf("repository: list students: %w", err)
	}
	return students, nil
}
