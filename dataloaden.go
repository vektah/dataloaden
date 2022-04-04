package dataloaden

import (
	"sync"
	"time"
)

// LoaderConfig captures the config to create a new Loader
type LoaderConfig[K comparable, T any] struct {
	// Fetch is a method that provides the data for the loader
	Fetch func(keys []K) ([]T, []error)

	// Wait is how long wait before sending a batch
	Wait time.Duration

	// MaxBatch will limit the maximum number of keys to send in one batch, 0 = not limit
	MaxBatch int
}

// NewLoader creates a new Loader given a fetch, wait, and maxBatch
func NewLoader[K comparable, T any](config LoaderConfig[K, T]) *Loader[K, T] {
	return &Loader[K, T]{
		fetch:    config.Fetch,
		wait:     config.Wait,
		maxBatch: config.MaxBatch,
	}
}

// Loader batches and caches requests
type Loader[K comparable, T any] struct {
	// this method provides the data for the loader
	fetch func(keys []K) ([]T, []error)

	// how long to done before sending a batch
	wait time.Duration

	// this will limit the maximum number of keys to send in one batch, 0 = no limit
	maxBatch int

	// INTERNAL

	// lazily created cache
	cache map[K]T

	// the current batch. keys will continue to be collected until timeout is hit,
	// then everything will be sent to the fetch method and out to the listeners
	batch *loaderBatch[K, T]

	// mutex to prevent races
	mu sync.Mutex
}

type loaderBatch[K comparable, T any] struct {
	keys    []K
	data    []T
	error   []error
	closing bool
	done    chan struct{}
}

// Load a User by key, batching and caching will be applied automatically
func (l *Loader[K, T]) Load(key K) (T, error) {
	return l.LoadThunk(key)()
}

// LoadThunk returns a function that when called will block waiting for a User.
// This method should be used if you want one goroutine to make requests to many
// different data loaders without blocking until the thunk is called.
func (l *Loader[K, T]) LoadThunk(key K) func() (T, error) {
	l.mu.Lock()
	if it, ok := l.cache[key]; ok {
		l.mu.Unlock()
		return func() (T, error) {
			return it, nil
		}
	}
	if l.batch == nil {
		l.batch = &loaderBatch[K, T]{done: make(chan struct{})}
	}
	batch := l.batch
	pos := batch.keyIndex(l, key)
	l.mu.Unlock()

	return func() (T, error) {
		<-batch.done

		var data T
		if pos < len(batch.data) {
			data = batch.data[pos]
		}

		var err error
		// its convenient to be able to return a single error for everything
		if len(batch.error) == 1 {
			err = batch.error[0]
		} else if batch.error != nil {
			err = batch.error[pos]
		}

		if err == nil {
			l.mu.Lock()
			l.unsafeSet(key, data)
			l.mu.Unlock()
		}

		return data, err
	}
}

// LoadAll fetches many keys at once. It will be broken into appropriate sized
// sub batches depending on how the loader is configured
func (l *Loader[K, T]) LoadAll(keys []K) ([]T, []error) {
	results := make([]func() (T, error), len(keys))

	for i, key := range keys {
		results[i] = l.LoadThunk(key)
	}

	users := make([]T, len(keys))
	errors := make([]error, len(keys))
	for i, thunk := range results {
		users[i], errors[i] = thunk()
	}
	return users, errors
}

// LoadAllThunk returns a function that when called will block waiting for a Users.
// This method should be used if you want one goroutine to make requests to many
// different data loaders without blocking until the thunk is called.
func (l *Loader[K, T]) LoadAllThunk(keys []K) func() ([]T, []error) {
	results := make([]func() (T, error), len(keys))
	for i, key := range keys {
		results[i] = l.LoadThunk(key)
	}
	return func() ([]T, []error) {
		users := make([]T, len(keys))
		errors := make([]error, len(keys))
		for i, thunk := range results {
			users[i], errors[i] = thunk()
		}
		return users, errors
	}
}

// Prime the cache with the provided key and value. If the key already exists, no change is made
// and false is returned.
// (To forcefully prime the cache, clear the key first with loader.clear(key).prime(key, value).)
func (l *Loader[K, T]) Prime(key K, value T) bool {
	l.mu.Lock()
	var found bool
	if _, found = l.cache[key]; !found {
		l.unsafeSet(key, value)
	}
	l.mu.Unlock()
	return !found
}

// Clear the value at key from the cache, if it exists
func (l *Loader[K, T]) Clear(key K) {
	l.mu.Lock()
	delete(l.cache, key)
	l.mu.Unlock()
}

func (l *Loader[K, T]) unsafeSet(key K, value T) {
	if l.cache == nil {
		l.cache = map[K]T{}
	}
	l.cache[key] = value
}

// keyIndex will return the location of the key in the batch, if its not found
// it will add the key to the batch
func (b *loaderBatch[K, T]) keyIndex(l *Loader[K, T], key K) int {
	for i, existingKey := range b.keys {
		if key == existingKey {
			return i
		}
	}

	pos := len(b.keys)
	b.keys = append(b.keys, key)
	if pos == 0 {
		go b.startTimer(l)
	}

	if l.maxBatch != 0 && pos >= l.maxBatch-1 {
		if !b.closing {
			b.closing = true
			l.batch = nil
			go b.end(l)
		}
	}

	return pos
}

func (b *loaderBatch[K, T]) startTimer(l *Loader[K, T]) {
	time.Sleep(l.wait)
	l.mu.Lock()

	// we must have hit a batch limit and are already finalizing this batch
	if b.closing {
		l.mu.Unlock()
		return
	}

	l.batch = nil
	l.mu.Unlock()

	b.end(l)
}

func (b *loaderBatch[K, T]) end(l *Loader[K, T]) {
	b.data, b.error = l.fetch(b.keys)
	close(b.done)
}
