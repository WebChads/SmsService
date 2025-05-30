package smsgen

import (
	"crypto/rand"
	"math/big"
)

func GenerateMockSmsCode() (string, error) {
	return "1234", nil
}

func GenerateSmsCode() (string, error) {
	const digits = "0123456789"
	const length = 4

	result := make([]byte, length)

	for i := range length {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		result[i] = digits[num.Int64()]
	}

	return string(result), nil
}
