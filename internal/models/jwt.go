package models

import "github.com/dgrijalva/jwt-go"

type JWTClaims struct {
	UserID int `json:"user_id"`
	jwt.StandardClaims
}

type contextKey string

const UserIDKey contextKey = "user_id"