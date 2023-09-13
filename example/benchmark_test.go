package example

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"testing"
	"time"
)

func BenchmarkLoader(b *testing.B) {
	dl := &UserLoader{
		wait:     500 * time.Nanosecond,
		maxBatch: 100,
		fetch: func(fCtx context.Context, keys []string) ([]*User, []error) {
			users := make([]*User, len(keys))
			errors := make([]error, len(keys))

			for i, key := range keys {
				if rand.Int()%100 == 1 {
					errors[i] = fmt.Errorf("user not found")
				} else if rand.Int()%100 == 1 {
					users[i] = nil
				} else {
					users[i] = &User{ID: key, Name: "user " + key}
				}
			}
			return users, errors
		},
	}

	b.Run("caches", func(b *testing.B) {
		thunks := make([]func() (*User, error), b.N)
		for i := 0; i < b.N; i++ {
			thunks[i] = dl.LoadThunk(context.Background(), strconv.Itoa(rand.Int()%300))
		}

		for i := 0; i < b.N; i++ {
			thunks[i]()
		}
	})

	b.Run("random spread", func(b *testing.B) {
		thunks := make([]func() (*User, error), b.N)
		for i := 0; i < b.N; i++ {
			thunks[i] = dl.LoadThunk(context.Background(), strconv.Itoa(rand.Int()))
		}

		for i := 0; i < b.N; i++ {
			thunks[i]()
		}
	})

	b.Run("concurently", func(b *testing.B) {
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				for j := 0; j < b.N; j++ {
					dl.Load(context.Background(), strconv.Itoa(rand.Int()))
				}
				wg.Done()
			}()
		}
		wg.Wait()
	})
}
