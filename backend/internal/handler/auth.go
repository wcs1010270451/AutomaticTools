package handler

import (
	"net/http"

	"automatictools/backend/internal/logic"

	"github.com/gin-gonic/gin"
)

func (h *Handler) SendRegistrationEmailCode(c *gin.Context) {
	var req logic.SendRegistrationEmailCodeRequest
	if !h.decodeOrFail(c, &req) {
		return
	}

	response, err := h.logic.SendRegistrationEmailCode(c.Request.Context(), req, requestMeta(c))
	if err != nil {
		h.handleErr(c, err)
		return
	}
	writeJSON(c, http.StatusOK, response)
}

func (h *Handler) Register(c *gin.Context) {
	var req logic.RegisterRequest
	if !h.decodeOrFail(c, &req) {
		return
	}

	response, err := h.logic.Register(c.Request.Context(), req, requestMeta(c))
	if err != nil {
		h.handleErr(c, err)
		return
	}
	writeJSON(c, http.StatusCreated, response)
}

func (h *Handler) Login(c *gin.Context) {
	var req logic.LoginRequest
	if !h.decodeOrFail(c, &req) {
		return
	}

	response, err := h.logic.Login(c.Request.Context(), req, requestMeta(c))
	if err != nil {
		h.handleErr(c, err)
		return
	}
	writeJSON(c, http.StatusOK, response)
}

func (h *Handler) AdminLogin(c *gin.Context) {
	var req logic.AdminLoginRequest
	if !h.decodeOrFail(c, &req) {
		return
	}

	response, err := h.logic.AdminLogin(c.Request.Context(), req, requestMeta(c))
	if err != nil {
		h.handleErr(c, err)
		return
	}
	writeJSON(c, http.StatusOK, response)
}
