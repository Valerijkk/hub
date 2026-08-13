package apigateway_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	cryptopkg "github.com/valerii/hub/backend/internal/platform/crypto"
)

func apigatewayTestCipher(t *testing.T) *cryptopkg.Cipher {
	t.Helper()
	c, err := cryptopkg.NewCipherFromKey(cryptopkg.DeriveDevKey())
	require.NoError(t, err)
	return c
}
