package handlers

import (
	"net/http"

	apperrors "github.com/evelyndaianabejarano-coder/aluna-be/internal/errors"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/services"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService services.AuthService
}

func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type registerRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Nombre   string `json:"nombre"   binding:"required"`
}

type loginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type logoutRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.Respond(c, apperrors.ErrValidation)
		return
	}

	resp, err := h.authService.Register(req.Email, req.Password, req.Nombre)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			apperrors.Respond(c, appErr)
			return
		}
		apperrors.Respond(c, apperrors.ErrInternal)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.Respond(c, apperrors.ErrValidation)
		return
	}

	resp, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			apperrors.Respond(c, appErr)
			return
		}
		apperrors.Respond(c, apperrors.ErrInternal)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.Respond(c, apperrors.ErrValidation)
		return
	}

	tokens, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			apperrors.Respond(c, appErr)
			return
		}
		apperrors.Respond(c, apperrors.ErrInternal)
		return
	}

	c.JSON(http.StatusOK, tokens)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	var req logoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.Respond(c, apperrors.ErrValidation)
		return
	}

	if err := h.authService.Logout(req.RefreshToken); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			apperrors.Respond(c, appErr)
			return
		}
		apperrors.Respond(c, apperrors.ErrInternal)
		return
	}

	c.Status(http.StatusNoContent)
}
