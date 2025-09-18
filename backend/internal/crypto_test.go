package internal

import (
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestCM() *CryptoManager {
	return NewCryptoManager(Configuration{
		EncryptionPassphrase: "test-passphrase",
		EncryptionSalt:       "test-salt",
		Argon2:               Argon2Config{Time: 1, Memory: 65536, Threads: 2},
	})
}

type nonEmptyASCII string

func (nonEmptyASCII) Generate(r *rand.Rand, size int) reflect.Value {
	length := r.Intn(64) + 1
	b := make([]byte, length)
	for i := 0; i < length; i++ {

		b[i] = byte(33 + r.Intn(94))
	}
	return reflect.ValueOf(nonEmptyASCII(string(b)))
}

func TestEncryptDecrypt_Property(t *testing.T) {
	cm := newTestCM()

	prop := func(p nonEmptyASCII, a nonEmptyASCII) bool {
		plain := string(p)
		addr := string(a)
		ct, err := cm.Encrypt(plain, addr)
		if !assert.NoError(t, err, "encrypt failed: plain=%q addr=%q", plain, addr) {
			return false
		}
		pt, err := cm.Decrypt(ct, addr)
		if !assert.NoError(t, err, "decrypt failed: addr=%q", addr) {
			return false
		}
		return assert.Equal(t, plain, pt, "roundtrip mismatch")
	}

	cfg := &quick.Config{MaxCount: 200, Rand: rand.New(rand.NewSource(time.Now().UnixNano()))}
	require.NoError(t, quick.Check(prop, cfg))
}
