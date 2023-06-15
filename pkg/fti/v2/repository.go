package fti

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/DataIntelligenceCrew/go-faiss"
	"github.com/greenboxal/aip/aip-langchain/pkg/chunkers"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm"
	"github.com/greenboxal/aip/aip-langchain/pkg/providers/openai"
	"github.com/pkg/errors"
	ignore "github.com/sabhiram/go-gitignore"
)

var ErrNoConfig = errors.New("no config file found")
var ErrAbort = errors.New("abort")

type Repository struct {
	config Config

	repoPath   string
	ftiPath    string
	configPath string

	embedder llm.Embedder
	chunker  chunkers.Chunker
	index    *OnlineIndex

	ignore *ignore.GitIgnore
}

// NewRepository creates a new Repository with the given repository path.
// It initializes the repository by loading the configuration and ignore file,
// creating a new online index, and loading the index if it exists.
func NewRepository(repoPath string) (r *Repository, err error) {
	r = &Repository{}

	r.chunker = chunkers.TikToken{}
	r.embedder = &openai.Embedder{
		Client: openai.NewClient(),
		Model:  openai.AdaEmbeddingV2,
	}

	r.repoPath = repoPath
	r.ftiPath = filepath.Join(r.repoPath, ".fti")
	r.configPath = r.ResolveDbPath("config.json")

	if err := r.loadConfig(); err != nil {
		if err != ErrNoConfig {
			return nil, err
		}
	}

	if err := r.loadIgnoreFile(); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	r.index, err = NewOnlineIndex(r)

	if err != nil {
		return nil, err
	}

	if err := r.loadIndex(); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	return r, nil
}

func (r *Repository) RepoPath() string { return r.repoPath }

func (r *Repository) ResolveDbPath(p ...string) string {
	return filepath.Join(r.ftiPath, filepath.Join(p...))
}

func (r *Repository) ResolvePath(p ...string) string {
	return filepath.Join(r.repoPath, filepath.Join(p...))
}

func (r *Repository) FileExists(filePath string) bool {
	_, err := os.Stat(filePath)

	return err == nil
}

func (r *Repository) OpenFile(filePath string) (fs.File, error) {
	return os.OpenFile(filePath, os.O_RDONLY, os.ModePerm)
}

func (r *Repository) IsIgnored(name string) bool {
	p := r.RelativeToRoot(name)

	return r.ignore.MatchesPath(p)
}

func (r *Repository) RelativeToRoot(name string) string {
	p, err := filepath.Rel(r.repoPath, name)

	if err != nil {
		return name
	}

	return p
}

func (r *Repository) loadConfig() error {
	if !r.FileExists(r.configPath) {
		return ErrNoConfig
	}

	f, err := r.OpenFile(r.configPath)

	if err != nil {
		return err
	}

	defer f.Close()

	data, err := io.ReadAll(f)

	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &r.config); err != nil {
		return err
	}

	return nil
}

func (r *Repository) loadIgnoreFile() error {
	ignoreFilePath := r.ResolvePath(".ftiignore")

	if !r.FileExists(ignoreFilePath) {
		return nil
	}

	result, err := ignore.CompileIgnoreFile(ignoreFilePath)

	if err != nil {
		return err
	}

	r.ignore = result

	return nil
}

// IterateFiles returns an iterator that iterates over the files in the repository.
// It filters directories, files outside the repository path, and ignored files based on the repository's ignore file.
// The context parameter can be used to cancel the iteration.
func (r *Repository) IterateFiles(ctx context.Context) Iterator[FileCursor] {
	files := IterateFiles(ctx, r.repoPath)

	files = Filter(files, func(f FileCursor) bool {
		if f.IsDir() {
			return false
		}

		relPath, err := filepath.Rel(r.ftiPath, f.Path)

		if err == nil && !strings.HasPrefix(relPath, "..") {
			return false
		}

		if r.IsIgnored(f.Path) {
			return false
		}

		return true
	})

	return files
}

// Init initializes the repository by creating the necessary directories and configuration file.
// It creates the .fti directory and writes the default configuration to the config.json file.
func (r *Repository) Init() error {
	err := os.Mkdir(r.ftiPath, os.ModePerm)
	if err != nil {
		return err
	}

	configData, err := json.MarshalIndent(defaultConfig, "", "\t")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(r.configPath, configData, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) Update(ctx context.Context) error {
	for it := r.IterateFiles(ctx); it.Next(); {
		f := it.Item()

		if err := r.UpdateFile(ctx, f); err != nil {
			return err
		}
	}

	err := faiss.WriteIndex(r.index.idx, r.ResolveDbPath("index.faiss"))

	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) UpdateFile(ctx context.Context, f FileCursor) error {
	fmt.Printf("Updating file %s\n", f.Path)

	fh, err := r.OpenFile(f.Path)

	if err != nil {
		return err
	}

	defer fh.Close()

	hasher := sha256.New()
	reader := io.TeeReader(fh, hasher)
	data, err := io.ReadAll(reader)

	if err != nil {
		return err
	}

	h := hasher.Sum(nil)
	fileHash := hex.EncodeToString(h)

	fileDir := r.ResolveDbPath("objects", fileHash)
	metaPath := filepath.Join(fileDir, "meta.json")

	if err := os.MkdirAll(fileDir, os.ModePerm); err != nil {
		return err
	}

	meta := &ObjectSnapshotMetadata{
		Path:       f.Path,
		Hash:       fileHash,
		ChunkCount: make([]int, len(r.config.ChunkSpecs)),
	}

	for i, chunkSpec := range r.config.ChunkSpecs {
		img, err := r.updateFileWithSpec(ctx, chunkSpec, fileDir, data)

		if err != nil {
			return err
		}

		meta.ChunkCount[i] = len(img.Chunks)
	}

	serialized, err := json.MarshalIndent(meta, "", "\t")

	if err != nil {
		return err
	}

	if err := os.WriteFile(metaPath, serialized, os.ModePerm); err != nil {
		return err
	}

	return nil
}

func (r *Repository) updateFileWithSpec(ctx context.Context, spec ChunkSpec, fileDir string, data []byte) (*ObjectSnapshotImage, error) {
	imagePath := filepath.Join(fileDir, fmt.Sprintf("%dm%d.png", spec.MaxTokens, spec.Overlap))

	chunks, err := r.chunker.SplitTextIntoChunks(ctx, string(data), spec.MaxTokens, spec.Overlap)

	if err != nil {
		return nil, err
	}

	chunksStr := make([]string, len(chunks))

	for i, chunk := range chunks {
		chunksStr[i] = chunk.Content
	}

	embeddings, err := r.embedder.GetEmbeddings(ctx, chunksStr)

	if err != nil {
		return nil, err
	}

	img := &ObjectSnapshotImage{
		Chunks:     chunks,
		Embeddings: embeddings,
	}

	for i, chunk := range chunks {
		emb := embeddings[i]

		chunkPath := filepath.Join(fileDir, fmt.Sprintf("%dm%d.%d.txt", spec.MaxTokens, spec.Overlap, chunk.Index))
		embPath := filepath.Join(fileDir, fmt.Sprintf("%dm%d.%d.f32", spec.MaxTokens, spec.Overlap, chunk.Index))

		if err := os.WriteFile(chunkPath, []byte(chunk.Content), os.ModePerm); err != nil {
			return nil, err
		}

		buffer := make([]byte, len(emb.Embeddings)*4)

		for j, f := range emb.Embeddings {
			binary.LittleEndian.PutUint32(buffer[j*4:], math.Float32bits(f))
		}

		if err := os.WriteFile(embPath, buffer, os.ModePerm); err != nil {
			return nil, err
		}
	}

	fh, err := os.OpenFile(imagePath, os.O_CREATE|os.O_WRONLY, os.ModePerm)

	if err != nil {
		return nil, err
	}

	defer fh.Close()

	if _, err := img.WriteTo(fh); err != nil {
		return nil, err
	}

	if err := r.index.Add(img); err != nil {
		return nil, err
	}

	return img, nil
}

func (r *Repository) Query(ctx context.Context, query string, k int64) ([]OnlineIndexQueryHit, error) {
	embs, err := r.embedder.GetEmbeddings(ctx, []string{query})

	if err != nil {
		return nil, err
	}

	hits, err := r.index.Query(embs[0], k)

	if err != nil {
		return nil, err
	}

	return hits, nil
}

func (r *Repository) loadIndex() error {
	p := r.ResolveDbPath("index.faiss")

	if !r.FileExists(p) {
		return nil
	}

	idx, err := faiss.ReadIndex(p, faiss.IOFlagMmap)

	if err != nil {
		return nil
	}

	if r.index.idx != nil {
		r.index.idx.Delete()
		r.index.idx = nil
	}

	r.index.idx = idx

	return nil
}
