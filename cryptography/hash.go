package cryptography

import (
	"log"

	"github.com/matthewhartstonge/argon2"
)

func HashString(data string, cfg *argon2.Config) []byte {
	defaultCfg := argon2.DefaultConfig()

	if cfg == nil {
		cfg = &defaultCfg
	}

	raw, err := cfg.Hash([]byte(data), nil)
	if err != nil {
		log.Fatalln("Error while hashing data")
		panic(err)
	}

	return raw.Encode()
}

func VerifyData(hash string, password string) bool {
	raw, err := argon2.Decode([]byte(hash))
	if err != nil {
		log.Fatalln("Error while decoding hash")
		panic(err)
	}
	ok, err := raw.Verify([]byte(password))
	if err != nil {
		log.Fatalln("Error while verifying data")
		panic(err)
	}

	return ok
}
