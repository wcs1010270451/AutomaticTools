package handler

import (
	"errors"
	"net/http"

	"automatictools/backend/internal/logic"
	"automatictools/backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

type errorResponse struct {
	Error     string `json:"error"`
	RequestID string `json:"requestId"`
}

func decodeJSON(c *gin.Context, target any) error {
	return c.ShouldBindJSON(target)
}

func writeJSON(c *gin.Context, status int, payload any) {
	c.JSON(status, payload)
}

func (h *Handler) handleErr(c *gin.Context, err error) {
	var appErr logic.Error
	if errors.As(err, &appErr) {
		status := statusForError(appErr.Code)
		h.logger.Warn(
			"request failed",
			"requestId", middleware.RequestID(c),
			"status", status,
			"error", appErr.Message,
			"cause", appErr.Cause,
		)
		writeJSON(c, status, errorResponse{
			Error:     appErr.Message,
			RequestID: middleware.RequestID(c),
		})
		return
	}

	h.logger.Error(
		"unexpected error",
		"requestId", middleware.RequestID(c),
		"error", err,
	)
	writeJSON(c, http.StatusInternalServerError, errorResponse{
		Error:     "服务器内部错误。",
		RequestID: middleware.RequestID(c),
	})
}

func (h *Handler) decodeOrFail(c *gin.Context, target any) bool {
	if err := decodeJSON(c, target); err != nil {
		h.handleErr(c, logic.Error{
			Code:    logic.ErrorCodeBadRequest,
			Message: "请求体必须是合法 JSON。",
		})
		return false
	}
	return true
}

func statusForError(code logic.ErrorCode) int {
	switch code {
	case logic.ErrorCodeBadRequest:
		return http.StatusBadRequest
	case logic.ErrorCodeUnauthorized:
		return http.StatusUnauthorized
	case logic.ErrorCodeForbidden:
		return http.StatusForbidden
	case logic.ErrorCodeConflict:
		return http.StatusConflict
	case logic.ErrorCodeNotFound:
		return http.StatusNotFound
	case logic.ErrorCodeUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
