package combinator

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParser(t *testing.T) {
	_, err := json.Marshal(struct {
		Int    int
		Float  float64
		String string
		Array  []string
		Object struct {
			Bool bool
		}
	}{
		Int:    1,
		Float:  1.0,
		String: "hello",
		Array:  []string{"a", "b", "c"},
		Object: struct {
			Bool bool
		}{
			Bool: true,
		},
	})

	require.NoError(t, err)

}
