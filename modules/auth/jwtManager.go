package auth

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"time"
)

// UserClaims is a custom JWT claims that contains some user's information
type UserClaims struct {
	jwt.StandardClaims
	ID   string `json:"id"`
	Role string `json:"role"`
}

type User interface {
	ID() string
	Role() string
}

type JWTManager interface {
	Generate(user User, timeout time.Duration) (string, error)
	Verify(accessToken string) (*UserClaims, error)
}

// JWTManager is a JSON web token manager
type jwtManager struct {
	secretKey string
}

// NewJWTManager returns a new JWT manager
func NewJWTManager(secretKey string) JWTManager {
	return &jwtManager{secretKey}
}

// Generate generates and signs a new token for a user
func (manager *jwtManager) Generate(user User, timeout time.Duration) (string, error) {
	claims := UserClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(timeout).Unix(),
		},
		ID:   user.ID(),
		Role: user.Role(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(manager.secretKey))

}

// Verify verifies the access token string and return a user claim if the token is valid
func (manager *jwtManager) Verify(accessToken string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(
		accessToken,
		&UserClaims{},
		func(token *jwt.Token) (interface{}, error) {
			_, ok := token.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return nil, fmt.Errorf("unexpected token signing method")
			}
			return []byte(manager.secretKey), nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}
