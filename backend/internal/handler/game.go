package handler

import (
	"net/http"
	"strconv"

	"automatictools/backend/internal/logic"

	"github.com/gin-gonic/gin"
)

// GameInit initializes or retrieves the player's game profile.
func (h *Handler) GameInit(c *gin.Context) {
	userID, err := h.currentUser(c)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	player, err := h.logic.GetOrCreateGamePlayer(c.Request.Context(), userID)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": player})
}

// GameOpenBox handles box opening.
func (h *Handler) GameOpenBox(c *gin.Context) {
	userID, err := h.currentUser(c)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	result, err := h.logic.OpenBox(c.Request.Context(), userID)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

// GameListEquipments returns all player equipment.
func (h *Handler) GameListEquipments(c *gin.Context) {
	userID, err := h.currentUser(c)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	equipments, err := h.logic.ListPlayerEquipments(c.Request.Context(), userID)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": equipments})
}

// GameEquip handles equipping an item.
func (h *Handler) GameEquip(c *gin.Context) {
	userID, err := h.currentUser(c)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	var req logic.EquipRequest
	if !h.decodeOrFail(c, &req) {
		return
	}
	if err := h.logic.EquipItem(c.Request.Context(), userID, req); err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": "ok"})
}

// GameGetOpponents returns available opponents for PK.
func (h *Handler) GameGetOpponents(c *gin.Context) {
	userID, err := h.currentUser(c)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	opponents, err := h.logic.GetOpponents(c.Request.Context(), userID)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": opponents})
}

// GameChallenge initiates a PK battle.
func (h *Handler) GameChallenge(c *gin.Context) {
	userID, err := h.currentUser(c)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	var req logic.ChallengeRequest
	if !h.decodeOrFail(c, &req) {
		return
	}
	result, err := h.logic.Challenge(c.Request.Context(), userID, req)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}

// GameLeaderboard returns ranking data.
func (h *Handler) GameLeaderboard(c *gin.Context) {
	rankType := c.DefaultQuery("type", "power")
	limitStr := c.DefaultQuery("limit", "50")
	limit := 50
	if n, err := strconv.Atoi(limitStr); err == nil && n > 0 {
		limit = n
	}

	entries, err := h.logic.GetLeaderboard(c.Request.Context(), rankType, limit)
	if err != nil {
		h.handleErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": entries})
}
