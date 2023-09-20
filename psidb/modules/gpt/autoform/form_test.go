package autoform

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type TestFormData struct {
	Title string   `json:"title"`
	Array []string `json:"array"`

	NestedForm *TestFormData `json:"nested_form"`
}

func TestForm(t *testing.T) {
	f := NewForm(WithSchemaFor[TestFormData]())
	f.MustSetField("/title", "Hello, World!")
	f.MustSetField("/array/0", "Hello")
	f.MustSetField("/array/2", "World")
	f.MustSetField("/nested_form/title", "Hello again")
	f.MustSetField("/nested_form/array/0", "Hello")
	f.MustSetField("/nested_form/array/2", "again")

	require.EqualValues(t, f.MustGetField("/title"), "Hello, World!")
	require.EqualValues(t, f.MustGetField("/array/0"), "Hello")
	require.EqualValues(t, f.MustGetField("/array/2"), "World")
	require.EqualValues(t, f.MustGetField("/nested_form/title"), "Hello again")
	require.EqualValues(t, f.MustGetField("/nested_form/array/0"), "Hello")
	require.EqualValues(t, f.MustGetField("/nested_form/array/2"), "again")
}
