package agent

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/talx-hub/malerter/internal/constants"
)

func Test_trySign_withSecret(t *testing.T) {
	data := []byte("some important content")
	secret := "top-secret"
	signature := trySign(data, secret)

	assert.NotEmpty(t, signature)
}

func Test_trySign_noSecret(t *testing.T) {
	data := []byte("some content")
	signature := trySign(data, constants.NoSecret)

	assert.Equal(t, "", signature)
}

func Test_tryEncrypt_nilEncrypter(t *testing.T) {
	data := []byte("unencrypted")
	out, err := tryEncrypt(data, nil)

	assert.NoError(t, err)
	assert.Equal(t, data, out)
}
