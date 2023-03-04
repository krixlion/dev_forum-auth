package entity

import "time"

type Token struct {
	Id      string
	Type    TokenType
	Expires time.Time
}

type TokenType string

const RefreshToken TokenType = "refresh-token"
const AccessToken TokenType = "access-token"
