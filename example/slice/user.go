//go:generate go run github.com/vektah/dataloaden UserSliceLoader int []github.com/vektah/dataloaden/example.User

package slice

import (
	"context"
	"strconv"
	"time"

	"github.com/vektah/dataloaden/example"
)

func NewLoader() *UserSliceLoader {
	return &UserSliceLoader{
		wait:     2 * time.Millisecond,
		maxBatch: 100,
		fetch: func(ctx context.Context, keys []int) ([][]example.User, []error) {
			users := make([][]example.User, len(keys))
			errors := make([]error, len(keys))

			for i, key := range keys {
				users[i] = []example.User{{ID: strconv.Itoa(key), Name: "user " + strconv.Itoa(key)}}
			}
			return users, errors
		},
	}
}
