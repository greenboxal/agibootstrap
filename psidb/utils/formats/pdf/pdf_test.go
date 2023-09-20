package pdf

import (
	"testing"

	pdfcpu "github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/stretchr/testify/require"
)

func TestPDF(t *testing.T) {
	ctx, err := pdfcpu.ReadContextFile("/Users/jonathanlima/Downloads/inst._53.pdf")

	require.NoError(t, err)
	require.NotNil(t, ctx)

}
