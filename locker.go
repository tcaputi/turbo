package turbo

import (
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

func (locker *Locker) lock(path string) {
	cascadePath(path, false, func(currPath string) {
		locker.lockOne(currPath)
	})
}

func (locker *Locker) unlock(path string) {
	cascadePath(path, false, func(currPath string) {
		locker.unlockOne(currPath)
	})
}

func (locker *Locker) lockOne(key string) {
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

func (locker *Locker) unlockOne(key string) {
	if locker.locks[key] == nil {
		return
	}
	locker.locks[key].unlock()
}

func (lock *Lock) lock() {
	// Make sure the mutex queue exists
	if lock.queue == nil {
		lock.queue = &sync.Mutex{}
	}

	lock.queue.Lock()
	// Make sure the main mutex exists
	if lock.main == nil {
		lock.main = &sync.Mutex{}
	}
	// Check if the lock.locker knows this lock exists
	if lock.locker.locks[lock.key] == nil {
		lock.locker.locks[lock.key] = lock
	}
	lock.count += 1
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
	if lock.count <= 0 {
		// Remove this lock from the lock.locker
		delete(lock.locker.locks, lock.key)
		// Dispose of the mutexes
		lock.queue.Unlock()
		lock.main = nil
		lock.queue = nil
	} else {
		lock.queue.Unlock()
	}
}
