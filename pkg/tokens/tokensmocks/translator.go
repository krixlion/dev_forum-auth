package tokensmocks

import (
	"github.com/krixlion/dev_forum-auth/pkg/tokens"
	"github.com/krixlion/dev_forum-auth/pkg/tokens/translator"
	"github.com/stretchr/testify/mock"
)

var _ tokens.Translator = (*TokenTranslator)(nil)

type TokenTranslator struct {
	*mock.Mock
}

func NewTokenTranslator() TokenTranslator {
	return TokenTranslator{
		Mock: new(mock.Mock),
	}
}

func (m TokenTranslator) TranslateAccessToken(opaqueAccessToken string) (string, error) {
	args := m.Called(opaqueAccessToken)
	return args.String(0), args.Error(1)
}
