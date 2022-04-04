package example

import (
	"time"

	"github.com/vektah/dataloaden"
)

// User is some kind of database backed model
type User struct {
	ID   string
	Name string
}

// NewLoader will collect user requests for 2 milliseconds and send them as a single batch to the fetch func
// normally fetch would be a database call.
func NewLoader() *dataloaden.Loader[string, *User] {
	return dataloaden.NewLoader(dataloaden.LoaderConfig[string, *User]{
		Wait:     2 * time.Millisecond,
		MaxBatch: 100,
		Fetch: func(keys []string) ([]*User, []error) {
			users := make([]*User, len(keys))
			errors := make([]error, len(keys))

			for i, key := range keys {
				users[i] = &User{ID: key, Name: "user " + key}
			}
			return users, errors
		},
	})
}
