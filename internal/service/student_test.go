package service

import (
	"context"
	"errors"
	"testing"

	"net/http"

	"github.com/aniket/student-course-api/internal/apperr"
	"github.com/aniket/student-course-api/internal/model"
	"github.com/aniket/student-course-api/internal/repository"
)

// fakeRepo is a configurable StudentRepository for unit tests.
type fakeRepo struct {
	getFn  func(ctx context.Context, id int) (model.Student, error)
	listFn func(ctx context.Context) ([]model.Student, error)
}

func (f *fakeRepo) GetByID(ctx context.Context, id int) (model.Student, error) {
	return f.getFn(ctx, id)
}

func (f *fakeRepo) List(ctx context.Context) ([]model.Student, error) {
	return f.listFn(ctx)
}

func TestGetByID(t *testing.T) {
	want := model.Student{StudentID: 1, Name: "Ada", Address: "London", Grade: 10}

	tests := []struct {
		name       string
		id         int
		repoFn     func(ctx context.Context, id int) (model.Student, error)
		wantErr    bool
		wantStatus int // expected HTTP status via apperr.HTTPStatus
		wantStud   model.Student
	}{
		{
			name:     "found",
			id:       1,
			repoFn:   func(ctx context.Context, id int) (model.Student, error) { return want, nil },
			wantStud: want,
		},
		{
			name: "zero id is validation error",
			id:   0,
			repoFn: func(ctx context.Context, id int) (model.Student, error) {
				t.Fatal("repo should not be called")
				return model.Student{}, nil
			},
			wantErr:    true,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "negative id is validation error",
			id:   -5,
			repoFn: func(ctx context.Context, id int) (model.Student, error) {
				t.Fatal("repo should not be called")
				return model.Student{}, nil
			},
			wantErr:    true,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "not found maps to NotFound",
			id:   99,
			repoFn: func(ctx context.Context, id int) (model.Student, error) {
				return model.Student{}, repository.ErrNotFound
			},
			wantErr:    true,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "unexpected repo error propagates as internal",
			id:         1,
			repoFn:     func(ctx context.Context, id int) (model.Student, error) { return model.Student{}, errors.New("boom") },
			wantErr:    true,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewStudentService(&fakeRepo{getFn: tc.repoFn})
			got, err := svc.GetByID(context.Background(), tc.id)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if status, _ := apperr.HTTPStatus(err); status != tc.wantStatus {
					t.Fatalf("status = %d, want %d (err: %v)", status, tc.wantStatus, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.wantStud {
				t.Fatalf("student = %+v, want %+v", got, tc.wantStud)
			}
		})
	}
}

func TestList(t *testing.T) {
	t.Run("returns students", func(t *testing.T) {
		want := []model.Student{{StudentID: 1, Name: "Ada"}, {StudentID: 2, Name: "Alan"}}
		svc := NewStudentService(&fakeRepo{listFn: func(ctx context.Context) ([]model.Student, error) {
			return want, nil
		}})
		got, err := svc.List(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != len(want) {
			t.Fatalf("len = %d, want %d", len(got), len(want))
		}
	})

	t.Run("propagates repo error", func(t *testing.T) {
		svc := NewStudentService(&fakeRepo{listFn: func(ctx context.Context) ([]model.Student, error) {
			return nil, errors.New("boom")
		}})
		if _, err := svc.List(context.Background()); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
