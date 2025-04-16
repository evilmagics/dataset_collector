package utils

import (
	"testing"

	"github.com/panjf2000/ants/v2"
)

func TestIncrements(t *testing.T) {
	incr := NewIncrement()
	pool, err := ants.NewPool(500, ants.WithPreAlloc(true))
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < 100000; i++ {
		pool.Submit(func() {
			incr.Increase()
		})
	}
}

func BenchmarkIncrement(b *testing.B) {
	incr := NewIncrement()
	b.Run("increment", func(b *testing.B) {
		incr.Increase()
	})
}
