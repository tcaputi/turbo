package turbo

import (
	"fmt"
	"sync"
	"testing"
)

func TestLocker(t *testing.T) {
	initialWg := &sync.WaitGroup{}
	finalWg := &sync.WaitGroup{}
	locker := NewLocker()
	str := "this is a test string"

	initialWg.Add(30)
	finalWg.Add(30)
	for i := 0; i < 30; i++ {
		go (func() {
			initialWg.Done()
			locker.lock(str)
			// Wait for all goroutines to lock first
			fmt.Println("<->")
			initialWg.Wait()
			fmt.Println("->")
			locker.unlock(str)
			fmt.Println("<-")
			finalWg.Done()
		})()
	}
	finalWg.Wait()
}
