// internal/auth/application/service.go - FIXED VERSION
package application

import (
	"fmt"
	"s29-be/internal/auth/adapters/repository"
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
}

type LoginResponse struct {
	AccessToken string   `json:"access_token"`
	TokenType   string   `json:"token_type"`
	ExpiresIn   int      `json:"expires_in"`
	User        UserInfo `json:"user"`
}

type UserInfo struct {
	ID               uuid.UUID `json:"id"`
	KratosIdentityID string    `json:"kratos_identity_id"`
	Email            string    `json:"email"`
	DisplayName      string    `json:"display_name"`
	UserType         string    `json:"user_type"`
	IsActive         bool      `json:"is_active"`
}

func NewAuthService(authRepo *repository.AuthRepository, kratosClient *kratos.Client, jwtService *jwt.JWTService) *AuthService {
	return &AuthService{
		authRepo:     authRepo,
		kratosClient: kratosClient,
		jwtService:   jwtService,
	}
}

// VerifySessionAndIssueJWT validates Kratos session and issues Audora JWT
func (s *AuthService) VerifySessionAndIssueJWT(sessionToken string) (*LoginResponse, error) {
	// Step 1: Verify session with Kratos
	session, err := s.kratosClient.VerifySession(sessionToken)
	if err != nil {
		if kratosErr, ok := err.(*kratos.KratosError); ok {
			return nil, appError.NewUnauthorizedError(err, kratosErr.Message)
		}
		return nil, appError.NewInternalError(err, "failed to verify session with Kratos")
	}

	// Step 2: Find user in our database using Kratos identity ID
	kratosIdentityID, err := uuid.Parse(session.Identity.ID)
	if err != nil {
		return nil, appError.NewBadRequestError(err, "invalid kratos identity ID")
	}

	user, err := s.authRepo.FindUserByKratosIdentityID(kratosIdentityID)
	if err != nil {
		return nil, appError.NewNotFoundError(err, "user not found in Audora database")
	}

	// Step 3: Validate user is active
	if !user.IsActive {
		return nil, appError.NewForbiddenError(nil, "user account is deactivated")
	}

	// Step 4: Update last login time
	now := time.Now()
	user.LastLoginAt = &now
	if err := s.authRepo.UpdateUserLastLogin(user); err != nil {
		// Log error but don't fail the login
		fmt.Printf("Failed to update last login time: %v\n", err)
	}

	// Step 5: Generate Audora JWT
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

	// Step 6: Return response
	return &LoginResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(tokenLifetime.Seconds()),
		User: UserInfo{
			ID:               user.ID,
			KratosIdentityID: user.KratosIdentityID.String(),
			Email:            user.Email,
			IsActive:         user.IsActive,
		},
	}, nil
}

// ValidateJWT validates an Audora JWT token and checks user status
func (s *AuthService) ValidateJWT(tokenString string) (*jwt.Claims, error) {
	claims, err := s.jwtService.ValidateToken(tokenString)
	if err != nil {
		return nil, appError.NewUnauthorizedError(err, "invalid or expired token")
	}

	// CRITICAL: Always check if user is still active in database
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
func (s *AuthService) RefreshToken(currentTokenString string, sessionToken string) (*LoginResponse, error) {
	// Step 1: Validate the current JWT to get user info (but allow expired tokens for refresh)
	claims, err := s.jwtService.ValidateToken(currentTokenString)
	if err != nil {
		// Allow parsing of expired tokens for refresh, but still validate signature
		claims, err = s.parseExpiredToken(currentTokenString)
		if err != nil {
			return nil, appError.NewUnauthorizedError(err, "invalid token for refresh")
		}
	}

	// Step 2: CRITICAL - Re-validate with Kratos session
	// This ensures the Kratos session is still valid and not revoked
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

	// Step 3: Ensure the Kratos identity matches the JWT claims
	if session.Identity.ID != claims.KratosIdentityID {
		return nil, appError.NewUnauthorizedError(nil, "session identity mismatch")
	}

	// Step 4: Re-fetch user from database to get latest data
	kratosIdentityID, err := uuid.Parse(session.Identity.ID)
	if err != nil {
		return nil, appError.NewBadRequestError(err, "invalid kratos identity ID")
	}

	user, err := s.authRepo.FindUserByKratosIdentityID(kratosIdentityID)
	if err != nil {
		return nil, appError.NewNotFoundError(err, "user not found in Audora database")
	}

	// Step 5: Validate user is still active
	if !user.IsActive {
		return nil, appError.NewForbiddenError(nil, "user account is deactivated")
	}

	// Step 6: Generate new JWT with FRESH user data
	tokenLifetime := 24 * time.Hour
	newToken, err := s.jwtService.GenerateToken(
		user.ID,
		user.KratosIdentityID.String(),
		user.Email,    // Fresh from DB
		user.IsActive, // Fresh from DB
	)
	if err != nil {
		return nil, appError.NewInternalError(err, "failed to generate new token")
	}

	return &LoginResponse{
		AccessToken: newToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(tokenLifetime.Seconds()),
		User: UserInfo{
			ID:               user.ID,
			KratosIdentityID: user.KratosIdentityID.String(),
			Email:            user.Email,
			IsActive:         user.IsActive,
		},
	}, nil
}

// Helper method to parse expired tokens (for refresh only)
func (s *AuthService) parseExpiredToken(tokenString string) (*jwt.Claims, error) {
	// This should parse the token without validating expiration
	// Implementation depends on your JWT library
	return s.jwtService.ValidateTokenIgnoringExpiry(tokenString)
}

// GetCurrentUser returns user info from JWT claims
func (s *AuthService) GetCurrentUser(claims *jwt.Claims) (*UserInfo, error) {
	return &UserInfo{
		ID:               claims.UserID,
		KratosIdentityID: claims.KratosIdentityID,
		Email:            claims.Email,
		DisplayName:      claims.DisplayName,
		UserType:         claims.UserType,
		IsActive:         claims.IsActive,
	}, nil
}
