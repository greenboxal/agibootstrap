package golang

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/greenboxal/agibootstrap/psidb/services/codeanalysis"
)

func TestParser(t *testing.T) {
	ctx := context.Background()
	src := loadSourceFile("/Users/jonathanlima/mark/agibootstrap/staging/psidb/services/codeanalysis/langs/go/parser.go")

	p := NewParser()
	result, err := p.Parse(ctx, src)

	require.NoError(t, err)
	require.NotNil(t, result)
}

func loadSourceFile(s string) *codeanalysis.SourceFile {
	data, err := os.ReadFile(s)

	if err != nil {
		panic(err)
	}

	return codeanalysis.NewSourceFile(s, string(data))
}
