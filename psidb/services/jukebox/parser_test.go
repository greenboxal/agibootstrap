package jukebox

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/greenboxal/agibootstrap/psidb/utils/sparsing"
)

var InputData = []byte(`
@1 PlayNote 1 C4 1
@2 PlayNote 1 D4 1
@3 PlayNote 1 E4 1
@4 SetBPM 120
@5 PlayNote 2 C4 1

@6 PitchBend 323

`)

func TestParser(t *testing.T) {
	p := NewParser()

	n, err := p.Write(InputData)

	require.NoError(t, err)
	require.Equal(t, len(InputData), n)

	err = p.Close()

	require.NoError(t, err)

	parsed := sparsing.ConsumeAs[*CommandSheetNode](p)

	parsed = parsed
}
