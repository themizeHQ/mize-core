package cryptography

import (
	"errors"

	"github.com/matthewhartstonge/argon2"
	"go.uber.org/zap"
	"mize.app/logger"
)

func HashString(data string, cfg *argon2.Config) []byte {
	defaultCfg := argon2.DefaultConfig()

	if cfg == nil {
		cfg = &defaultCfg
	}

	raw, err := cfg.Hash([]byte(data), nil)
	if err != nil {
		logger.Error(errors.New("error while hashing data"), zap.Error(err))
	}

	return raw.Encode()
}

func VerifyData(hash string, password string) bool {
	raw, err := argon2.Decode([]byte(hash))
	if err != nil {
		logger.Error(errors.New("could not decode data - argon"), zap.Error(err))
	}
	ok, err := raw.Verify([]byte(password))
	if err != nil {
		logger.Error(errors.New("error while verifying password - argon"), zap.Error(err))
	}

	return ok
}
