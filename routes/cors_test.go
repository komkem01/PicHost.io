package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// newCORSRouter wires the real buildAllowOriginFunc into a router so these tests
// exercise production behaviour rather than a re-implementation of it.
func newCORSRouter(t *testing.T, allowlist, environment string) *gin.Engine {
	t.Helper()

	allowOrigin, err := buildAllowOriginFunc(allowlist, environment)
	if err != nil {
		t.Fatalf("buildAllowOriginFunc(%q, %q) returned error: %v", allowlist, environment, err)
	}

	router := gin.New()
	router.Use(cors.New(cors.Config{
		AllowOriginFunc:  allowOrigin,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	router.GET("/api/v1/auth/me", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	return router
}

func TestCORS_PreflightAndVercelOrigins(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := newCORSRouter(t, "http://localhost:3000,*.vercel.app", "production")

	// 1. OPTIONS preflight from the deployed frontend must be accepted.
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodOptions, "/api/v1/auth/me", nil)
	req1.Header.Set("Origin", "https://pichost-web.vercel.app")
	req1.Header.Set("Access-Control-Request-Method", "GET")
	req1.Header.Set("Access-Control-Request-Headers", "Authorization")
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusNoContent && w1.Code != http.StatusOK {
		t.Errorf("expected 204 or 200 for OPTIONS preflight, got %d", w1.Code)
	}
	if got := w1.Header().Get("Access-Control-Allow-Origin"); got != "https://pichost-web.vercel.app" {
		t.Errorf("expected Access-Control-Allow-Origin to echo the origin, got %q", got)
	}
	if got := w1.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("expected Access-Control-Allow-Credentials to be true, got %q", got)
	}

	// 2. Actual GET request carries the same headers.
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req2.Header.Set("Origin", "https://pichost-web.vercel.app")
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("expected 200 OK for GET request, got %d", w2.Code)
	}
	if got := w2.Header().Get("Access-Control-Allow-Origin"); got != "https://pichost-web.vercel.app" {
		t.Errorf("expected Access-Control-Allow-Origin on GET response, got %q", got)
	}
}

func TestCORS_RejectsOriginOutsideAllowlist(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := newCORSRouter(t, "https://pichost.io", "production")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for origin outside allowlist, got %d", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("expected no Access-Control-Allow-Origin for rejected origin, got %q", got)
	}
}

func TestBuildAllowOriginFunc_Matching(t *testing.T) {
	allowOrigin, err := buildAllowOriginFunc(
		" https://pichost.io/ , *.vercel.app , http://localhost:3000 ",
		"production",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cases := []struct {
		origin string
		want   bool
		why    string
	}{
		{"https://pichost.io", true, "exact match after trailing-slash and space trimming"},
		{"https://PicHost.io", true, "origin comparison is case-insensitive"},
		{"https://pichost-web.vercel.app", true, "wildcard suffix match"},
		{"http://localhost:3000", true, "port is part of the exact match"},
		{"https://pichost.io.evil.com", false, "suffix of an exact entry must not match"},
		{"https://notvercel.app", false, "wildcard requires the dot-prefixed suffix"},
		{"http://localhost:3001", false, "different port is a different origin"},
		{"", false, "empty origin is not a CORS origin"},
	}

	for _, tc := range cases {
		if got := allowOrigin(tc.origin); got != tc.want {
			t.Errorf("allowOrigin(%q) = %v, want %v (%s)", tc.origin, got, tc.want, tc.why)
		}
	}
}

func TestBuildAllowOriginFunc_RejectsUnsafeConfig(t *testing.T) {
	if _, err := buildAllowOriginFunc("*", "production"); err == nil {
		t.Error("expected wildcard-all to be rejected in production, got nil error")
	}
	if _, err := buildAllowOriginFunc("", "production"); err == nil {
		t.Error("expected empty allowlist to be rejected, got nil error")
	}
	if _, err := buildAllowOriginFunc("  ,  ", "local"); err == nil {
		t.Error("expected whitespace-only allowlist to be rejected, got nil error")
	}
}

func TestBuildAllowOriginFunc_WildcardAllOutsideProduction(t *testing.T) {
	allowOrigin, err := buildAllowOriginFunc("*", "local")
	if err != nil {
		t.Fatalf("expected wildcard-all to be permitted outside production, got error: %v", err)
	}
	if !allowOrigin("https://anything.example.com") {
		t.Error("expected wildcard-all to accept any origin in local environment")
	}
}
