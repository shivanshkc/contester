package utils

import (
	"math"
)

// GetSmallestMajority provides the count of smallest majority for the given member count.
func GetSmallestMajority(memberCount int) int {
	return int(math.Ceil(float64(memberCount+1) / 2))
}
