package tokensmocks

import (
	"github.com/krixlion/dev_forum-auth/pkg/tokens"
	"github.com/stretchr/testify/mock"
)

var _ tokens.Validator = (*TokenValidator)(nil)

type TokenValidator struct {
	*mock.Mock
}

func NewTokenValidator() TokenValidator {
	return TokenValidator{
		Mock: new(mock.Mock),
	}
}

func (m TokenValidator) ValidateToken(token string) error {
	args := m.Called(token)
	return args.Error(0)
}
