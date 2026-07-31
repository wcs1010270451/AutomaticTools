package handler

import (
	"net/http"
	"strconv"

	"automatictools/backend/internal/logic"

	"github.com/gin-gonic/gin"
)

func (h *Handler) AdminListAdmins(c *gin.Context) {
	if _, err := h.currentAdmin(c); err != nil {
		h.handleErr(c, err)
		return
	}

	admins, err := h.logic.ListAdmins(c.Request.Context())
	if err != nil {
		h.handleErr(c, err)
		return
	}
	writeJSON(c, http.StatusOK, map[string]any{"admins": admins})
}

func (h *Handler) AdminCreateAdmin(c *gin.Context) {
	actorAdminID, err := h.currentAdmin(c)
	if err != nil {
		h.handleErr(c, err)
		return
	}

	var req logic.CreateAdminRequest
	if !h.decodeOrFail(c, &req) {
		return
	}

	admin, err := h.logic.CreateAdmin(
		c.Request.Context(),
		actorAdminID,
		req,
		requestMeta(c),
	)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	writeJSON(c, http.StatusCreated, map[string]any{"admin": admin})
}

func (h *Handler) AdminUpdateAdmin(c *gin.Context) {
	actorAdminID, err := h.currentAdmin(c)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	adminID, ok := h.adminIDParam(c)
	if !ok {
		return
	}

	var req logic.UpdateAdminRequest
	if !h.decodeOrFail(c, &req) {
		return
	}

	admin, err := h.logic.UpdateAdmin(
		c.Request.Context(),
		actorAdminID,
		adminID,
		req,
		requestMeta(c),
	)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	writeJSON(c, http.StatusOK, map[string]any{"admin": admin})
}

func (h *Handler) AdminDeleteAdmin(c *gin.Context) {
	actorAdminID, err := h.currentAdmin(c)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	adminID, ok := h.adminIDParam(c)
	if !ok {
		return
	}

	if err := h.logic.DeleteAdmin(c.Request.Context(), actorAdminID, adminID, requestMeta(c)); err != nil {
		h.handleErr(c, err)
		return
	}
	writeJSON(c, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) adminIDParam(c *gin.Context) (int64, bool) {
	adminID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || adminID <= 0 {
		h.handleErr(c, logic.Error{
			Code:    logic.ErrorCodeBadRequest,
			Message: "管理员ID无效。",
		})
		return 0, false
	}
	return adminID, true
}
