package tokensmocks

import (
	"context"

	"github.com/krixlion/dev_forum-auth/pkg/tokens"
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

func (m TokenTranslator) TranslateAccessToken(ctx context.Context, opaqueAccessToken string) (string, error) {
	args := m.Called(ctx, opaqueAccessToken)
	return args.String(0), args.Error(1)
}
