package entity

import (
	"time"
)

type Token struct {
	Id        string // Token's ID is it's related decoded opaque token.
	UserId    string
	Type      TokenType
	ExpiresAt time.Time
	IssuedAt  time.Time
}

type TokenType string

const (
	RefreshToken TokenType = "refresh-token"
	AccessToken  TokenType = "access-token"
)
