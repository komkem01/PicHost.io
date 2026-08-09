package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAuthorizationMatrix_GuestEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	api := router.Group("/api/v1")
	// Public endpoint simulation
	api.GET("/public/plans", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Protected admin endpoint simulation
	admin := api.Group("/admin")
	admin.Use(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		if authHeader != "Bearer admin-token" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	})
	admin.GET("/stats", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"stats": true})
	})

	// 1. Guest accesses public endpoint -> 200 OK
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/api/v1/public/plans", nil)
	router.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Errorf("expected 200 OK for guest accessing public plans, got %d", w1.Code)
	}

	// 2. Guest accesses admin endpoint -> 401 Unauthorized
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/stats", nil)
	router.ServeHTTP(w2, req2)
	if w2.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized for guest accessing admin endpoint, got %d", w2.Code)
	}

	// 3. User with non-admin token accesses admin endpoint -> 403 Forbidden
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/stats", nil)
	req3.Header.Set("Authorization", "Bearer user-token")
	router.ServeHTTP(w3, req3)
	if w3.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for standard user accessing admin endpoint, got %d", w3.Code)
	}

	// 4. Admin token accesses admin endpoint -> 200 OK
	w4 := httptest.NewRecorder()
	req4, _ := http.NewRequest(http.MethodGet, "/api/v1/admin/stats", nil)
	req4.Header.Set("Authorization", "Bearer admin-token")
	router.ServeHTTP(w4, req4)
	if w4.Code != http.StatusOK {
		t.Errorf("expected 200 OK for admin token accessing admin endpoint, got %d", w4.Code)
	}
}
