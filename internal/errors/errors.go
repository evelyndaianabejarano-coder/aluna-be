package errors

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func (e *AppError) Error() string {
	return e.Message
}

func New(code, message string, status int) *AppError {
	return &AppError{Code: code, Message: message, Status: status}
}

func Respond(c *gin.Context, err *AppError) {
	c.JSON(err.Status, gin.H{"error": err})
}

// Autenticación
var (
	ErrInvalidCredentials    = New("AUTH_INVALID_CREDENTIALS", "Credenciales inválidas", http.StatusUnauthorized)
	ErrTokenExpired          = New("AUTH_TOKEN_EXPIRED", "El token expiró", http.StatusUnauthorized)
	ErrUnauthorized          = New("AUTH_UNAUTHORIZED", "No tenés permiso para realizar esta acción", http.StatusForbidden)
	ErrUserNotFound          = New("AUTH_USER_NOT_FOUND", "Usuario no encontrado", http.StatusNotFound)
	ErrEmailAlreadyExists    = New("AUTH_EMAIL_ALREADY_EXISTS", "El email ya está registrado", http.StatusConflict)
	ErrInvalidToken          = New("AUTH_INVALID_TOKEN", "Token inválido", http.StatusUnauthorized)
	ErrInvalidRefreshToken   = New("AUTH_INVALID_REFRESH_TOKEN", "Refresh token inválido o expirado", http.StatusUnauthorized)
)

// Reservas
var (
	ErrReservationClassFull        = New("RESERVATION_CLASS_FULL", "La clase no tiene cupos disponibles", http.StatusConflict)
	ErrReservationConflict         = New("RESERVATION_CONFLICT", "Ya tenés una reserva activa para esta clase", http.StatusConflict)
	ErrReservationNotFound         = New("RESERVATION_NOT_FOUND", "Reserva no encontrada", http.StatusNotFound)
	ErrReservationAlreadyCancelled = New("RESERVATION_ALREADY_CANCELLED", "La reserva ya fue cancelada", http.StatusBadRequest)
)

// Lista de espera
var (
	ErrWaitlistAlreadyJoined = New("WAITLIST_ALREADY_JOINED", "Ya estás en la lista de espera de esta clase", http.StatusConflict)
	ErrWaitlistNotFound      = New("WAITLIST_NOT_FOUND", "No estás en la lista de espera de esta clase", http.StatusNotFound)
	ErrClassNotFull          = New("CLASS_NOT_FULL", "La clase tiene cupos disponibles, podés reservar directamente", http.StatusBadRequest)
)

// Clases
var (
	ErrClassNotFound        = New("CLASS_NOT_FOUND", "Clase no encontrada", http.StatusNotFound)
	ErrClassAlreadyCancelled = New("CLASS_ALREADY_CANCELLED", "La clase ya fue cancelada", http.StatusBadRequest)
	ErrClassInvalidDate     = New("CLASS_INVALID_DATE", "La fecha de la clase no es válida", http.StatusUnprocessableEntity)
)

// Pagos
var (
	ErrPaymentAlreadyConfirmed = New("PAYMENT_ALREADY_CONFIRMED", "El pago ya fue confirmado", http.StatusConflict)
	ErrPaymentNotFound         = New("PAYMENT_NOT_FOUND", "Pago no encontrado", http.StatusNotFound)
)

// Categorías
var (
	ErrCategoryNotFound = New("CATEGORY_NOT_FOUND", "Categoría no encontrada", http.StatusNotFound)
)

// General
var (
	ErrValidation        = New("VALIDATION_ERROR", "Los datos enviados no son válidos", http.StatusUnprocessableEntity)
	ErrInternal          = New("INTERNAL_ERROR", "Error interno del servidor", http.StatusInternalServerError)
	ErrNotFound          = New("NOT_FOUND", "Recurso no encontrado", http.StatusNotFound)
	ErrRateLimitExceeded = New("RATE_LIMIT_EXCEEDED", "Demasiados intentos. Intentá de nuevo más tarde", http.StatusTooManyRequests)
)
