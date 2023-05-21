package serialize_test

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/krixlion/dev_forum-auth/pkg/grpc/deserialize"
	"github.com/krixlion/dev_forum-auth/pkg/storage/vault"
)

func TestKeySerializationFlowCompatibilityForRSA(t *testing.T) {
	original, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Errorf("Failed to generate a RSA key: error = %v", err)
		return
	}

	serialized, err := vault.EncodeRSA(original)
	if err != nil {
		t.Errorf("Failed to encode RSA key: error = %v", err)
		return
	}

	pubKey, err := deserialize.Key(serialized)
	if err != nil {
		t.Errorf("Failed to serialize RSA key: error = %v", err)
		return
	}

	want := original.PublicKey
	got, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		t.Errorf("Received key is of invalid type: got = %T, want = %T", pubKey, got)
		return
	}

	if !want.Equal(got) {
		t.Errorf("Public Keys are not equal: %v", cmp.Diff(want, got))
	}
}
