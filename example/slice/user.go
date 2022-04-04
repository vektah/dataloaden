package slice

import (
	"strconv"
	"time"

	"github.com/vektah/dataloaden"
	"github.com/vektah/dataloaden/example"
)

func NewLoader() *dataloaden.Loader[int, []example.User] {
	return dataloaden.NewLoader(dataloaden.LoaderConfig[int, []example.User]{
		Wait:     2 * time.Millisecond,
		MaxBatch: 100,
		Fetch: func(keys []int) ([][]example.User, []error) {
			users := make([][]example.User, len(keys))
			errors := make([]error, len(keys))

			for i, key := range keys {
				users[i] = []example.User{{ID: strconv.Itoa(key), Name: "user " + strconv.Itoa(key)}}
			}
			return users, errors
		},
	})
}
