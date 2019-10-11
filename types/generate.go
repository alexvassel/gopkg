package types

import (
	"crypto/rand"
	uuid "gopkg.in/satori/go.uuid.v1"
	"math/big"
)

func GeneratePassword(length int) (string, error) {
	result := ""
	for {
		if len(result) >= length {
			return result, nil
		}
		num, err := rand.Int(rand.Reader, big.NewInt(int64(127)))
		if err != nil {
			return "", err
		}
		n := num.Int64()
		if n > 32 && n < 127 {
			result += string(n)
		}
	}
}

func GenerateUUID() uuid.UUID {
	return uuid.NewV4()
}
