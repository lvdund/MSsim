package work_load_model

import (
	"fmt"
	"math/rand"
	"sync"
)

type RandGenerator struct {
	mu *sync.Mutex
	rd *rand.Rand
}

func NewRandGenerator(seed int64) *RandGenerator {
	return &RandGenerator{mu: new(sync.Mutex), rd: rand.New(rand.NewSource(seed))}
}

// Float64 returns a random float64 in [0.0, 1.0)
func (ung RandGenerator) Float64() float64 {
	ung.mu.Lock()
	defer ung.mu.Unlock()
	return ung.rd.Float64()
}

// Float32Range returns a random float32 in [a, b)
func (ung RandGenerator) Float64Range(a, b float64) float64 {
	if !(a < b) {
		panic(fmt.Sprintf("Invalid range: %.2f ~ %.2f", a, b))
	}
	return a + ung.Float64()*(b-a)
}

func SortIncrease(l []int64) {
	i := 0
	for i < len(l) {
		var j = i
		for j >= 1 && l[j] < l[j-1] {
			l[j], l[j-1] = l[j-1], l[j]
			j--
		}
		i++
	}
}
