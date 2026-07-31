package router

import (
	"log/slog"
	"net/http"

	"automatictools/backend/internal/config"
	"automatictools/backend/internal/handler"
	"automatictools/backend/internal/logic"
	"automatictools/backend/internal/mailer"
	"automatictools/backend/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gorm.io/gorm"
)

type Dependencies struct {
	Config config.Config
	DB     *gorm.DB
	Logger *slog.Logger
}

func New(deps Dependencies) http.Handler {
	binding.EnableDecoderDisallowUnknownFields = true
	gin.SetMode(gin.ReleaseMode)

	service := logic.New(logic.Dependencies{
		Config:      deps.Config,
		DB:          deps.DB,
		EmailSender: mailer.NewSMTP(deps.Config),
	})
	handlers := handler.New(service, deps.Logger)

	engine := gin.New()
	engine.HandleMethodNotAllowed = true
	_ = engine.SetTrustedProxies(nil)
	engine.Use(middleware.RequestContext(deps.Logger))
	engine.Use(middleware.Recovery(deps.Logger))
	registerRoutes(engine, handlers)
	return engine
}

func registerRoutes(engine *gin.Engine, handlers *handler.Handler) {
	engine.GET("/health", handlers.Health)

	api := engine.Group("/api")

	// App APIs
	api.POST("/auth/register", handlers.Register)
	api.POST("/auth/email-code", handlers.SendRegistrationEmailCode)
	api.POST("/auth/login", handlers.Login)
	api.GET("/me", handlers.Me)
	api.GET("/products", handlers.Products)
	api.GET("/tools", handlers.Tools)
	api.POST("/orders", handlers.CreateOrder)
	api.GET("/me/orders", handlers.MyOrders)
	api.GET("/me/entitlements", handlers.MyEntitlements)
	api.POST("/devices/bind", handlers.BindDevice)

	// Admin APIs
	api.POST("/admin/auth/login", handlers.AdminLogin)
	api.GET("/admin/users", handlers.AdminListUsers)
	api.GET("/admin/admins", handlers.AdminListAdmins)
	api.POST("/admin/admins", handlers.AdminCreateAdmin)
	api.PUT("/admin/admins/:id", handlers.AdminUpdateAdmin)
	api.DELETE("/admin/admins/:id", handlers.AdminDeleteAdmin)
	api.GET("/admin/tools", handlers.AdminListTools)
	api.POST("/admin/tools", handlers.AdminCreateTool)
	api.PUT("/admin/tools/:code", handlers.AdminUpdateTool)
	api.POST("/admin/entitlements/grant", handlers.AdminGrantEntitlement)
	api.POST("/admin/orders/confirm", handlers.AdminConfirmOrder)
}
