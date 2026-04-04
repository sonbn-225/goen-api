package utils

import (
	"math/big"
)

// IsValidDecimal returns true if the string is a valid decimal representation.
func IsValidDecimal(s string) bool {
	_, ok := new(big.Rat).SetString(s)
	return ok
}
