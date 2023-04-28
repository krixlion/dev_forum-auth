package tokensmocks

import (
	"github.com/krixlion/dev_forum-auth/pkg/entity"
	"github.com/krixlion/dev_forum-auth/pkg/tokens"
	"github.com/stretchr/testify/mock"
)

type TokenManager struct {
	*mock.Mock
}

func NewTokenManager() TokenManager {
	return TokenManager{
		Mock: new(mock.Mock),
	}
}

func (m TokenManager) Encode(privateKey entity.Key, token entity.Token) ([]byte, error) {
	args := m.Called(privateKey, token)
	return args.Get(0).([]byte), args.Error(1)
}

func (m TokenManager) GenerateOpaque(typ tokens.OpaqueTokenPrefix) (opaqueAccessToken string, seed string, err error) {
	args := m.Called(typ)
	return args.Get(0).(string), args.Get(1).(string), args.Error(2)
}

func (m TokenManager) DecodeOpaque(typ tokens.OpaqueTokenPrefix, encodedOpaqueToken string) (string, error) {
	args := m.Called(typ, encodedOpaqueToken)
	return args.Get(0).(string), args.Error(1)
}
