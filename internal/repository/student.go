// Package repository provides data access for domain models via GORM.
package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/aniket/student-course-api/internal/model"
)

// StudentRepository describes read and write access to student records.
type StudentRepository interface {
	GetByID(ctx context.Context, id int) (model.Student, error)
	List(ctx context.Context) ([]model.Student, error)
	Create(ctx context.Context, s model.Student) (model.Student, error)
	Update(ctx context.Context, s model.Student) (model.Student, error)
	Delete(ctx context.Context, id int) error
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

// Create inserts a new student. student_id is DB-assigned (IDENTITY column), so
// GORM omits it on insert and populates s.StudentID from the RETURNING clause.
func (r *studentRepo) Create(ctx context.Context, s model.Student) (model.Student, error) {
	// Defend against a caller-supplied PK: the IDENTITY column rejects it and
	// GORM should not attempt to write it.
	s.StudentID = 0
	if err := r.db.WithContext(ctx).Create(&s).Error; err != nil {
		return model.Student{}, fmt.Errorf("repository: create student: %w", err)
	}
	return s, nil
}

// Update overwrites the mutable columns (name, address, grade) of the student
// identified by s.StudentID. Select is used so a zero-value grade is still
// written. A zero RowsAffected means no such row, surfaced as ErrNotFound.
func (r *studentRepo) Update(ctx context.Context, s model.Student) (model.Student, error) {
	res := r.db.WithContext(ctx).
		Model(&model.Student{}).
		Where("student_id = ?", s.StudentID).
		Select("name", "address", "grade").
		Updates(s)
	if res.Error != nil {
		return model.Student{}, fmt.Errorf("repository: update student %d: %w", s.StudentID, res.Error)
	}
	if res.RowsAffected == 0 {
		return model.Student{}, ErrNotFound
	}
	return s, nil
}

// Delete removes the student with the given id. A zero RowsAffected means no
// such row, surfaced as ErrNotFound so the service maps it to a 404.
func (r *studentRepo) Delete(ctx context.Context, id int) error {
	res := r.db.WithContext(ctx).
		Where("student_id = ?", id).
		Delete(&model.Student{})
	if res.Error != nil {
		return fmt.Errorf("repository: delete student %d: %w", id, res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
