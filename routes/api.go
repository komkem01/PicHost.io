package routes

import (
	"fmt"
	"net/http"
	"time"

	"pichost.io/app/modules"
	"pichost.io/app/modules/net/ratelimit"

	"github.com/gin-gonic/gin"
)

func WarpH(router *gin.RouterGroup, prefix string, handler http.Handler) {
	router.Any(fmt.Sprintf("%s/*w", prefix), gin.WrapH(http.StripPrefix(fmt.Sprintf("%s%s", router.BasePath(), prefix), handler)))
}

func apiStorage(r *gin.RouterGroup, mod *modules.Modules) {
	storagePublic := r.Group("/storage")
	{
		storagePublic.POST("/upload-file-guest", ratelimit.New(nil, "guest_upload", 5, time.Minute, ratelimit.IPKeyFunc), mod.Storage.Ctl.UploadFileGuest)
	}

	storageOptAuth := r.Group("/storage")
	storageOptAuth.Use(mod.Auth.Ctl.OptionalAuthMiddleware())
	{
		storageOptAuth.GET("/files/:id", mod.Storage.Ctl.GetFile)
		storageOptAuth.GET("/presign-url", mod.Storage.Ctl.GetPresignURL)
	}

	storageAuth := r.Group("/storage")
	storageAuth.Use(mod.Auth.Ctl.AuthMiddleware())
	{
		storageAuth.POST("/upload", ratelimit.New(nil, "upload", 20, time.Minute, ratelimit.UserOrIPKeyFunc), mod.Storage.Ctl.Upload)
		storageAuth.POST("/upload-file", ratelimit.New(nil, "upload_file", 20, time.Minute, ratelimit.UserOrIPKeyFunc), mod.Storage.Ctl.UploadFile)
		storageAuth.GET("/files", mod.Storage.Ctl.ListFiles)
		storageAuth.DELETE("/files/:id", mod.Storage.Ctl.DeleteFile)
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
		imageAuth.DELETE("/:id", mod.Storage.Ctl.DeleteFile)
	}
}

func apiPublic(r *gin.RouterGroup, mod *modules.Modules) {
	public := r.Group("/public")
	{
		public.GET("/plans", mod.Admin.Ctl.ListPublicPlanSettings)
		public.POST("/payments/webhook", mod.Payment.Ctl.ConfirmPaymentWebhook)

		auth := public.Group("/auth")
		auth.POST("/login", ratelimit.New(nil, "auth_login", 5, time.Minute, ratelimit.IPKeyFunc), mod.Auth.Ctl.Login)
		auth.POST("/register", ratelimit.New(nil, "auth_register", 5, time.Minute, ratelimit.IPKeyFunc), mod.Auth.Ctl.Register)
		auth.POST("/forgot-password", ratelimit.New(nil, "auth_forgot", 5, time.Minute, ratelimit.IPKeyFunc), mod.Auth.Ctl.ForgotPassword)
		auth.POST("/reset-password", ratelimit.New(nil, "auth_reset", 5, time.Minute, ratelimit.IPKeyFunc), mod.Auth.Ctl.ResetPassword)
		auth.POST("/verify-email", mod.Auth.Ctl.VerifyEmail)
		auth.POST("/refresh", mod.Auth.Ctl.Refresh)
		auth.POST("/logout", mod.Auth.Ctl.Logout)
		auth.GET("/google", mod.Auth.Ctl.GoogleLogin)
		auth.GET("/google/callback", mod.Auth.Ctl.GoogleCallback)
	}
}

func apiBilling(r *gin.RouterGroup, mod *modules.Modules) {
	billing := r.Group("/billing")
	billing.Use(mod.Auth.Ctl.AuthMiddleware())
	{
		billing.POST("/checkout", mod.Payment.Ctl.CreateCheckout)
		billing.GET("/payments", mod.Payment.Ctl.ListMyPayments)
		billing.GET("/payments/:id", mod.Payment.Ctl.GetMyPayment)
		billing.POST("/payments/:id/slip", mod.Payment.Ctl.SubmitSlip)
		billing.GET("/payment-methods", mod.Payment.Ctl.GetPaymentMethods)
		billing.POST("/cancel", mod.Payment.Ctl.CancelSubscription)
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
		auth.POST("/resend-verification", ratelimit.New(nil, "auth_resend_verify", 3, time.Minute, ratelimit.UserOrIPKeyFunc), mod.Auth.Ctl.ResendVerification)
		auth.GET("/quota", mod.Auth.Ctl.GetQuota)
	}
}

func apiAdmin(r *gin.RouterGroup, mod *modules.Modules) {
	adminGrp := r.Group("/admin")
	adminGrp.Use(mod.Auth.Ctl.AdminMiddleware())
	{
		adminGrp.GET("/stats", mod.Admin.Ctl.Stats)
		adminGrp.GET("/audit-logs", mod.Audit.Ctl.ListAuditLogs)

		images := adminGrp.Group("/images")
		{
			images.GET("", mod.Admin.Ctl.ListImages)
			images.DELETE("/:id", mod.Admin.Ctl.DeleteImage)
		}


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

		payments := adminGrp.Group("/payments")
		{
			payments.GET("", mod.Payment.Ctl.AdminListPayments)
			payments.PATCH("/:id/confirm", mod.Payment.Ctl.AdminConfirmPayment)
			payments.PATCH("/:id/refund", mod.Payment.Ctl.AdminRefundPayment)
		}
	}
}
