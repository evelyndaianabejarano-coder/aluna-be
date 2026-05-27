package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	apperrors "github.com/evelyndaianabejarano-coder/aluna-be/internal/errors"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/models"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/rs/zerolog/log"
)

type ResetService interface {
	ForgotPassword(email string) error
	ResetPassword(token, newPassword string) error
}

type resetService struct {
	userRepo      repository.UserRepository
	resetTokenRepo repository.PasswordResetTokenRepository
}

func NewResetService(userRepo repository.UserRepository, resetTokenRepo repository.PasswordResetTokenRepository) ResetService {
	return &resetService{
		userRepo:      userRepo,
		resetTokenRepo: resetTokenRepo,
	}
}

func (s *resetService) ForgotPassword(email string) error {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// No revelamos si el email existe o no
			return nil
		}
		return apperrors.ErrInternal
	}

	rawToken, err := generateRandomToken()
	if err != nil {
		return apperrors.ErrInternal
	}
	tokenHash := hashToken(rawToken)

	prt := &models.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     tokenHash,
		ExpiresAt: time.Now().Add(time.Hour),
	}
	if err := s.resetTokenRepo.Create(prt); err != nil {
		return apperrors.ErrInternal
	}

	// Fase 7 enviará el email — por ahora lo logueamos
	log.Info().Str("reset_token", rawToken).Str("user_id", user.ID.String()).Msg("token de reset generado")

	return nil
}

func (s *resetService) ResetPassword(token, newPassword string) error {
	tokenHash := hashToken(token)

	prt, err := s.resetTokenRepo.FindByToken(tokenHash)
	if err != nil {
		return apperrors.ErrInvalidToken
	}

	if prt.UsedAt != nil {
		return apperrors.ErrInvalidToken
	}
	if time.Now().After(prt.ExpiresAt) {
		return apperrors.ErrTokenExpired
	}

	user, err := s.userRepo.FindByID(prt.UserID)
	if err != nil {
		return apperrors.ErrUserNotFound
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return apperrors.ErrInternal
	}
	user.PasswordHash = string(hash)
	if err := s.userRepo.Update(user); err != nil {
		return apperrors.ErrInternal
	}

	now := time.Now()
	prt.UsedAt = &now
	if err := s.resetTokenRepo.MarkUsed(prt); err != nil {
		return apperrors.ErrInternal
	}

	return nil
}

func generateRandomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
