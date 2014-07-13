package turbo

import (
	"sync"
	"testing"
)

func TestLocker(t *testing.T) {
	initialWg := &sync.WaitGroup{}
	finalWg := &sync.WaitGroup{}
	locker := NewLocker()
	str := "this is a test string"

	initialWg.Add(100)
	finalWg.Add(100)
	for i := 0; i < 100; i++ {
		go (func() {
			initialWg.Done()
			locker.lock(str)
			// Wait for all goroutines to lock first
			initialWg.Wait()
			locker.unlock(str)
			finalWg.Done()
		})()
	}
	finalWg.Wait()
}
