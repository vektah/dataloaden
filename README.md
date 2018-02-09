### The DATALOADer gENerator

This is a tool for generating type safe data loaders for go, inspired by https://github.com/facebook/dataloader.


#### Getting started

First grab it:
```bash
go get -u github.com/vektah/dataloaden


# from inside the package you want to have the dataloader in:
dataloaden -type github.com/dataloaden/example.User
```

In another file in the same package, create the constructor method:
```go 
func NewLoader() *UserLoader {
	return &UserLoader{
		wait:     2 * time.Millisecond,
		maxBatch: 100,
		fetch: func(keys []string) ([]*User, []error) {
			users := make([]*User, len(keys))
			errors := make([]error, len(keys))

			for i, key := range keys {
				users[i] = &User{ID: key, Name: "user " + key}
			}
			return users, errors
		},
	}
}
``` 

Then wherever you want to call the dataloader
```go
loader := NewLoader()

user, err := loader.Load("123")
``` 

#### Wait, how do I use context with this?

I don't think context makes sense to be passed through a data loader. Consider a few scenarios:
1. a dataloader shared between requests: request A and B both get batched together, which context should be passed to the DB? context.Background is probably more suitable.
2. a dataloader per request for graphql: two different nodes in the graph get batched together, they have different context for tracing purposes, which should be passed to the db? neither, you should just use the root request context. 


So be explicit about your context:
```go
func NewLoader(ctx context.Context) *UserLoader {
	return &UserLoader{
		wait:     2 * time.Millisecond,
		maxBatch: 100,
		fetch: func(keys []string) ([]*User, []error) {
			// you now have a ctx to work with 
		},
	}
}
```

If you feel like I'm wrong please raise an issue.