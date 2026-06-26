package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aniket/student-course-api/internal/apperr"
	"github.com/aniket/student-course-api/internal/model"
)

// fakeService is a configurable studentService for handler tests.
type fakeService struct {
	getFn  func(ctx context.Context, id int) (model.Student, error)
	listFn func(ctx context.Context) ([]model.Student, error)
}

func (f *fakeService) GetByID(ctx context.Context, id int) (model.Student, error) {
	return f.getFn(ctx, id)
}

func (f *fakeService) List(ctx context.Context) ([]model.Student, error) {
	return f.listFn(ctx)
}

// newTestServer builds a router wired to the given fake service.
func newTestServer(svc *fakeService) http.Handler {
	return NewRouter(NewStudentHandler(svc), nil)
}

func TestGetStudent(t *testing.T) {
	ada := model.Student{StudentID: 1, Name: "Ada", Address: "London", Grade: 10}

	tests := []struct {
		name       string
		path       string
		svc        *fakeService
		wantStatus int
		wantBody   string // substring expected in the body
	}{
		{
			name: "found returns 200 and student",
			path: "/students/1",
			svc: &fakeService{getFn: func(ctx context.Context, id int) (model.Student, error) {
				return ada, nil
			}},
			wantStatus: http.StatusOK,
			wantBody:   `"studentId":1`,
		},
		{
			name: "not found returns 404",
			path: "/students/99",
			svc: &fakeService{getFn: func(ctx context.Context, id int) (model.Student, error) {
				return model.Student{}, apperr.NotFound("student not found", nil)
			}},
			wantStatus: http.StatusNotFound,
			wantBody:   `"error":"student not found"`,
		},
		{
			name:       "non-integer id returns 400",
			path:       "/students/abc",
			svc:        &fakeService{},
			wantStatus: http.StatusBadRequest,
			wantBody:   `"error":"student id must be an integer"`,
		},
		{
			name: "internal error returns 500 with generic message",
			path: "/students/1",
			svc: &fakeService{getFn: func(ctx context.Context, id int) (model.Student, error) {
				return model.Student{}, context.DeadlineExceeded
			}},
			wantStatus: http.StatusInternalServerError,
			wantBody:   `"error":"internal server error"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := newTestServer(tc.svc)
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			srv.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d (body: %s)", rec.Code, tc.wantStatus, rec.Body.String())
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Fatalf("Content-Type = %q, want application/json", ct)
			}
			if body := rec.Body.String(); tc.wantBody != "" && !contains(body, tc.wantBody) {
				t.Fatalf("body = %s, want substring %q", body, tc.wantBody)
			}
		})
	}
}

func TestListStudents(t *testing.T) {
	t.Run("returns array", func(t *testing.T) {
		svc := &fakeService{listFn: func(ctx context.Context) ([]model.Student, error) {
			return []model.Student{{StudentID: 1, Name: "Ada"}}, nil
		}}
		rec := doGet(newTestServer(svc), "/students")
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		var got []model.Student
		if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
			t.Fatalf("unmarshal: %v (body %s)", err, rec.Body.String())
		}
		if len(got) != 1 {
			t.Fatalf("len = %d, want 1", len(got))
		}
	})

	t.Run("empty returns [] not null", func(t *testing.T) {
		svc := &fakeService{listFn: func(ctx context.Context) ([]model.Student, error) {
			return []model.Student{}, nil
		}}
		rec := doGet(newTestServer(svc), "/students")
		if got := rec.Body.String(); !contains(got, "[]") || contains(got, "null") {
			t.Fatalf("body = %s, want []", got)
		}
	})
}

func TestHealthz(t *testing.T) {
	t.Run("ok when check passes", func(t *testing.T) {
		router := NewRouter(NewStudentHandler(&fakeService{}), func(ctx context.Context) error { return nil })
		rec := doGet(router, "/healthz")
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
	})

	t.Run("503 when check fails", func(t *testing.T) {
		router := NewRouter(NewStudentHandler(&fakeService{}), func(ctx context.Context) error { return context.DeadlineExceeded })
		rec := doGet(router, "/healthz")
		if rec.Code != http.StatusServiceUnavailable {
			t.Fatalf("status = %d, want 503", rec.Code)
		}
	})
}

func doGet(h http.Handler, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
