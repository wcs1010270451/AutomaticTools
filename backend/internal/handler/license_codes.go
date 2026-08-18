package handler

import (
	"net/http"
	"strconv"
	"strings"

	"automatictools/backend/internal/logic"

	"github.com/gin-gonic/gin"
)

func (h *Handler) RedeemLicenseCode(c *gin.Context) {
	userID, err := h.currentUser(c)
	if err != nil {
		h.handleErr(c, err)
		return
	}

	var req logic.RedeemLicenseCodeRequest
	if !h.decodeOrFail(c, &req) {
		return
	}
	result, err := h.logic.RedeemLicenseCode(c.Request.Context(), userID, req, requestMeta(c))
	if err != nil {
		h.handleErr(c, err)
		return
	}
	writeJSON(c, http.StatusOK, result)
}

func (h *Handler) AdminGenerateLicenseCodes(c *gin.Context) {
	adminID, err := h.currentAdmin(c)
	if err != nil {
		h.handleErr(c, err)
		return
	}

	var req logic.GenerateLicenseCodesRequest
	if !h.decodeOrFail(c, &req) {
		return
	}
	batchNo, codes, err := h.logic.GenerateLicenseCodes(
		c.Request.Context(),
		adminID,
		req,
		requestMeta(c),
	)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	writeJSON(c, http.StatusCreated, map[string]any{
		"batchNo": batchNo,
		"codes":   codes,
	})
}

func (h *Handler) AdminListLicenseCodes(c *gin.Context) {
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
	if pageSize > 100 {
		h.handleErr(c, logic.Error{Code: logic.ErrorCodeBadRequest, Message: "pageSize 不能超过 100。"})
		return
	}

	result, err := h.logic.ListLicenseCodes(c.Request.Context(), logic.LicenseCodeListFilter{
		Page:     page,
		PageSize: pageSize,
		Status:   c.Query("status"),
		ToolCode: c.Query("toolCode"),
		Search:   c.Query("search"),
	})
	if err != nil {
		h.handleErr(c, err)
		return
	}
	writeJSON(c, http.StatusOK, result)
}

func (h *Handler) AdminRevokeLicenseCode(c *gin.Context) {
	adminID, err := h.currentAdmin(c)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	licenseCodeID, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || licenseCodeID <= 0 {
		h.handleErr(c, logic.Error{Code: logic.ErrorCodeBadRequest, Message: "授权码ID无效。"})
		return
	}

	code, err := h.logic.RevokeLicenseCode(
		c.Request.Context(),
		adminID,
		licenseCodeID,
		requestMeta(c),
	)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	writeJSON(c, http.StatusOK, map[string]any{"code": code})
}
