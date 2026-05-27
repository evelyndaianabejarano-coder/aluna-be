package services_test

import (
	"testing"

	apperrors "github.com/evelyndaianabejarano-coder/aluna-be/internal/errors"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/models"
	repomocks "github.com/evelyndaianabejarano-coder/aluna-be/internal/repository/mocks"
	"github.com/evelyndaianabejarano-coder/aluna-be/internal/services"
	svcmocks "github.com/evelyndaianabejarano-coder/aluna-be/internal/services/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func newAuthService(repo *repomocks.UserRepository, tokenSvc *svcmocks.TokenService) services.AuthService {
	return services.NewAuthService(repo, tokenSvc)
}

func TestRegister_EmailDuplicado(t *testing.T) {
	repo := &repomocks.UserRepository{}
	tokenSvc := &svcmocks.TokenService{}

	existingUser := &models.User{ID: uuid.New(), Email: "test@mail.com"}
	repo.On("FindByEmail", "test@mail.com").Return(existingUser, nil)

	svc := newAuthService(repo, tokenSvc)
	_, err := svc.Register("test@mail.com", "password123", "Ana")

	assert.ErrorIs(t, err, apperrors.ErrEmailAlreadyExists)
	repo.AssertExpectations(t)
}

func TestRegister_Exitoso(t *testing.T) {
	repo := &repomocks.UserRepository{}
	tokenSvc := &svcmocks.TokenService{}

	repo.On("FindByEmail", "nueva@mail.com").Return(nil, gorm.ErrRecordNotFound)
	repo.On("Create", mock.AnythingOfType("*models.User")).Return(nil)
	tokenSvc.On("GenerateAccessToken", mock.Anything, models.RoleAlumno).Return("access-token", nil)
	tokenSvc.On("GenerateRefreshToken", mock.Anything).Return("refresh-token", nil)

	svc := newAuthService(repo, tokenSvc)
	resp, err := svc.Register("nueva@mail.com", "password123", "Ana")

	assert.NoError(t, err)
	assert.Equal(t, "nueva@mail.com", resp.User.Email)
	assert.Equal(t, "access-token", resp.AccessToken)
	assert.Equal(t, "refresh-token", resp.RefreshToken)

	// Verifica que la contraseña fue hasheada con bcrypt
	createdUser := repo.Calls[1].Arguments.Get(0).(*models.User)
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(createdUser.PasswordHash), []byte("password123")))

	repo.AssertExpectations(t)
	tokenSvc.AssertExpectations(t)
}

func TestLogin_UsuarioNoExiste(t *testing.T) {
	repo := &repomocks.UserRepository{}
	tokenSvc := &svcmocks.TokenService{}

	repo.On("FindByEmail", "noexiste@mail.com").Return(nil, gorm.ErrRecordNotFound)

	svc := newAuthService(repo, tokenSvc)
	_, err := svc.Login("noexiste@mail.com", "pass")

	assert.ErrorIs(t, err, apperrors.ErrInvalidCredentials)
}

func TestLogin_PasswordIncorrecta(t *testing.T) {
	repo := &repomocks.UserRepository{}
	tokenSvc := &svcmocks.TokenService{}

	hash, _ := bcrypt.GenerateFromPassword([]byte("correcta"), bcrypt.DefaultCost)
	user := &models.User{ID: uuid.New(), Email: "u@mail.com", PasswordHash: string(hash)}
	repo.On("FindByEmail", "u@mail.com").Return(user, nil)

	svc := newAuthService(repo, tokenSvc)
	_, err := svc.Login("u@mail.com", "incorrecta")

	assert.ErrorIs(t, err, apperrors.ErrInvalidCredentials)
}

func TestLogin_Exitoso(t *testing.T) {
	repo := &repomocks.UserRepository{}
	tokenSvc := &svcmocks.TokenService{}

	hash, _ := bcrypt.GenerateFromPassword([]byte("correcta"), bcrypt.DefaultCost)
	user := &models.User{ID: uuid.New(), Email: "u@mail.com", PasswordHash: string(hash), Role: models.RoleAlumno}
	repo.On("FindByEmail", "u@mail.com").Return(user, nil)
	tokenSvc.On("GenerateAccessToken", user.ID, models.RoleAlumno).Return("access-token", nil)
	tokenSvc.On("GenerateRefreshToken", user.ID).Return("refresh-token", nil)

	svc := newAuthService(repo, tokenSvc)
	resp, err := svc.Login("u@mail.com", "correcta")

	assert.NoError(t, err)
	assert.Equal(t, "access-token", resp.AccessToken)
	repo.AssertExpectations(t)
	tokenSvc.AssertExpectations(t)
}

func TestRefreshToken_TokenInvalido(t *testing.T) {
	repo := &repomocks.UserRepository{}
	tokenSvc := &svcmocks.TokenService{}

	tokenSvc.On("ValidateRefreshToken", "bad-token").Return(nil, assert.AnError)

	svc := newAuthService(repo, tokenSvc)
	_, err := svc.RefreshToken("bad-token")

	assert.ErrorIs(t, err, apperrors.ErrInvalidRefreshToken)
}

func TestRefreshToken_RotacionExitosa(t *testing.T) {
	repo := &repomocks.UserRepository{}
	tokenSvc := &svcmocks.TokenService{}

	userID := uuid.New()
	claims := &services.Claims{UserID: userID}
	user := &models.User{ID: userID, Role: models.RoleAlumno}

	tokenSvc.On("ValidateRefreshToken", "old-refresh").Return(claims, nil)
	tokenSvc.On("RevokeRefreshToken", "old-refresh").Return(nil)
	repo.On("FindByID", userID).Return(user, nil)
	tokenSvc.On("GenerateAccessToken", userID, models.RoleAlumno).Return("new-access", nil)
	tokenSvc.On("GenerateRefreshToken", userID).Return("new-refresh", nil)

	svc := newAuthService(repo, tokenSvc)
	tokens, err := svc.RefreshToken("old-refresh")

	assert.NoError(t, err)
	assert.Equal(t, "new-access", tokens.AccessToken)
	assert.Equal(t, "new-refresh", tokens.RefreshToken)

	// Verifica que revocó el token viejo antes de emitir el nuevo
	calls := tokenSvc.Calls
	revokeIdx, generateIdx := -1, -1
	for i, c := range calls {
		if c.Method == "RevokeRefreshToken" {
			revokeIdx = i
		}
		if c.Method == "GenerateAccessToken" {
			generateIdx = i
		}
	}
	assert.Less(t, revokeIdx, generateIdx, "RevokeRefreshToken debe ejecutarse antes de GenerateAccessToken")

	tokenSvc.AssertExpectations(t)
	repo.AssertExpectations(t)
}

func TestLogout_Exitoso(t *testing.T) {
	repo := &repomocks.UserRepository{}
	tokenSvc := &svcmocks.TokenService{}

	tokenSvc.On("RevokeRefreshToken", "valid-refresh").Return(nil)

	svc := newAuthService(repo, tokenSvc)
	err := svc.Logout("valid-refresh")

	assert.NoError(t, err)
	tokenSvc.AssertExpectations(t)
}
