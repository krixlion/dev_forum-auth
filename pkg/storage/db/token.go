package db

import (
	"time"

	"github.com/krixlion/dev_forum-auth/pkg/entity"
)

type tokenDocument struct {
	Id        string    `bson:"_id,omitempty"`
	UserId    string    `bson:"user_id,omitempty"`
	Type      string    `bson:"type,omitempty"`
	ExpiresAt time.Time `bson:"expires_at,omitempty"`
	IssuedAt  time.Time `bson:"issued_at,omitempty"`
}

func makeTokenDocument(token entity.Token) tokenDocument {
	return tokenDocument{
		Id:        token.Id,
		UserId:    token.UserId,
		Type:      string(token.Type),
		ExpiresAt: token.ExpiresAt,
		IssuedAt:  token.IssuedAt,
	}
}

func makeTokenFromDocument(v tokenDocument) entity.Token {
	return entity.Token{
		Id:        v.Id,
		UserId:    v.UserId,
		Type:      entity.TokenType(v.Type),
		ExpiresAt: v.ExpiresAt,
		IssuedAt:  v.IssuedAt,
	}
}
