package jsonstream

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/alecthomas/repr"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestParser(t *testing.T) {
	data, err := json.Marshal(struct {
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

	p := NewParser(func(path ParserPath, node Node) error {
		fmt.Printf("%s (%d): %s- %T\n", path, path.Depth(), strings.Repeat("  ", path.Depth()), node)

		return nil
	})

	n, err := p.Write(data)

	require.NoError(t, err)
	require.Equal(t, len(data), n)

	node := p.PopArgument()
	val := node.(Value).GetValue()

	repr.Println(val)
}

func TestParserError(t *testing.T) {
	testCase := func(input string) func(t *testing.T) {
		return func(t *testing.T) {
			p := NewParser(func(path ParserPath, node Node) error {
				fmt.Printf("%s (%d): %s- %T\n", path, path.Depth(), strings.Repeat("  ", path.Depth()), node)

				return nil
			})

			_, err := p.Write([]byte(input))

			require.Error(t, err, "expected error")
		}
	}

	t.Run("missing closing brace", testCase(`{"a": 1, "b": 2, "c": 3`))
	t.Run("colon after value without key", testCase(`{"a": 1,: 2, "c": 3,}`))
	t.Run("invalid characters", testCase(`]'/`))
	t.Run("missing quotes around key", testCase(`{a: 1, "b": 2, "c": 3}`))
	t.Run("missing quotes around string value", testCase(`{"a": "hello, "world": 2}`))
	t.Run("comma after last element", testCase(`{"a": 1, "b": 2, "c": 3,}`))
	t.Run("missing key", testCase(`{"a": 1, :2, "c": 3}`))
	t.Run("unquoted key", testCase(`{a: 1, "b": 2, "c": 3}`))
	t.Run("missing comma", testCase(`{"a": 1 "b": 2, "c": 3}`))
	t.Run("extra comma", testCase(`{"a": 1, "b": 2, "c": 3,}`))
	t.Run("unclosed string", testCase(`{"a": "1, "b": 2, "c": 3}`))
}

var BrokenInput = `{"a": "1, "b": 2, "c": 3}`
var BrokenInputFixed = `{"a": "1", "b": 2, "c": 3}`

func TestParserRecovery(t *testing.T) {
	var recoverable *ParsingError

	p := NewParser(func(path ParserPath, node Node) error {
		fmt.Printf("%s (%d): %s- %T\n", path, path.Depth(), strings.Repeat("  ", path.Depth()), node)

		return nil
	})

	n, err := writeChunked(p, []byte(BrokenInput), 4)

	require.Error(t, err)

	if !errors.As(err, &recoverable) {
		require.Fail(t, "expected recoverable error")
	}

	p.Recover(recoverable.RecoverablePosition.Offset)

	recoveryBase := BrokenInput[:p.Position().Offset]
	recoveryExtra := BrokenInputFixed[len(recoveryBase):]
	recoveryTotal := recoveryBase + recoveryExtra

	n, err = writeChunked(p, []byte(recoveryExtra), 1)

	require.NoError(t, err)
	require.Equal(t, len(recoveryTotal), n)
}

func writeChunked(w io.WriteCloser, data []byte, size int) (int, error) {
	total := 0
	buffer := make([]byte, size)

	for len(data) > 0 {
		n := copy(buffer, data)
		data = data[n:]

		n, err := w.Write(buffer[:n])

		total += n

		if err != nil {
			return total, err
		}
	}

	if err := w.Close(); err != nil {
		return total, err
	}

	return total, nil
}
