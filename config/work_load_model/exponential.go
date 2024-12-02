package work_load_model

import (
	"fmt"
	"math"
)

type ExpGenerator struct {
	rand *RandGenerator
}

func NewExpGenerator(seed int64) *ExpGenerator {
	r := NewRandGenerator(seed)
	return &ExpGenerator{r}
}

// Exp returns a random number of exponential distribution
func (erng ExpGenerator) Exp(lambda float64) float64 {
	if !(lambda > 0.0) {
		panic(fmt.Sprintf("Invalid lambda: %.2f", lambda))
	}
	return erng.exp(lambda)
}

func (erng ExpGenerator) exp(lambda float64) float64 {
	return -math.Log(1-erng.rand.Float64()) / lambda
}
