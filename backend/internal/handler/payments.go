package handler

import (
	"net/http"
	"strings"

	"automatictools/backend/internal/logic"
	"automatictools/backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

func (h *Handler) MyPurchases(c *gin.Context) {
	userID, err := h.currentUser(c)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	purchases, err := h.logic.ListPurchases(c.Request.Context(), userID)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	writeJSON(c, http.StatusOK, map[string]any{"purchases": purchases})
}

func (h *Handler) CreateAlipayPayment(c *gin.Context) {
	userID, err := h.currentUser(c)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	var req logic.CreateAlipayPaymentRequest
	if !h.decodeOrFail(c, &req) {
		return
	}
	session, err := h.logic.CreateAlipayPayment(
		c.Request.Context(),
		userID,
		req,
		requestMeta(c),
	)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	writeJSON(c, http.StatusCreated, session)
}

func (h *Handler) PaymentOrderStatus(c *gin.Context) {
	userID, err := h.currentUser(c)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	order, err := h.logic.GetPaymentOrder(
		c.Request.Context(),
		userID,
		c.Param("orderNo"),
	)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	writeJSON(c, http.StatusOK, map[string]any{"order": order})
}

func (h *Handler) AlipayNotification(c *gin.Context) {
	if err := c.Request.ParseForm(); err != nil {
		c.String(http.StatusBadRequest, "failure")
		return
	}
	err := h.logic.HandleAlipayNotification(
		c.Request.Context(),
		c.Request.PostForm,
		requestMeta(c),
	)
	if err != nil {
		h.logger.Warn(
			"alipay notification rejected",
			"requestId", middleware.RequestID(c),
			"error", err,
			"orderNo", strings.TrimSpace(c.Request.PostForm.Get("out_trade_no")),
		)
		c.String(http.StatusBadRequest, "failure")
		return
	}
	c.String(http.StatusOK, "success")
}
