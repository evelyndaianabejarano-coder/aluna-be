package handlers

import (
	"net/http"
	"strconv"

	apperrors "github.com/evelyndaianabejarano-coder/aluna-be/internal/errors"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/middleware"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/models"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ReservationHandler struct {
	reservSvc  services.ReservationService
	waitSvc    services.WaitlistService
}

func NewReservationHandler(reservSvc services.ReservationService, waitSvc services.WaitlistService) *ReservationHandler {
	return &ReservationHandler{reservSvc: reservSvc, waitSvc: waitSvc}
}

func (h *ReservationHandler) Reserve(c *gin.Context) {
	var req struct {
		ClaseID string `json:"clase_id" binding:"required,uuid"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.Respond(c, apperrors.ErrValidation)
		return
	}

	claseID, _ := uuid.Parse(req.ClaseID)
	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)

	reservation, err := h.reservSvc.Reserve(userID, claseID)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			apperrors.Respond(c, appErr)
			return
		}
		apperrors.Respond(c, apperrors.ErrInternal)
		return
	}

	c.JSON(http.StatusCreated, reservation)
}

func (h *ReservationHandler) Cancel(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apperrors.Respond(c, apperrors.ErrReservationNotFound)
		return
	}

	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)
	role := c.MustGet(middleware.ContextKeyRole).(models.Role)

	if err := h.reservSvc.Cancel(id, userID, role); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			apperrors.Respond(c, appErr)
			return
		}
		apperrors.Respond(c, apperrors.ErrInternal)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ReservationHandler) GetMy(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}

	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)

	reservations, total, err := h.reservSvc.GetMy(userID, page, limit)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			apperrors.Respond(c, appErr)
			return
		}
		apperrors.Respond(c, apperrors.ErrInternal)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  reservations,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *ReservationHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		apperrors.Respond(c, apperrors.ErrReservationNotFound)
		return
	}

	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)
	role := c.MustGet(middleware.ContextKeyRole).(models.Role)

	reservation, err := h.reservSvc.Get(id, userID, role)
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			apperrors.Respond(c, appErr)
			return
		}
		apperrors.Respond(c, apperrors.ErrInternal)
		return
	}

	c.JSON(http.StatusOK, reservation)
}

func (h *ReservationHandler) JoinWaitlist(c *gin.Context) {
	var req struct {
		ClaseID string `json:"clase_id" binding:"required,uuid"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apperrors.Respond(c, apperrors.ErrValidation)
		return
	}

	claseID, _ := uuid.Parse(req.ClaseID)
	userID := c.MustGet(middleware.ContextKeyUserID).(uuid.UUID)

	if err := h.waitSvc.Join(userID, claseID); err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			apperrors.Respond(c, appErr)
			return
		}
		apperrors.Respond(c, apperrors.ErrInternal)
		return
	}

	c.Status(http.StatusCreated)
}
