package handler

import (
	"log/slog"

	"automatictools/backend/internal/logic"
	"automatictools/backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	logic  *logic.Service
	logger *slog.Logger
}

func New(service *logic.Service, logger *slog.Logger) *Handler {
	return &Handler{
		logic:  service,
		logger: logger,
	}
}

func (h *Handler) currentUser(c *gin.Context) (int64, error) {
	userID, _, err := h.logic.Authenticate(c.Request.Context(), c.GetHeader("Authorization"))
	return userID, err
}

func (h *Handler) currentAdmin(c *gin.Context) (int64, error) {
	adminID, _, err := h.logic.AuthenticateAdmin(c.Request.Context(), c.GetHeader("Authorization"))
	return adminID, err
}

func requestMeta(c *gin.Context) logic.RequestMeta {
	return logic.RequestMeta{
		IP:        middleware.ClientIP(c.Request),
		UserAgent: c.Request.UserAgent(),
	}
}
