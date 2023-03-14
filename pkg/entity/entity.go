package entity

import (
	"time"
)

type Token struct {
	Id        string
	UserId    string
	Type      TokenType
	ExpiresAt time.Time
	IssuedAt  time.Time
}

type TokenType string

const RefreshToken TokenType = "refresh-token"
const AccessToken TokenType = "access-token"
