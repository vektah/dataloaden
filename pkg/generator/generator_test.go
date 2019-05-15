package generator

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseType(t *testing.T) {
	require.Equal(t, &goType{Name: "string"}, parse("string"))
	require.Equal(t, &goType{Name: "Time", ImportPath: "time", ImportName: "time"}, parse("time.Time"))
	require.Equal(t, &goType{
		Name:       "Foo",
		ImportPath: "github.com/vektah/dataloaden/pkg/generator/testdata/mismatch",
		ImportName: "mismatched",
	}, parse("github.com/vektah/dataloaden/pkg/generator/testdata/mismatch.Foo"))
}

func parse(s string) *goType {
	t, err := parseType(s)
	if err != nil {
		panic(err)
	}

	return t
}
