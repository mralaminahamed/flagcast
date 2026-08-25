package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestCreateValidation(t *testing.T) {
	// Store is never reached on these validation-failure paths, so nil is fine.
	h := New(nil, nil, "")
	tests := []struct {
		name, body string
		want       int
	}{
		{"invalid json", `{`, http.StatusBadRequest},
		{"bad key", `{"key":"Bad Key","name":"x"}`, http.StatusBadRequest},
		{"missing name", `{"key":"ok-key"}`, http.StatusBadRequest},
		{"rollout too high", `{"key":"ok-key","name":"x","rollout":101}`, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodPost, "/api/flags", strings.NewReader(tt.body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			if err := h.Create(e.NewContext(req, rec)); err != nil {
				t.Fatalf("Create: %v", err)
			}
			if rec.Code != tt.want {
				t.Errorf("status = %d, want %d (%s)", rec.Code, tt.want, rec.Body.String())
			}
		})
	}
}
