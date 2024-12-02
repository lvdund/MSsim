package work_load_model

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestPoissonGenerator(t *testing.T) {
	grng := NewPoissonGenerator(time.Now().UnixNano())
	hist := map[int64]int{}
	for i := 0; i < 10000; i++ {
		x := grng.Poisson(20.0)
		hist[int64(x)] += 1
	}

	keys := []int64{}
	for k := range hist {
		keys = append(keys, k)
	}
	SortIncrease(keys)

	for _, key := range keys {
		fmt.Printf("%d:\t%s\n", key, strings.Repeat("*", hist[key]/100))
	}
}
