package routes

import (
	"fmt"
	"net/http"

	"pichost.io/app/modules"

	"github.com/gin-gonic/gin"
)

func WarpH(router *gin.RouterGroup, prefix string, handler http.Handler) {
	router.Any(fmt.Sprintf("%s/*w", prefix), gin.WrapH(http.StripPrefix(fmt.Sprintf("%s%s", router.BasePath(), prefix), handler)))
}

func api(r *gin.RouterGroup, mod *modules.Modules) {
	r.GET("/example/:id", mod.Example.Ctl.Get)
	r.GET("/example-http", mod.Example.Ctl.GetHttpReq)
	r.POST("/example", mod.Example.Ctl.Create)
}

func apiUser(r *gin.RouterGroup, mod *modules.Modules) {
	users := r.Group("/users")
	{
		users.POST("", mod.Users.Ctl.Create)
		users.GET("", mod.Users.Ctl.List)
		users.GET("/email/:email", mod.Users.Ctl.GetByEmail)
		users.GET("/:id", mod.Users.Ctl.Get)
		users.PATCH("/:id", mod.Users.Ctl.Update)
		users.DELETE("/:id", mod.Users.Ctl.Delete)
	}
}

func apiStorage(r *gin.RouterGroup, mod *modules.Modules) {
	storage := r.Group("/storage")
	{
		storage.POST("/upload", mod.Storage.Ctl.Upload)
		storage.GET("/files", mod.Storage.Ctl.ListFiles)
		storage.GET("/files/:id", mod.Storage.Ctl.GetFile)
		storage.GET("/presign-url", mod.Storage.Ctl.GetPresignURL)
		storage.DELETE("/files/:id", mod.Storage.Ctl.DeleteFile)
	}
}

func apiImage(r *gin.RouterGroup, mod *modules.Modules) {
	image := r.Group("/images")
	{
		image.GET("/:id", mod.Image.Ctl.GetImage)
		image.GET("/presign-url", mod.Image.Ctl.GetPresignURL)
	}
}

func apiPublic(r *gin.RouterGroup, mod *modules.Modules) {
	public := r.Group("/public")
	{
		auth := public.Group("/auth")
		auth.POST("/login", mod.Auth.Ctl.Login)
		auth.POST("/register", mod.Auth.Ctl.Register)
		auth.POST("/refresh", mod.Auth.Ctl.Refresh)
		auth.POST("/logout", mod.Auth.Ctl.Logout)
		auth.GET("/google", mod.Auth.Ctl.GoogleLogin)
		auth.GET("/google/callback", mod.Auth.Ctl.GoogleCallback)
	}
}

func apiAuth(r *gin.RouterGroup, mod *modules.Modules) {
	auth := r.Group("/auth")
	auth.Use(mod.Auth.Ctl.AuthMiddleware())
	{
		auth.GET("/me", mod.Auth.Ctl.Me)
	}
}
