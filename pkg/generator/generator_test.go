package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTypeParsing(t *testing.T) {
	valueType, err := NewValueTypeFromString("github.com/dataloaden/example.User")

	require.NoError(t, err)

	assert.Equal(t, &ValueType{
		Name:       "User",
		ImportPath: "github.com/dataloaden/example",
	}, valueType)
}

func TestTypeParsingWithSlice(t *testing.T) {
	valueType, err := NewValueTypeFromString("[]github.com/dataloaden/example.User")

	require.NoError(t, err)

	assert.Equal(t, &ValueType{
		Name:       "User",
		ImportPath: "github.com/dataloaden/example",
		IsSlice:    true,
	}, valueType)
}

func TestTypeParsingWithPointer(t *testing.T) {
	valueType, err := NewValueTypeFromString("*github.com/dataloaden/example.User")

	require.NoError(t, err)

	assert.Equal(t, &ValueType{
		Name:       "User",
		ImportPath: "github.com/dataloaden/example",
		IsPointer:  true,
	}, valueType)
}

func TestTypeParsingWithSliceOfPointers(t *testing.T) {
	valueType, err := NewValueTypeFromString("[]*github.com/dataloaden/example.User")

	require.NoError(t, err)

	assert.Equal(t, &ValueType{
		Name:       "User",
		ImportPath: "github.com/dataloaden/example",
		IsSlice:    true,
		IsPointer:  true,
	}, valueType)
}

func TestTypeParsingInvalidType(t *testing.T) {
	valueType, err := NewValueTypeFromString("somethingWrong")

	require.Error(t, err)
	assert.Nil(t, valueType)
}
