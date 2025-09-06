// internal/auth/application/service.go - FIXED VERSION
package application

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
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
}

func NewAuthService(authRepo *repository.AuthRepository, kratosClient *kratos.Client, jwtService *jwt.JWTService) *AuthService {
	return &AuthService{
		authRepo:     authRepo,
		kratosClient: kratosClient,
		jwtService:   jwtService,
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
func (s *AuthService) RefreshToken(currentTokenString string, sessionToken string) (*domain.LoginResponse, error) {
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

func (s *AuthService) parseExpiredToken(tokenString string) (*jwt.Claims, error) {
	return s.jwtService.ValidateTokenIgnoringExpiry(tokenString)
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

func (s *AuthService) InitiatePasswordRecovery(email string) (*domain.RecoveryFlow, error) {
	// Create a new recovery flow with Kratos
	url := fmt.Sprintf("%s/self-service/recovery/api", s.kratosClient.GetPublicURL())

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, appError.NewInternalError(err, "failed to create recovery flow request")
	}

	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, appError.NewInternalError(err, "failed to initiate recovery flow")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, appError.NewInternalError(nil, fmt.Sprintf("kratos returned status %d", resp.StatusCode))
	}

	var flow domain.RecoveryFlow
	if err := json.NewDecoder(resp.Body).Decode(&flow); err != nil {
		return nil, appError.NewInternalError(err, "failed to decode recovery flow response")
	}

	// Now submit the email to trigger OTP code sending
	_, err = s.SubmitRecoveryFlow(flow.ID, email, "")
	if err != nil {
		return nil, err
	}

	return &flow, nil
}

func (s *AuthService) SubmitRecoveryFlow(flowID, email, csrfToken string) (*domain.RecoverySubmissionResult, error) {
	url := fmt.Sprintf("%s/self-service/recovery?flow=%s", s.kratosClient.GetPublicURL(), flowID)

	payload := map[string]interface{}{
		"method": "code",
		"email":  email,
	}

	if csrfToken != "" {
		payload["csrf_token"] = csrfToken
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, appError.NewInternalError(err, "failed to marshal recovery submission")
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, appError.NewInternalError(err, "failed to create recovery submission request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, appError.NewInternalError(err, "failed to submit recovery flow")
	}
	defer resp.Body.Close()

	var result domain.RecoverySubmissionResult
	if err := json.NewDecoder(resp.Body).Decode(&result.Flow); err != nil {
		return nil, appError.NewInternalError(err, "failed to decode recovery submission response")
	}

	// Check if the submission was successful
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.Success = true
		result.Message = "Recovery code sent successfully"
		result.Step = "code_sent"
	} else {
		result.Success = false
		result.Message = "Failed to send recovery code"
	}

	return &result, nil
}

func (s *AuthService) SetNewPassword(flowID, password string) (*domain.RecoverySubmissionResult, error) {
	url := fmt.Sprintf("%s/self-service/recovery?flow=%s", s.kratosClient.GetPublicURL(), flowID)

	payload := map[string]interface{}{
		"method":   "code",
		"password": password,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, appError.NewInternalError(err, "failed to marshal password update")
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, appError.NewInternalError(err, "failed to create password update request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, appError.NewInternalError(err, "failed to set new password")
	}
	defer resp.Body.Close()

	var result domain.RecoverySubmissionResult
	if err := json.NewDecoder(resp.Body).Decode(&result.Flow); err != nil {
		return nil, appError.NewInternalError(err, "failed to decode password update response")
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.Success = true
		result.Message = "Password updated successfully"
		result.Step = "password_set"
	} else {
		result.Success = false
		result.Message = "Failed to update password"
	}

	return &result, nil
}

func (s *AuthService) VerifyRecoveryCode(flowID, code string) (*domain.RecoverySubmissionResult, error) {
	url := fmt.Sprintf("%s/self-service/recovery?flow=%s", s.kratosClient.GetPublicURL(), flowID)

	payload := map[string]interface{}{
		"method": "code",
		"code":   code,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, appError.NewInternalError(err, "failed to marshal code verification")
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, appError.NewInternalError(err, "failed to create code verification request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, appError.NewInternalError(err, "failed to verify recovery code")
	}
	defer resp.Body.Close()

	var result domain.RecoverySubmissionResult
	if err := json.NewDecoder(resp.Body).Decode(&result.Flow); err != nil {
		return nil, appError.NewInternalError(err, "failed to decode code verification response")
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.Success = true
		result.Message = "Code verified successfully"
		result.Step = "code_verified"
	} else {
		result.Success = false
		result.Message = "Invalid or expired code"
	}

	return &result, nil
}
