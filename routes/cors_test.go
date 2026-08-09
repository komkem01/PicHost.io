package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func TestCORS_PreflightAndVercelOrigins(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	allowAll := false
	allowedOrigins := map[string]bool{
		"http://localhost:3000": true,
	}

	router.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			if allowAll || origin == "" {
				return true
			}
			cleanOrigin := origin
			if len(cleanOrigin) > 0 && cleanOrigin[len(cleanOrigin)-1] == '/' {
				cleanOrigin = cleanOrigin[:len(cleanOrigin)-1]
			}
			if allowedOrigins[cleanOrigin] {
				return true
			}
			if len(cleanOrigin) >= 11 && cleanOrigin[len(cleanOrigin)-11:] == ".vercel.app" {
				return true
			}
			return false
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.GET("/api/v1/auth/me", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 1. Test OPTIONS preflight request from pichost-web.vercel.app
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodOptions, "/api/v1/auth/me", nil)
	req1.Header.Set("Origin", "https://pichost-web.vercel.app")
	req1.Header.Set("Access-Control-Request-Method", "GET")
	req1.Header.Set("Access-Control-Request-Headers", "Authorization")
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusNoContent && w1.Code != http.StatusOK {
		t.Errorf("expected 204 or 200 for OPTIONS preflight, got %d", w1.Code)
	}
	if w1.Header().Get("Access-Control-Allow-Origin") != "https://pichost-web.vercel.app" {
		t.Errorf("expected Access-Control-Allow-Origin header to be https://pichost-web.vercel.app, got %s", w1.Header().Get("Access-Control-Allow-Origin"))
	}
	if w1.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Errorf("expected Access-Control-Allow-Credentials to be true, got %s", w1.Header().Get("Access-Control-Allow-Credentials"))
	}

	// 2. Test GET request with origin
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req2.Header.Set("Origin", "https://pichost-web.vercel.app")
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("expected 200 OK for GET request, got %d", w2.Code)
	}
	if w2.Header().Get("Access-Control-Allow-Origin") != "https://pichost-web.vercel.app" {
		t.Errorf("expected Access-Control-Allow-Origin header on GET response, got %s", w2.Header().Get("Access-Control-Allow-Origin"))
	}
}
