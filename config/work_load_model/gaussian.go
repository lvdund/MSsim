package work_load_model

import "math"

type GaussianGenerator struct {
	rand *RandGenerator
}

func NewGaussianGenerator(seed int64) *GaussianGenerator {
	r := NewRandGenerator(seed)
	return &GaussianGenerator{r}
}

// Gaussian returns a random number of gaussian distribution Gauss(mean, stddev^2)
func (grng GaussianGenerator) Gaussian(mean, stddev float64) float64 {
	return mean + stddev*grng.gaussian()
}

func (grng GaussianGenerator) gaussian() float64 {
	// Box-Muller Transform
	var r, x, y float64
	for r >= 1 || r == 0 {
		x = grng.rand.Float64Range(-1.0, 1.0)
		y = grng.rand.Float64Range(-1.0, 1.0)
		r = x*x + y*y
	}
	return x * math.Sqrt(-2*math.Log(r)/r)
}
