package handler

import (
	"net/http"

	"automatictools/backend/internal/logic"

	"github.com/gin-gonic/gin"
)

func (h *Handler) Health(c *gin.Context) {
	writeJSON(c, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) Me(c *gin.Context) {
	userID, err := h.currentUser(c)
	if err != nil {
		h.handleErr(c, err)
		return
	}

	user, err := h.logic.GetUser(c.Request.Context(), userID)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	writeJSON(c, http.StatusOK, map[string]any{"user": user})
}

func (h *Handler) Products(c *gin.Context) {
	h.Tools(c)
}

func (h *Handler) Tools(c *gin.Context) {
	tools, err := h.logic.ListTools(c.Request.Context())
	if err != nil {
		h.handleErr(c, err)
		return
	}
	writeJSON(c, http.StatusOK, map[string]any{"tools": tools})
}

func (h *Handler) CreateOrder(c *gin.Context) {
	userID, err := h.currentUser(c)
	if err != nil {
		h.handleErr(c, err)
		return
	}

	var req logic.CreateOrderRequest
	if !h.decodeOrFail(c, &req) {
		return
	}

	order, err := h.logic.CreateOrder(c.Request.Context(), userID, req, requestMeta(c))
	if err != nil {
		h.handleErr(c, err)
		return
	}
	writeJSON(c, http.StatusCreated, map[string]any{"order": order})
}

func (h *Handler) MyOrders(c *gin.Context) {
	userID, err := h.currentUser(c)
	if err != nil {
		h.handleErr(c, err)
		return
	}

	orders, err := h.logic.ListOrders(c.Request.Context(), userID)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	writeJSON(c, http.StatusOK, map[string]any{"orders": orders})
}

func (h *Handler) MyEntitlements(c *gin.Context) {
	userID, err := h.currentUser(c)
	if err != nil {
		h.handleErr(c, err)
		return
	}

	entitlements, err := h.logic.ListEntitlements(c.Request.Context(), userID)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	writeJSON(c, http.StatusOK, map[string]any{"entitlements": entitlements})
}

func (h *Handler) BindDevice(c *gin.Context) {
	userID, err := h.currentUser(c)
	if err != nil {
		h.handleErr(c, err)
		return
	}

	var req logic.BindDeviceRequest
	if !h.decodeOrFail(c, &req) {
		return
	}

	if err := h.logic.BindDevice(c.Request.Context(), userID, req); err != nil {
		h.handleErr(c, err)
		return
	}
	writeJSON(c, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) AdminGrantEntitlement(c *gin.Context) {
	adminID, err := h.currentAdmin(c)
	if err != nil {
		h.handleErr(c, err)
		return
	}

	var req logic.GrantEntitlementRequest
	if !h.decodeOrFail(c, &req) {
		return
	}

	err = h.logic.GrantEntitlement(
		c.Request.Context(),
		adminID,
		req,
		requestMeta(c),
	)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	writeJSON(c, http.StatusOK, map[string]any{"ok": true})
}

func (h *Handler) AdminConfirmOrder(c *gin.Context) {
	adminID, err := h.currentAdmin(c)
	if err != nil {
		h.handleErr(c, err)
		return
	}

	var req logic.ConfirmOrderRequest
	if !h.decodeOrFail(c, &req) {
		return
	}

	err = h.logic.ConfirmOrder(
		c.Request.Context(),
		adminID,
		req,
		requestMeta(c),
	)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	writeJSON(c, http.StatusOK, map[string]any{"ok": true})
}
