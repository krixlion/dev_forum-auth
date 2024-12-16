package tokensmocks

import (
	"github.com/krixlion/dev_forum-auth/pkg/entity"
	"github.com/krixlion/dev_forum-auth/pkg/tokens"
	"github.com/stretchr/testify/mock"
)

var _ tokens.Parser = (*TokenParser)(nil)

type TokenParser struct {
	*mock.Mock
}

func NewTokenParser() TokenParser {
	return TokenParser{
		Mock: new(mock.Mock),
	}
}

func (m TokenParser) ParseToken(token string) (entity.Token, error) {
	args := m.Called(token)
	return args.Get(0).(entity.Token), args.Error(1)
}
