package entity

import (
	"time"
)

const Issuer = "http://auth-service"

type Token struct {
	Id        string
	UserId    string
	Type      TokenType
	Issuer    string
	ExpiresAt time.Time
	IssuedAt  time.Time
}

type TokenType string

const RefreshToken TokenType = "refresh-token"
const AccessToken TokenType = "access-token"
