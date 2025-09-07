// internal/auth/application/service.go - FIXED VERSION
package application

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"s29-be/internal/auth/adapters/repository"
	"s29-be/internal/auth/domain"
	model "s29-be/internal/user/domain"
	appError "s29-be/pkg/error"
	"s29-be/pkg/jwt"
	"s29-be/pkg/kratos"
	"time"

	"github.com/google/uuid"
)

type AuthService struct {
	authRepo     *repository.AuthRepository
	kratosClient *kratos.Client
	jwtService   *jwt.JWTService
	// Temporary storage for recovery codes (in production, use Redis or similar)
	recoveryCodeCache map[string]string // flowID -> code
}

func NewAuthService(authRepo *repository.AuthRepository, kratosClient *kratos.Client, jwtService *jwt.JWTService) *AuthService {
	return &AuthService{
		authRepo:          authRepo,
		kratosClient:      kratosClient,
		jwtService:        jwtService,
		recoveryCodeCache: make(map[string]string),
	}
}

func (s *AuthService) VerifySessionAndIssueJWT(sessionToken string) (*domain.LoginResponse, error) {
	session, err := s.kratosClient.VerifySession(sessionToken)
	if err != nil {
		if kratosErr, ok := err.(*kratos.KratosError); ok {
			return nil, appError.NewUnauthorizedError(err, kratosErr.Message)
		}
		return nil, appError.NewInternalError(err, "failed to verify session with Kratos")
	}

	kratosIdentityID, err := uuid.Parse(session.Identity.ID)
	if err != nil {
		return nil, appError.NewBadRequestError(err, "invalid kratos identity ID")
	}

	user, err := s.authRepo.FindUserByKratosIdentityID(kratosIdentityID)
	if err != nil {
		return nil, appError.NewNotFoundError(err, "user not found in Audora database")
	}

	if !user.IsActive {
		return nil, appError.NewForbiddenError(nil, "user account is deactivated")
	}

	now := time.Now()
	user.LastLoginAt = &now
	if err := s.authRepo.UpdateUserLastLogin(user); err != nil {
		fmt.Printf("Failed to update last login time: %v\n", err)
	}

	tokenLifetime := 24 * time.Hour // 24 hours
	accessToken, err := s.jwtService.GenerateToken(
		user.ID,
		user.KratosIdentityID.String(),
		user.Email,
		user.IsActive,
	)
	if err != nil {
		return nil, appError.NewInternalError(err, "failed to generate access token")
	}

	return &domain.LoginResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(tokenLifetime.Seconds()),
		User: domain.UserInfo{
			ID:               user.ID,
			KratosIdentityID: user.KratosIdentityID.String(),
			Email:            user.Email,
			IsActive:         user.IsActive,
		},
	}, nil
}

func (s *AuthService) ValidateJWT(tokenString string) (*jwt.Claims, error) {
	claims, err := s.jwtService.ValidateToken(tokenString)
	if err != nil {
		return nil, appError.NewUnauthorizedError(err, "invalid or expired token")
	}

	kratosIdentityID, err := uuid.Parse(claims.KratosIdentityID)
	if err != nil {
		return nil, appError.NewUnauthorizedError(err, "invalid identity ID in token")
	}

	user, err := s.authRepo.FindUserByKratosIdentityID(kratosIdentityID)
	if err != nil {
		return nil, appError.NewUnauthorizedError(err, "user not found")
	}

	if !user.IsActive {
		return nil, appError.NewForbiddenError(nil, "user account is deactivated")
	}

	return claims, nil
}

// FIXED: RefreshToken now validates Kratos session before issuing new JWT
func (s *AuthService) RefreshToken(sessionToken string) (*domain.LoginResponse, error) {
	if sessionToken == "" {
		return nil, appError.NewUnauthorizedError(nil, "session token required for refresh")
	}

	session, err := s.kratosClient.VerifySession(sessionToken)
	if err != nil {
		if kratosErr, ok := err.(*kratos.KratosError); ok {
			return nil, appError.NewUnauthorizedError(err, "kratos session invalid: "+kratosErr.Message)
		}
		return nil, appError.NewInternalError(err, "failed to verify session with Kratos during refresh")
	}

	kratosIdentityID, err := uuid.Parse(session.Identity.ID)
	if err != nil {
		return nil, appError.NewBadRequestError(err, "invalid kratos identity ID")
	}

	user, err := s.authRepo.FindUserByKratosIdentityID(kratosIdentityID)
	if err != nil {
		return nil, appError.NewNotFoundError(err, "user not found in Audora database")
	}

	if !user.IsActive {
		return nil, appError.NewForbiddenError(nil, "user account is deactivated")
	}

	tokenLifetime := 24 * time.Hour
	newToken, err := s.jwtService.GenerateToken(
		user.ID,
		user.KratosIdentityID.String(),
		user.Email,
		user.IsActive,
	)
	if err != nil {
		return nil, appError.NewInternalError(err, "failed to generate new token")
	}

	return &domain.LoginResponse{
		AccessToken: newToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(tokenLifetime.Seconds()),
		User: domain.UserInfo{
			ID:               user.ID,
			KratosIdentityID: user.KratosIdentityID.String(),
			Email:            user.Email,
			IsActive:         user.IsActive,
		},
	}, nil
}

func (s *AuthService) GetCurrentUser(claims *jwt.Claims) (*domain.UserInfo, error) {
	return &domain.UserInfo{
		ID:               claims.UserID,
		KratosIdentityID: claims.KratosIdentityID,
		Email:            claims.Email,
		DisplayName:      claims.DisplayName,
		UserType:         claims.UserType,
		IsActive:         claims.IsActive,
	}, nil
}

func (s *AuthService) FindUserByKratosIdentityID(kratosID uuid.UUID) (*model.User, error) {
	return s.authRepo.FindUserByKratosIdentityID(kratosID)
}

func (s *AuthService) UpdateUserLastLogin(user *model.User) error {
	return s.authRepo.UpdateUserLastLogin(user)
}

func (s *AuthService) InitiateOAuthFlow(provider, redirectURI, state string) (string, string, error) {
	if state == "" {
		stateBytes := make([]byte, 32)
		if _, err := rand.Read(stateBytes); err != nil {
			return "", "", appError.NewInternalError(err, "failed to generate state")
		}
		state = base64.URLEncoding.EncodeToString(stateBytes)
	}

	baseURL := "http://localhost:4433"
	authURL := fmt.Sprintf("%s/self-service/methods/oidc/auth/%s?return_to=%s&state=%s",
		baseURL, provider, redirectURI, state)

	return authURL, state, nil
}

func (s *AuthService) HandleOAuthCallback(provider, code, state string) (*domain.LoginResponse, error) {
	sessionToken, err := s.kratosClient.ExchangeCodeForSession(provider, code, state)
	if err != nil {
		return nil, appError.NewUnauthorizedError(err, "failed to exchange OAuth code")
	}

	return s.VerifySessionAndIssueJWT(sessionToken)
}
