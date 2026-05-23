package handlers

import (
	"net/http"

	apperrors "github.com/evelyndaianabejarano-coder/aluna-be/internal/errors"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/repository"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"errors"
)

type CategoryHandler struct {
	categoryRepo repository.CategoryRepository
}

func NewCategoryHandler(categoryRepo repository.CategoryRepository) *CategoryHandler {
	return &CategoryHandler{categoryRepo: categoryRepo}
}

type categoryRequest struct {
	Nombre      string `json:"nombre"      binding:"required"`
	Objetivo    string `json:"objetivo"    binding:"required"`
	Descripcion string `json:"descripcion" binding:"required"`
}

func (h *CategoryHandler) List(c *gin.Context) {
	categories, err := h.categoryRepo.FindAll()
	if err != nil {
		apperrors.Respond(c, apperrors.ErrInternal)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": categories})
}

func (h *CategoryHandler) Create(c *gin.Context) {
	var req categoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.Respond(c, apperrors.ErrValidation)
		return
	}

	category := &models.Category{
		Nombre:      req.Nombre,
		Objetivo:    req.Objetivo,
		Descripcion: req.Descripcion,
	}
	if err := h.categoryRepo.Create(category); err != nil {
		apperrors.Respond(c, apperrors.ErrInternal)
		return
	}

	c.JSON(http.StatusCreated, category)
}

func (h *CategoryHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apperrors.Respond(c, apperrors.ErrCategoryNotFound)
		return
	}

	category, err := h.categoryRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apperrors.Respond(c, apperrors.ErrCategoryNotFound)
			return
		}
		apperrors.Respond(c, apperrors.ErrInternal)
		return
	}

	var req categoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.Respond(c, apperrors.ErrValidation)
		return
	}

	category.Nombre = req.Nombre
	category.Objetivo = req.Objetivo
	category.Descripcion = req.Descripcion

	if err := h.categoryRepo.Update(category); err != nil {
		apperrors.Respond(c, apperrors.ErrInternal)
		return
	}

	c.JSON(http.StatusOK, category)
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apperrors.Respond(c, apperrors.ErrCategoryNotFound)
		return
	}

	if err := h.categoryRepo.Delete(id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apperrors.Respond(c, apperrors.ErrCategoryNotFound)
			return
		}
		apperrors.Respond(c, apperrors.ErrInternal)
		return
	}

	c.Status(http.StatusNoContent)
}
