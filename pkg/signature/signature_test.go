package signature_test

import (
	"testing"

	"github.com/talx-hub/malerter/pkg/signature"
)

func TestHash(t *testing.T) {
	data := []byte("test data")
	key := "secretkey"

	got := signature.Hash(data, key)

	// Заранее вычисленное правильное значение для этих данных и ключа
	want := "309bb1e52729870eca98b738dfd858eef04fab4e5bb91ec0057cd3e0ac246a89"

	if got != want {
		t.Errorf("Hash() = %q; want %q", got, want)
	}
}
