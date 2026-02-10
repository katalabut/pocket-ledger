package cryptoutil

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
)

func SHA256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// RandomDigits returns n digits as a string, zero-padded.
func RandomDigits(n int) (string, error) {
	if n <= 0 {
		return "", fmt.Errorf("n must be > 0")
	}
	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n)), nil)
	x, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	format := fmt.Sprintf("%%0%dd", n)
	return fmt.Sprintf(format, x.Int64()), nil
}
