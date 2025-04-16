package utils

import "sync"

type Increment struct {
	mu      *sync.Mutex
	counter int
}

func (i *Increment) Increase() int {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.counter++

	return i.counter
}

func NewIncrement() *Increment {
	return &Increment{
		mu:      new(sync.Mutex),
		counter: 0,
	}
}
