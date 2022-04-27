package cryptography

import (
	"crypto/rand"

	"golang.org/x/crypto/argon2"
)

type Options struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

var (
	defaultOpts = Options{
		Memory:      64 * 1024,
		Iterations:  3,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}
)

func HashString(data string, opts *Options) (hash []byte, err error) {
	if opts == nil {
		opts = &defaultOpts
	}
	// Generate a cryptographically secure random salt.
	salt, err := generateRandomBytes(opts.SaltLength)
	if err != nil {
		return nil, err
	}

	// Pass the plaintext data, salt and parameters to the argon2.IDKey
	// function. This will generate a hash of the data using the Argon2id
	// variant.
	hash = argon2.IDKey([]byte(data), salt, opts.Iterations, opts.Memory, opts.Parallelism, opts.KeyLength)

	return hash, nil
}

func generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}
