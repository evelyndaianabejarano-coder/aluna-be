package handlers

import (
	"net/http"
	"strconv"
	"time"

	apperrors "github.com/evelyndaianabejarano-coder/aluna-be/internal/errors"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/middleware"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/models"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/repository"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ClassHandler struct {
	classSvc services.ClassService
}

func NewClassHandler(classSvc services.ClassService) *ClassHandler {
	return &ClassHandler{classSvc: classSvc}
}

func (h *ClassHandler) List(c *gin.Context) {
	filters := repository.ClassFilters{}

	if v := c.Query("categoria"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filters.CategoriaID = &id
		}
	}
	if v := c.Query("profesor"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			filters.ProfesorID = &id
		}
	}
	if v := c.Query("modalidad"); v != "" {
		m := models.Modalidad(v)
		filters.Modalidad = &m
	}
	if v := c.Query("fecha"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			filters.Fecha = &t
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	filters.Offset = (page - 1) * limit
	filters.Limit = limit

	classes, total, err := h.classSvc.List(filters)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			apperrors.Respond(c, appErr)
			return
		}
		apperrors.Respond(c, apperrors.ErrInternal)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  classes,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *ClassHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apperrors.Respond(c, apperrors.ErrClassNotFound)
		return
	}

	class, err := h.classSvc.Get(id)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			apperrors.Respond(c, appErr)
			return
		}
		apperrors.Respond(c, apperrors.ErrInternal)
		return
	}

	c.JSON(http.StatusOK, class)
}

func (h *ClassHandler) Create(c *gin.Context) {
	var req services.CreateClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.Respond(c, apperrors.ErrValidation)
		return
	}

	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)
	role := c.MustGet(middleware.ContextKeyRole).(models.Role)

	class, err := h.classSvc.Create(req, userID, role)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			apperrors.Respond(c, appErr)
			return
		}
		apperrors.Respond(c, apperrors.ErrInternal)
		return
	}

	c.JSON(http.StatusCreated, class)
}

func (h *ClassHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apperrors.Respond(c, apperrors.ErrClassNotFound)
		return
	}

	var req services.UpdateClassRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.Respond(c, apperrors.ErrValidation)
		return
	}

	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)
	role := c.MustGet(middleware.ContextKeyRole).(models.Role)

	class, err := h.classSvc.Update(id, req, userID, role)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			apperrors.Respond(c, appErr)
			return
		}
		apperrors.Respond(c, apperrors.ErrInternal)
		return
	}

	c.JSON(http.StatusOK, class)
}

func (h *ClassHandler) Cancel(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apperrors.Respond(c, apperrors.ErrClassNotFound)
		return
	}

	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)
	role := c.MustGet(middleware.ContextKeyRole).(models.Role)

	if err := h.classSvc.Cancel(id, userID, role); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			apperrors.Respond(c, appErr)
			return
		}
		apperrors.Respond(c, apperrors.ErrInternal)
		return
	}

	c.Status(http.StatusNoContent)
}
