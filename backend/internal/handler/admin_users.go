package handler

import (
	"net/http"
	"strconv"

	"automatictools/backend/internal/logic"

	"github.com/gin-gonic/gin"
)

func (h *Handler) AdminListUsers(c *gin.Context) {
	if _, err := h.currentAdmin(c); err != nil {
		h.handleErr(c, err)
		return
	}

	page, err := parsePositiveQueryInt(c.Query("page"), "page", 1)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	pageSize, err := parsePositiveQueryInt(c.Query("pageSize"), "pageSize", 20)
	if err != nil {
		h.handleErr(c, err)
		return
	}

	result, err := h.logic.ListUsers(c.Request.Context(), logic.AdminUserListQuery{
		Page:     page,
		PageSize: pageSize,
		Search:   c.Query("search"),
		Status:   c.Query("status"),
	})
	if err != nil {
		h.handleErr(c, err)
		return
	}
	writeJSON(c, http.StatusOK, result)
}

func parsePositiveQueryInt(raw string, name string, defaultValue int) (int, error) {
	if raw == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, logic.Error{
			Code:    logic.ErrorCodeBadRequest,
			Message: name + " 必须是大于 0 的整数。",
		}
	}
	return value, nil
}
