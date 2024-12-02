package work_load_model

import (
	"fmt"
	"math"
)

type PoissonGenerator struct {
	rand *RandGenerator
}

func NewPoissonGenerator(seed int64) *PoissonGenerator {
	r := NewRandGenerator(seed)
	return &PoissonGenerator{r}
}

// Poisson returns a random number of possion distribution
func (prng PoissonGenerator) Poisson(lambda float64) int64 {
	if !(lambda > 0.0) {
		panic(fmt.Sprintf("Invalid lambda: %.2f", lambda))
	}
	return prng.poisson(lambda)
}

func (prng PoissonGenerator) poisson(lambda float64) int64 {
	// algorithm given by Knuth
	L := math.Pow(math.E, -lambda)
	var k int64 = 0
	var p float64 = 1.0

	for p > L {
		k++
		p *= prng.rand.Float64()
	}
	return (k - 1)
}
