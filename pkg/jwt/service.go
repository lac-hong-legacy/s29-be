// pkg/jwt/service.go - UPDATED with expired token parsing
package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID           uuid.UUID `json:"user_id"`
	KratosIdentityID string    `json:"kratos_identity_id"`
	Email            string    `json:"email"`
	UserType         string    `json:"user_type"`
	DisplayName      string    `json:"display_name"`
	IsActive         bool      `json:"is_active"`
	jwt.RegisteredClaims
}

type JWTService struct {
	secretKey     []byte
	issuer        string
	tokenLifetime time.Duration
}

func NewJWTService(secretKey, issuer string, tokenLifetime time.Duration) *JWTService {
	return &JWTService{
		secretKey:     []byte(secretKey),
		issuer:        issuer,
		tokenLifetime: tokenLifetime,
	}
}

func (j *JWTService) GenerateToken(userID uuid.UUID, kratosIdentityID, email string, isActive bool) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:           userID,
		KratosIdentityID: kratosIdentityID,
		Email:            email,
		IsActive:         isActive,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   kratosIdentityID,
			Audience:  []string{"s29-api"},
			ExpiresAt: jwt.NewNumericDate(now.Add(j.tokenLifetime)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

func (j *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// NEW: ValidateTokenIgnoringExpiry for refresh token validation
func (j *JWTService) ValidateTokenIgnoringExpiry(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return j.secretKey, nil
	}, jwt.WithoutClaimsValidation()) // This ignores expiry validation

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok {
		// Still validate the signature and basic structure, just not expiry
		return claims, nil
	}

	return nil, errors.New("invalid token structure")
}

// REMOVED: The insecure RefreshToken method that didn't validate with Kratos
