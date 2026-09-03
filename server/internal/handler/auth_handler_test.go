package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func assertAuthBindingErrorIsPrivate(t *testing.T, handler gin.HandlerFunc) {
	t.Helper()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/", handler)

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{"))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	r.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}

	body := recorder.Body.String()
	if !strings.Contains(body, "参数校验失败") {
		t.Fatalf("body = %q, want public validation error message", body)
	}
	if strings.Contains(body, "invalid character") {
		t.Fatalf("body = %q, must not expose JSON parser details", body)
	}
}

func TestAuthRegisterBindingErrorIsPrivate(t *testing.T) {
	h := &AuthHandler{}
	assertAuthBindingErrorIsPrivate(t, h.Register)
}

func TestAuthLoginBindingErrorIsPrivate(t *testing.T) {
	h := &AuthHandler{}
	assertAuthBindingErrorIsPrivate(t, h.Login)
}

func TestAuthRefreshBindingErrorIsPrivate(t *testing.T) {
	h := &AuthHandler{}
	assertAuthBindingErrorIsPrivate(t, h.RefreshToken)
}
