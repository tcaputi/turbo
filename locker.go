package turbo

import (
	"fmt"
	"sync"
)

type Lock struct {
	main   *sync.Mutex
	queue  *sync.Mutex
	count  uint
	key    string
	locker *Locker
}

type Locker struct {
	locks map[string]*Lock
}

func NewLocker() *Locker {
	return &Locker{
		locks: make(map[string]*Lock),
	}
}

func (locker *Locker) lock(key string) {
	if locker.locks[key] == nil {
		locker.locks[key] = &Lock{
			main:   &sync.Mutex{},
			queue:  &sync.Mutex{},
			count:  0,
			key:    key,
			locker: locker,
		}
	}
	locker.locks[key].lock()
}

func (locker *Locker) unlock(key string) {
	if locker.locks[key] == nil {
		return
	}
	locker.locks[key].unlock()
}

func (lock *Lock) lock() {
	// Make sure the mutex exists first
	if lock.main == nil {
		lock.main = &sync.Mutex{}
	}
	// Make sure the mutex exists first
	if lock.queue == nil {
		lock.queue = &sync.Mutex{}
	}

	lock.queue.Lock()
	// Check if the lock.locker knows this lock exists
	if lock.locker.locks[lock.key] == nil {
		lock.locker.locks[lock.key] = lock
	}
	lock.count += 1
	fmt.Println("lock on", lock.key, "has increased queue to size", lock.count)
	lock.queue.Unlock()
	// Do the mutal exclusion
	lock.main.Lock()
}

func (lock *Lock) unlock() {
	if lock.main == nil || lock.queue == nil {
		return
	}
	// Undo the mutal exclusion
	lock.main.Unlock()
	// Drop from the queue
	lock.queue.Lock()
	lock.count -= 1
	fmt.Println("lock on", lock.key, "has decreased queue to size", lock.count)
	if lock.count <= 0 {
		// Remove this lock from the lock.locker
		delete(lock.locker.locks, lock.key)
		lock.main = nil
		lock.queue = nil
	}
	if lock.queue != nil {
		lock.queue.Unlock()
	}
}
