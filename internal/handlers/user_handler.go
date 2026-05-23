package handlers

import (
	"net/http"
	"strconv"

	apperrors "github.com/evelyndaianabejarano-coder/aluna-be/internal/errors"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/middleware"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/repository"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	userRepo repository.UserRepository
}

func NewUserHandler(userRepo repository.UserRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

type updateMeRequest struct {
	Nombre string `json:"nombre" binding:"required"`
}

func (h *UserHandler) Me(c *gin.Context) {
	userID, ok := c.Get(middleware.ContextKeyUserID)
	if !ok {
		apperrors.Respond(c, apperrors.ErrUnauthorized)
		return
	}

	user, err := h.userRepo.FindByID(userID.(uuid.UUID))
	if err != nil {
		apperrors.Respond(c, apperrors.ErrUserNotFound)
		return
	}

	c.JSON(http.StatusOK, services.UserDTO{
		ID:     user.ID.String(),
		Email:  user.Email,
		Nombre: user.Nombre,
		Role:   user.Role,
	})
}

func (h *UserHandler) UpdateMe(c *gin.Context) {
	userID, ok := c.Get(middleware.ContextKeyUserID)
	if !ok {
		apperrors.Respond(c, apperrors.ErrUnauthorized)
		return
	}

	var req updateMeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.Respond(c, apperrors.ErrValidation)
		return
	}

	user, err := h.userRepo.FindByID(userID.(uuid.UUID))
	if err != nil {
		apperrors.Respond(c, apperrors.ErrUserNotFound)
		return
	}

	user.Nombre = req.Nombre
	if err := h.userRepo.Update(user); err != nil {
		apperrors.Respond(c, apperrors.ErrInternal)
		return
	}

	c.JSON(http.StatusOK, services.UserDTO{
		ID:     user.ID.String(),
		Email:  user.Email,
		Nombre: user.Nombre,
		Role:   user.Role,
	})
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	users, total, err := h.userRepo.List(offset, limit)
	if err != nil {
		apperrors.Respond(c, apperrors.ErrInternal)
		return
	}

	dtos := make([]services.UserDTO, len(users))
	for i, u := range users {
		dtos[i] = services.UserDTO{
			ID:     u.ID.String(),
			Email:  u.Email,
			Nombre: u.Nombre,
			Role:   u.Role,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  dtos,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apperrors.Respond(c, apperrors.ErrNotFound)
		return
	}

	user, err := h.userRepo.FindByID(id)
	if err != nil {
		apperrors.Respond(c, apperrors.ErrUserNotFound)
		return
	}

	c.JSON(http.StatusOK, services.UserDTO{
		ID:     user.ID.String(),
		Email:  user.Email,
		Nombre: user.Nombre,
		Role:   user.Role,
	})
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apperrors.Respond(c, apperrors.ErrNotFound)
		return
	}

	if err := h.userRepo.Delete(id); err != nil {
		apperrors.Respond(c, apperrors.ErrUserNotFound)
		return
	}

	c.Status(http.StatusNoContent)
}
