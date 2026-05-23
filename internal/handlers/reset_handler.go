package handlers

import (
	"net/http"

	apperrors "github.com/evelyndaianabejarano-coder/aluna-be/internal/errors"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/services"
	"github.com/gin-gonic/gin"
)

type ResetHandler struct {
	resetService services.ResetService
}

func NewResetHandler(resetService services.ResetService) *ResetHandler {
	return &ResetHandler{resetService: resetService}
}

type forgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type resetPasswordRequest struct {
	Token       string `json:"token"       binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=8"`
}

func (h *ResetHandler) ForgotPassword(c *gin.Context) {
	var req forgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.Respond(c, apperrors.ErrValidation)
		return
	}

	// Siempre 204 — no revelamos si el email existe
	_ = h.resetService.ForgotPassword(req.Email)
	c.Status(http.StatusNoContent)
}

func (h *ResetHandler) ResetPassword(c *gin.Context) {
	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.Respond(c, apperrors.ErrValidation)
		return
	}

	if err := h.resetService.ResetPassword(req.Token, req.NewPassword); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			apperrors.Respond(c, appErr)
			return
		}
		apperrors.Respond(c, apperrors.ErrInternal)
		return
	}

	c.Status(http.StatusNoContent)
}
