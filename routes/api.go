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
		storage.POST("/upload-file-guest", mod.Storage.Ctl.UploadFileGuest)
		storage.GET("/files", mod.Storage.Ctl.ListFiles)
		storage.GET("/files/:id", mod.Storage.Ctl.GetFile)
		storage.GET("/presign-url", mod.Storage.Ctl.GetPresignURL)
		storage.DELETE("/files/:id", mod.Storage.Ctl.DeleteFile)
	}

	storageAuth := r.Group("/storage")
	storageAuth.Use(mod.Auth.Ctl.AuthMiddleware())
	{
		storageAuth.POST("/upload-file", mod.Storage.Ctl.UploadFile)
	}
}

func apiImage(r *gin.RouterGroup, mod *modules.Modules) {
	image := r.Group("/images")
	image.Use(mod.Auth.Ctl.OptionalAuthMiddleware())
	{
		image.POST("", mod.Image.Ctl.CreateImage)
		image.GET("/:id", mod.Image.Ctl.GetImage)
		image.GET("/presign-url", mod.Image.Ctl.GetPresignURL)
	}

	imageAuth := r.Group("/images")
	imageAuth.Use(mod.Auth.Ctl.AuthMiddleware())
	{
		imageAuth.GET("", mod.Image.Ctl.ListImages)
	}
}

func apiPublic(r *gin.RouterGroup, mod *modules.Modules) {
	public := r.Group("/public")
	{
		public.GET("/plans", mod.Admin.Ctl.ListPublicPlanSettings)

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
		auth.PATCH("/me", mod.Auth.Ctl.UpdateMe)
		auth.DELETE("/me", mod.Auth.Ctl.DeleteMe)
		auth.PATCH("/change-password", mod.Auth.Ctl.ChangePassword)
		auth.GET("/quota", mod.Auth.Ctl.GetQuota)
	}
}

func apiAdmin(r *gin.RouterGroup, mod *modules.Modules) {
	adminGrp := r.Group("/admin")
	adminGrp.Use(mod.Auth.Ctl.AdminMiddleware())
	{
		adminGrp.GET("/stats", mod.Admin.Ctl.Stats)

		plans := adminGrp.Group("/plans")
		{
			plans.GET("", mod.Admin.Ctl.ListPlanSettings)
			plans.GET("/:key", mod.Admin.Ctl.GetPlanSetting)
			plans.PATCH("/:key", mod.Admin.Ctl.UpsertPlanSetting)
			plans.DELETE("/:key", mod.Admin.Ctl.DeletePlanSetting)
		}

		users := adminGrp.Group("/users")
		{
			users.GET("", mod.Admin.Ctl.ListUsers)
			users.GET("/:id", mod.Admin.Ctl.GetUser)
			users.PATCH("/:id/profile", mod.Admin.Ctl.UpdateProfile)
			users.PATCH("/:id/plan", mod.Admin.Ctl.SetUserPlan)
			users.PATCH("/:id/active", mod.Admin.Ctl.SetUserActive)
			users.PATCH("/:id/admin", mod.Admin.Ctl.SetUserAdmin)
			users.DELETE("/:id", mod.Admin.Ctl.DeleteUser)
		}
	}
}
