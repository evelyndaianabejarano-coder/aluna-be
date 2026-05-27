package services

import (
	"errors"

	apperrors "github.com/evelyndaianabejarano-coder/aluna-be/internal/errors"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/models"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthTokens struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type AuthResponse struct {
	User         *UserDTO   `json:"user"`
	AccessToken  string     `json:"accessToken"`
	RefreshToken string     `json:"refreshToken"`
}

type UserDTO struct {
	ID     string      `json:"id"`
	Email  string      `json:"email"`
	Nombre string      `json:"nombre"`
	Role   models.Role `json:"role"`
}

type AuthService interface {
	Register(email, password, nombre string) (*AuthResponse, error)
	Login(email, password string) (*AuthResponse, error)
	RefreshToken(refreshToken string) (*AuthTokens, error)
	Logout(refreshToken string) error
}

type authService struct {
	userRepo     repository.UserRepository
	tokenService TokenService
}

func NewAuthService(userRepo repository.UserRepository, tokenService TokenService) AuthService {
	return &authService{
		userRepo:     userRepo,
		tokenService: tokenService,
	}
}

func (s *authService) Register(email, password, nombre string) (*AuthResponse, error) {
	existing, err := s.userRepo.FindByEmail(email)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrInternal
	}
	if existing != nil {
		return nil, apperrors.ErrEmailAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperrors.ErrInternal
	}

	user := &models.User{
		Email:        email,
		PasswordHash: string(hash),
		Nombre:       nombre,
		Role:         models.RoleAlumno,
	}
	if err := s.userRepo.Create(user); err != nil {
		return nil, apperrors.ErrInternal
	}

	return s.buildAuthResponse(user)
}

func (s *authService) Login(email, password string) (*AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.ErrInvalidCredentials
		}
		return nil, apperrors.ErrInternal
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, apperrors.ErrInvalidCredentials
	}

	return s.buildAuthResponse(user)
}

func (s *authService) RefreshToken(refreshToken string) (*AuthTokens, error) {
	claims, err := s.tokenService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, apperrors.ErrInvalidRefreshToken
	}

	// Rotate: revoke old token before issuing new ones
	if err := s.tokenService.RevokeRefreshToken(refreshToken); err != nil {
		return nil, apperrors.ErrInternal
	}

	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, apperrors.ErrUserNotFound
	}

	accessToken, err := s.tokenService.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, apperrors.ErrInternal
	}
	newRefreshToken, err := s.tokenService.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, apperrors.ErrInternal
	}

	return &AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *authService) Logout(refreshToken string) error {
	if err := s.tokenService.RevokeRefreshToken(refreshToken); err != nil {
		return apperrors.ErrInvalidRefreshToken
	}
	return nil
}

func (s *authService) buildAuthResponse(user *models.User) (*AuthResponse, error) {
	accessToken, err := s.tokenService.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, apperrors.ErrInternal
	}
	refreshToken, err := s.tokenService.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, apperrors.ErrInternal
	}

	return &AuthResponse{
		User: &UserDTO{
			ID:     user.ID.String(),
			Email:  user.Email,
			Nombre: user.Nombre,
			Role:   user.Role,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
