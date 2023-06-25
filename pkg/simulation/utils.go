package simulation

import (
	"math/rand"
	"time"

	"github.com/goombaio/namegenerator"
)

// _randomSource is the common source for random stuff generation.
var _randomSource = rand.NewSource(time.Now().UnixNano())

// biasedBoolean returns a boolean randomly that is as likely to be true as specified.
func biasedBoolean(probabilityOfTrue float64) bool {
	if probabilityOfTrue > 1 || probabilityOfTrue < 0 {
		panic("probability should be between 0 and 1 both inclusive")
	}

	switch probabilityOfTrue {
	case 1:
		return true
	case 0:
		return false
	default:
		return probabilityOfTrue > rand.New(_randomSource).Float64()
	}
}

// randomDurationBetween returns a random time duration in the given range, both inclusive.
func randomDurationBetween(min, max time.Duration) time.Duration {
	// Special case for good performance.
	if max == 0 {
		return 0
	}

	minInt, maxInt := int(min), int(max)
	randomBW := rand.Intn(maxInt-minInt+1) + minInt
	return time.Duration(randomBW)
}

// _nameGen generates random readable values.
var _nameGen = namegenerator.NewNameGenerator(time.Now().UTC().UnixNano())

// getRandomValue generates a random string.
func getRandomValue() string {
	return _nameGen.Generate()
}
