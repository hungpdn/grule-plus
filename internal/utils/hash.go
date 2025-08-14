package utils

import (
	"crypto/sha256"
	"math/big"
)

// HashStringToRange hashes a string to a cryptographically secure integer
// within the specified range [min, max] with a good standard distribution.
func HashStringToRange(s string, min, max int64) int64 {

	rangeSize := big.NewInt(max - min + 1)
	hash := sha256.Sum256([]byte(s))
	hashInt := new(big.Int).SetBytes(hash[:])
	resultInt := new(big.Int).Mod(hashInt, rangeSize)

	return resultInt.Int64() + min
}
