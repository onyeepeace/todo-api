package middleware

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/joho/godotenv"
	"github.com/onyeepeace/todo-api/internal/models"
)

func init() {
	envErr := godotenv.Load()
	if envErr != nil {
		log.Println("Warning: .env file not found.")
	}
}

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

// Generate JWT token for authentication
func GenerateJWT(userID int) (string, error) {
	claims := &models.JWTClaims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: jwt.TimeFunc().Add(24 * time.Hour).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// Middleware to validate JWT
func ValidateJWT(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}

		tokenStr := authHeader[len("Bearer "):]
		claims := &models.JWTClaims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Add user ID to the request context for downstream handlers
		r = r.WithContext(context.WithValue(r.Context(), "user_id", claims.UserID))
		next(w, r)
	}
}