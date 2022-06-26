package example

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserLoader(t *testing.T) {
	var fetches [][]string
	var mu sync.Mutex

	dl := &UserLoader{
		wait:     10 * time.Millisecond,
		maxBatch: 5,
		fetch: func(keys []string) ([]*User, []error) {
			mu.Lock()
			fetches = append(fetches, keys)
			mu.Unlock()

			users := make([]*User, len(keys))
			errors := make([]error, len(keys))

			for i, key := range keys {
				if strings.HasPrefix(key, "E") {
					errors[i] = fmt.Errorf("user not found")
				} else if strings.HasPrefix(key, "P") {
					panic("something bad happened")
				} else {
					users[i] = &User{ID: key, Name: "user " + key}
				}
			}
			return users, errors
		},
	}

	t.Run("fetch concurrent data", func(t *testing.T) {
		t.Run("load user successfully", func(t *testing.T) {
			t.Parallel()
			u, err := dl.Load("U1")
			require.NoError(t, err)
			require.Equal(t, u.ID, "U1")
		})

		t.Run("load failed user", func(t *testing.T) {
			t.Parallel()
			u, err := dl.Load("E1")
			require.Error(t, err)
			require.Nil(t, u)
		})

		t.Run("load many users", func(t *testing.T) {
			t.Parallel()
			u, err := dl.LoadAll([]string{"U2", "E2", "E3", "U4"})
			require.Equal(t, u[0].Name, "user U2")
			require.Equal(t, u[3].Name, "user U4")
			require.Error(t, err[1])
			require.Error(t, err[2])
		})

		t.Run("load thunk", func(t *testing.T) {
			t.Parallel()
			thunk1 := dl.LoadThunk("U5")
			thunk2 := dl.LoadThunk("E5")

			u1, err1 := thunk1()
			require.NoError(t, err1)
			require.Equal(t, "user U5", u1.Name)

			u2, err2 := thunk2()
			require.Error(t, err2)
			require.Nil(t, u2)
		})
	})

	t.Run("it sent two batches", func(t *testing.T) {
		mu.Lock()
		defer mu.Unlock()

		require.Len(t, fetches, 2)
		assert.Len(t, fetches[0], 5)
		assert.Len(t, fetches[1], 3)
	})

	t.Run("fetch more", func(t *testing.T) {

		t.Run("previously cached", func(t *testing.T) {
			t.Parallel()
			u, err := dl.Load("U1")
			require.NoError(t, err)
			require.Equal(t, u.ID, "U1")
		})

		t.Run("load many users", func(t *testing.T) {
			t.Parallel()
			u, err := dl.LoadAll([]string{"U2", "U4"})
			require.NoError(t, err[0])
			require.NoError(t, err[1])
			require.Equal(t, u[0].Name, "user U2")
			require.Equal(t, u[1].Name, "user U4")
		})
	})

	t.Run("no round trips", func(t *testing.T) {
		mu.Lock()
		defer mu.Unlock()

		require.Len(t, fetches, 2)
	})

	t.Run("fetch partial", func(t *testing.T) {
		t.Run("errors not in cache cache value", func(t *testing.T) {
			t.Parallel()
			u, err := dl.Load("E2")
			require.Nil(t, u)
			require.Error(t, err)
		})

		t.Run("load all", func(t *testing.T) {
			t.Parallel()
			u, err := dl.LoadAll([]string{"U1", "U4", "E1", "U9", "U5"})
			require.Equal(t, u[0].ID, "U1")
			require.Equal(t, u[1].ID, "U4")
			require.Error(t, err[2])
			require.Equal(t, u[3].ID, "U9")
			require.Equal(t, u[4].ID, "U5")
		})
	})

	t.Run("one partial trip", func(t *testing.T) {
		mu.Lock()
		defer mu.Unlock()

		require.Len(t, fetches, 3)
		require.Len(t, fetches[2], 3) // E1 U9 E2 in some random order
	})

	t.Run("primed reads dont hit the fetcher", func(t *testing.T) {
		dl.Prime("U99", &User{ID: "U99", Name: "Primed user"})
		u, err := dl.Load("U99")
		require.NoError(t, err)
		require.Equal(t, "Primed user", u.Name)

		require.Len(t, fetches, 3)
	})

	t.Run("priming in a loop is safe", func(t *testing.T) {
		users := []User{
			{ID: "Alpha", Name: "Alpha"},
			{ID: "Omega", Name: "Omega"},
		}
		for _, user := range users {
			dl.Prime(user.ID, &user)
		}

		u, err := dl.Load("Alpha")
		require.NoError(t, err)
		require.Equal(t, "Alpha", u.Name)

		u, err = dl.Load("Omega")
		require.NoError(t, err)
		require.Equal(t, "Omega", u.Name)

		require.Len(t, fetches, 3)
	})

	t.Run("cleared results will go back to the fetcher", func(t *testing.T) {
		dl.Clear("U99")
		u, err := dl.Load("U99")
		require.NoError(t, err)
		require.Equal(t, "user U99", u.Name)

		require.Len(t, fetches, 4)
	})

	t.Run("load all thunk", func(t *testing.T) {
		thunk1 := dl.LoadAllThunk([]string{"U5", "U6"})
		thunk2 := dl.LoadAllThunk([]string{"U6", "E6"})

		users1, err1 := thunk1()

		require.NoError(t, err1[0])
		require.NoError(t, err1[1])
		require.Equal(t, "user U5", users1[0].Name)
		require.Equal(t, "user U6", users1[1].Name)

		users2, err2 := thunk2()

		require.NoError(t, err2[0])
		require.Error(t, err2[1])
		require.Equal(t, "user U6", users2[0].Name)
	})

	t.Run("fetch panic with recover func", func(t *testing.T) {
		expectedErr := errors.New("transformed")
		dl.recover = func(interface{}) error {
			return expectedErr
		}
		u, err := dl.Load("P1")
		require.Nil(t, u)
		require.Equal(t, err, expectedErr)
		dl.recover = nil
	})

	t.Run("fetch panic with no recover func", func(t *testing.T) {
		u, err := dl.Load("P1")
		require.Nil(t, u)
		require.Error(t, err)
	})
}
