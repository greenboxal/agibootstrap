package fti

import (
	"image"
	"image/color"
	"image/png"
	"io"

	"github.com/greenboxal/aip/aip-langchain/pkg/chunkers"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm"
)

type ObjectSnapshotMetadata struct {
	Path       string `json:"path"`
	Hash       string `json:"hash"`
	ChunkCount []int  `json:"chunk_count"`
}

type ObjectSnapshotImage struct {
	Chunks     []chunkers.Chunk
	Embeddings []llm.Embedding
}

func (osi *ObjectSnapshotImage) ReadFrom(r io.Reader) (int, error) {
	return 0, nil
}

// WriteTo writes the ObjectSnapshotImage to the given io.Writer in PNG format.
// It generates an image representation of the ObjectSnapshotImage by assigning colors
// based on the embedding values.
func (osi *ObjectSnapshotImage) WriteTo(w io.Writer) (int, error) {
	img := image.NewRGBA(image.Rect(0, 0, len(osi.Chunks)*25, 25))

	for i := range osi.Chunks {
		embedding := osi.Embeddings[i]
		n := 25

		for x := 0; x < n; x++ {
			for y := 0; y < n; y++ {
				idx := (y*n + x) * 3

				if idx >= len(embedding.Embeddings) {
					continue
				}

				c := color.RGBA{
					R: uint8(embedding.Embeddings[idx+0] * 256.0),
					G: uint8(embedding.Embeddings[idx+1] * 255.0),
					B: uint8(embedding.Embeddings[idx+2] * 255.0),
					A: 255,
				}

				img.Set(i*25+x, y, c)
			}
		}
	}

	err := png.Encode(w, img)
	if err != nil {
		return 0, err
	}

	return 0, nil
}
