package handler

import (
	"net/http"

	"automatictools/backend/internal/logic"

	"github.com/gin-gonic/gin"
)

func (h *Handler) AdminListTools(c *gin.Context) {
	if _, err := h.currentAdmin(c); err != nil {
		h.handleErr(c, err)
		return
	}

	tools, err := h.logic.ListAdminTools(c.Request.Context())
	if err != nil {
		h.handleErr(c, err)
		return
	}
	writeJSON(c, http.StatusOK, map[string]any{"tools": tools})
}

func (h *Handler) AdminCreateTool(c *gin.Context) {
	actorAdminID, err := h.currentAdmin(c)
	if err != nil {
		h.handleErr(c, err)
		return
	}

	var req logic.CreateToolRequest
	if !h.decodeOrFail(c, &req) {
		return
	}

	tool, err := h.logic.CreateTool(c.Request.Context(), actorAdminID, req, requestMeta(c))
	if err != nil {
		h.handleErr(c, err)
		return
	}
	writeJSON(c, http.StatusCreated, map[string]any{"tool": tool})
}

func (h *Handler) AdminUpdateTool(c *gin.Context) {
	actorAdminID, err := h.currentAdmin(c)
	if err != nil {
		h.handleErr(c, err)
		return
	}

	var req logic.UpdateToolRequest
	if !h.decodeOrFail(c, &req) {
		return
	}

	tool, err := h.logic.UpdateTool(
		c.Request.Context(),
		actorAdminID,
		c.Param("code"),
		req,
		requestMeta(c),
	)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	writeJSON(c, http.StatusOK, map[string]any{"tool": tool})
}
