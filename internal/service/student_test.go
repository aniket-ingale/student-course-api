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
	getFn    func(ctx context.Context, id int) (model.Student, error)
	listFn   func(ctx context.Context) ([]model.Student, error)
	createFn func(ctx context.Context, s model.Student) (model.Student, error)
	updateFn func(ctx context.Context, s model.Student) (model.Student, error)
	deleteFn func(ctx context.Context, id int) error
}

func (f *fakeRepo) GetByID(ctx context.Context, id int) (model.Student, error) {
	return f.getFn(ctx, id)
}

func (f *fakeRepo) List(ctx context.Context) ([]model.Student, error) {
	return f.listFn(ctx)
}

func (f *fakeRepo) Create(ctx context.Context, s model.Student) (model.Student, error) {
	return f.createFn(ctx, s)
}

func (f *fakeRepo) Update(ctx context.Context, s model.Student) (model.Student, error) {
	return f.updateFn(ctx, s)
}

func (f *fakeRepo) Delete(ctx context.Context, id int) error {
	return f.deleteFn(ctx, id)
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

func TestValidateStudentWrite(t *testing.T) {
	valid := model.Student{Name: "Ada", Address: "London", Grade: 10}

	tests := []struct {
		name       string
		in         model.Student
		wantErr    bool
		wantStatus int // expected HTTP status via apperr.HTTPStatus
	}{
		{name: "valid", in: valid},
		{name: "high grade has no upper bound", in: model.Student{Name: "Ada", Address: "London", Grade: 99}},
		{name: "empty name", in: model.Student{Name: "", Address: "London", Grade: 10}, wantErr: true, wantStatus: http.StatusBadRequest},
		{name: "whitespace name", in: model.Student{Name: "   ", Address: "London", Grade: 10}, wantErr: true, wantStatus: http.StatusBadRequest},
		{name: "empty address", in: model.Student{Name: "Ada", Address: "", Grade: 10}, wantErr: true, wantStatus: http.StatusBadRequest},
		{name: "whitespace address", in: model.Student{Name: "Ada", Address: "  ", Grade: 10}, wantErr: true, wantStatus: http.StatusBadRequest},
		{name: "zero grade", in: model.Student{Name: "Ada", Address: "London", Grade: 0}, wantErr: true, wantStatus: http.StatusBadRequest},
		{name: "negative grade", in: model.Student{Name: "Ada", Address: "London", Grade: -3}, wantErr: true, wantStatus: http.StatusBadRequest},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateStudentWrite(tc.in)
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
