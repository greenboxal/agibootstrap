package codex

import (
	"context"
	"fmt"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/google/uuid"
	"github.com/greenboxal/aip/aip-langchain/pkg/chunkers"
	"github.com/greenboxal/aip/aip-langchain/pkg/llm"
	"github.com/ipfs/go-datastore"
	badger "github.com/ipfs/go-ds-badger"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/greenboxal/agibootstrap/pkg/gpt"
	"github.com/greenboxal/agibootstrap/pkg/gpt/cache"
	"github.com/greenboxal/agibootstrap/pkg/platform/db/fti"
	"github.com/greenboxal/agibootstrap/pkg/platform/db/graphindex"
	"github.com/greenboxal/agibootstrap/pkg/platform/db/graphstore"
	"github.com/greenboxal/agibootstrap/pkg/platform/db/thoughtdb"
	"github.com/greenboxal/agibootstrap/pkg/platform/logging"
	"github.com/greenboxal/agibootstrap/pkg/platform/project"
	tasks "github.com/greenboxal/agibootstrap/pkg/platform/tasks"
	"github.com/greenboxal/agibootstrap/pkg/platform/vfs"
	"github.com/greenboxal/agibootstrap/pkg/platform/vfs/repofs"
	"github.com/greenboxal/agibootstrap/pkg/psi"
	"github.com/greenboxal/agibootstrap/pkg/psi/analysis"
)

const SourceFileEdge psi.TypedEdgeKind[project.SourceFile] = "SourceFile"

type Project struct {
	psi.NodeBase

	logger *zap.SugaredLogger

	uuid   string
	config project.Config

	ds datastore.Batching

	indexedGraph *graphstore.IndexedGraph
	indexManager *graphindex.Manager

	embeddingManager *cache.EmbeddingCacheManager
	embedder         llm.Embedder

	vfsManager *vfs.Manager
	rootFs     vfs.FileSystem
	vcsFs      repofs.FS

	fileTypeRegistry *project.FileTypeProvider
	langRegistry     *project.LanguageProvider

	repo            *fti.Repository
	thoughtRepo     *thoughtdb.Repo
	taskManager     *tasks.Manager
	syncManager     *SyncManager
	analysisManager *AnalysisManager

	rootPath string
	rootNode vfs.Node

	fset *token.FileSet
}

var ProjectType = psi.DefineNodeType[*Project](psi.WithRuntimeOnly())

func NewBareProject(rootPath string) (*Project, error) {
	repo, err := fti.NewRepository(rootPath)

	if err != nil {
		return nil, err
	}

	dsOpts := badger.DefaultOptions
	dsPath := repo.ResolveDbPath("codex", "db")

	if err := os.MkdirAll(dsPath, 0755); err != nil {
		return nil, errors.Wrap(err, "failed to create datastore directory")
	}

	ds, err := badger.NewDatastore(dsPath, &dsOpts)

	if err != nil {
		return nil, errors.Wrap(err, "failed to create datastore")
	}

	p := &Project{
		logger: logging.GetLogger("codex"),

		rootPath: rootPath,

		ds:   ds,
		repo: repo,
	}

	p.fset = token.NewFileSet()
	p.fileTypeRegistry = project.NewFileTypeProvider()
	p.langRegistry = project.NewLanguageProvider(p)
	p.analysisManager = NewAnalysisManager(p)

	p.vfsManager, err = vfs.NewManager(repo.ResolveDbPath("vfs-cache"))

	if err != nil {
		return nil, err
	}

	p.embeddingManager, err = cache.NewEmbeddingCacheManager(repo.ResolveDbPath("embedding-cache"))

	if err != nil {
		return nil, err
	}

	p.embedder = p.embeddingManager.GetEmbedder(gpt.GlobalEmbedder)

	p.vcsFs, err = repofs.NewFS(rootPath)

	if err != nil {
		return nil, err
	}

	p.rootFs, err = p.vfsManager.CreateLocalFS(rootPath, vfs.WithPathFilter(func(p string) bool {
		if repo.IsIgnored(p) {
			return false
		}

		return true
	}))

	if err != nil {
		return nil, err
	}

	return p, nil
}

// LoadProject creates a new codex project with the given root path.
// It initializes the project file system, repository, and other required data structures.
// It returns a pointer to the created Project object and an error if any.
func LoadProject(ctx context.Context, rootPath string) (*Project, error) {
	p, err := NewBareProject(rootPath)

	if err != nil {
		return nil, err
	}

	if err := p.Load(ctx); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Project) Init(self psi.Node, projectUuid string) {
	p.uuid = projectUuid

	p.NodeBase.Init(self, psi.WithNodeType(ProjectType))
}

func (p *Project) IsProjectValid(ctx context.Context) (bool, error) {
	projectUuid, err := p.ds.Get(ctx, datastore.NewKey("project-uuid"))

	if err != nil {
		if errors.Is(err, datastore.ErrNotFound) {
			projectUuid = nil
		} else {
			return false, errors.Wrap(err, "failed to get project uuid")
		}
	}

	if len(projectUuid) == 0 {
		return false, nil
	}

	return true, nil
}

func (p *Project) Create(ctx context.Context) error {
	if p.uuid != "" {
		return errors.New("project already initialized")
	}

	projectUuid := []byte(uuid.New().String())

	if err := os.MkdirAll(path.Join(p.rootPath, ".codex"), 0755); err != nil {
		return errors.Wrap(err, "failed to create codex directory")
	}

	if err := os.WriteFile(path.Join(p.rootPath, ".codex", "project-uuid"), projectUuid, 0644); err != nil {
		return errors.Wrap(err, "failed to write project uuid")
	}

	err := p.ds.Put(ctx, datastore.NewKey("project-uuid"), projectUuid)

	if err != nil {
		return errors.Wrap(err, "failed to put project uuid")
	}

	if err := p.Initialize(ctx, string(projectUuid)); err != nil {
		return err
	}

	if err := p.indexedGraph.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (p *Project) Load(ctx context.Context) error {
	projectUuid, err := p.ds.Get(ctx, datastore.NewKey("project-uuid"))

	if err != nil {
		if errors.Is(err, datastore.ErrNotFound) {
			projectUuid = nil
		} else {
			return errors.Wrap(err, "failed to get project uuid")
		}
	}

	if len(projectUuid) == 0 {
		return p.Create(ctx)
	}

	if err := p.Initialize(ctx, string(projectUuid)); err != nil {
		return err
	}

	return nil
}

func (p *Project) Initialize(ctx context.Context, projectUuid string) error {
	var err error

	if p.uuid != "" {
		return errors.New("project already initialized")
	}

	p.Init(p, projectUuid)

	if err := p.loadConfig(ctx); err != nil {
		return err
	}

	p.indexedGraph, err = graphstore.NewIndexedGraph(p.ds, p.repo.ResolveDbPath("codex", "wal"), p)

	if err != nil {
		return errors.Wrap(err, "failed to create graph")
	}

	p.indexManager = graphindex.NewManager(p.indexedGraph, p.repo.ResolveDbPath("codex", "index"))

	p.taskManager = tasks.NewManager()
	p.taskManager.SetParent(p)
	p.indexedGraph.Add(p.taskManager)

	p.thoughtRepo = thoughtdb.NewRepo(p.indexedGraph)
	p.thoughtRepo.SetParent(p)
	p.indexedGraph.Add(p.thoughtRepo)

	p.syncManager = NewSyncManager(p)
	p.syncManager.SetParent(p)
	p.indexedGraph.Add(p.syncManager)

	p.rootNode, err = p.vfsManager.GetNodeForPath(ctx, p.rootPath)

	if err != nil {
		return errors.Wrap(err, "failed to get root node")
	}

	p.rootNode.SetParent(p)
	p.indexedGraph.Add(p.rootNode)

	pathIndex, err := p.indexManager.OpenNodeIndex(ctx, "node-by-path", &graphindex.AnchoredEmbedder{
		Base:    p.embedder,
		Root:    p,
		Anchor:  p.rootNode,
		Chunker: &chunkers.TikToken{},
	})

	if err != nil {
		return errors.Wrap(err, "failed to open path index")
	}

	p.indexedGraph.AddListener(graphstore.IndexedGraphListenerFunc(func(node psi.Node) {
		switch node.(type) {
		case project.AstNode:
			if analysis.GetDirectNodeScope(node) == nil {
				return
			}
		}

		if err := pathIndex.IndexNode(context.Background(), node); err != nil {
			p.logger.Error(err)
		}
	}))

	if err := p.indexedGraph.RefreshNode(ctx, p); err != nil && !errors.Is(err, psi.ErrNodeNotFound) {
		return errors.Wrap(err, "failed to refresh project node")
	}

	if err := p.indexedGraph.RefreshNode(ctx, p.thoughtRepo); err != nil && !errors.Is(err, psi.ErrNodeNotFound) {
		return errors.Wrap(err, "failed to refresh thoughtdb node")
	}

	if err := p.bootstrap(ctx); err != nil {
		return err
	}

	return nil
}

func (p *Project) bootstrap(ctx context.Context) error {
	if err := p.syncManager.RequestSync(ctx, true); err != nil {
		return err
	}

	for _, lang := range p.config.Project.EnabledLanguages {
		if err := p.bootstrapLanguage(ctx, lang); err != nil {
			return err
		}
	}

	/*for _, mod := range p.config.Modules {
		if err := p.bootstrapModule(ctx, mod); err != nil {
			return err
		}
	}*/

	if err := p.syncManager.WaitForInitialSync(ctx); err != nil {
		return err
	}

	p.logger.Info("Project bootstrapped")

	return nil
}

func (p *Project) bootstrapLanguage(ctx context.Context, name string) error {
	factory := project.GetLanguageFactory(project.LanguageID(name))

	if factory == nil {
		return errors.Errorf("language %s not found", name)
	}

	p.langRegistry.Register(factory(p))

	return nil
}

func (p *Project) bootstrapModule(ctx context.Context, mod project.ModuleConfig) error {
	lang := p.langRegistry.GetLanguage(project.LanguageID(mod.Language))

	if lang == nil {
		return errors.Errorf("language %s not configured", mod.Language)
	}

	root, err := psi.Resolve(ctx, p.rootNode, mod.Path)

	if err != nil {
		return errors.Wrapf(err, "failed to resolve module root %s", mod.Path)
	}

	m, err := NewModule(p, mod, lang, root.(*vfs.Directory))

	if err != nil {
		return err
	}

	m.SetParent(p)

	return nil
}

func (p *Project) loadConfig(ctx context.Context) error {
	configPath := path.Join(p.rootPath, "Codex.project.toml")

	data, err := os.ReadFile(configPath)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return err
	}

	if _, err := toml.Decode(string(data), &p.config); err != nil {
		return errors.Wrap(err, "failed to decode config")
	}

	return nil
}

func (p *Project) UUID() string { return p.uuid }

func (p *Project) Graph() *graphstore.IndexedGraph             { return p.indexedGraph }
func (p *Project) LanguageProvider() *project.LanguageProvider { return p.langRegistry }
func (p *Project) FileTypeProvider() *project.FileTypeProvider { return p.fileTypeRegistry }
func (p *Project) TaskManager() *tasks.Manager                 { return p.taskManager }
func (p *Project) LogManager() *thoughtdb.Repo                 { return p.thoughtRepo }
func (p *Project) IndexManager() *graphindex.Manager           { return p.indexManager }
func (p *Project) AnalysisManager() *AnalysisManager           { return p.analysisManager }
func (p *Project) Repo() *fti.Repository                       { return p.repo }

// RootPath returns the root path of the project.
func (p *Project) RootPath() string   { return p.rootPath }
func (p *Project) RootNode() psi.Node { return p.rootNode }

// VcsFileSystem returns the file system interface of the project.
// It provides methods for managing the project's files and directories.
func (p *Project) VcsFileSystem() repofs.FS { return p.vcsFs }

func (p *Project) FileSet() *token.FileSet { return p.fset }
func (p *Project) Embedder() llm.Embedder  { return p.embedder }

// Sync synchronizes the project with the file system.
// It scans all files in the project directory and imports
// valid files into the project. Valid files are those that
// have valid file extensions specified in the `validExtensions`
// slice. The function returns an error if any occurs during
// the sync process.
func (p *Project) Sync(ctx context.Context) error {
	return p.syncManager.RequestSync(ctx, false)
}

// GetSourceFile retrieves the source file with the given filename from the project.
// It returns a pointer to the psi.SourceFile and any error that occurred during the process.
func (p *Project) GetSourceFile(ctx context.Context, filename string) (_ project.SourceFile, err error) {
	relPath, err := filepath.Rel(p.rootPath, filename)

	if err != nil {
		return nil, err
	}

	psiPath := psi.MustParsePath(relPath)
	fileNode, err := psi.ResolvePath(ctx, p.rootNode, psiPath)

	if err != nil {
		return nil, err
	}

	existing := psi.ResolveEdge(fileNode, SourceFileEdge.Singleton())

	if existing == nil {
		if err != nil {
			return nil, err
		}

		lang := p.langRegistry.ResolveExtension(filepath.Base(filename))

		if lang == nil {
			return nil, fmt.Errorf("failed to resolve language for file %s", filename)
		}

		existing = lang.CreateSourceFile(ctx, filename, &repofs.FsFileHandle{
			FS:   p.vcsFs,
			Path: strings.TrimPrefix(filename, p.rootPath+"/"),
		})

		existing.SetParent(fileNode)
		fileNode.SetEdge(SourceFileEdge.Singleton(), existing)

		err = existing.Load(ctx)

		if err != nil {
			p.logger.Warn(err)
			return nil, err
		}
	}

	return existing, nil
}

// Reindex is a method that performs the reindexing operation for the project.
// It updates the index of the project to reflect any changes made to its files.
// The function returns an error if any error occurs during the reindexing process.
func (p *Project) Reindex() error {
	return nil
}

func (p *Project) Shutdown(ctx context.Context) error {
	if err := p.analysisManager.Close(); err != nil {
		return errors.Wrap(err, "failed to close analysis manager")
	}

	if err := p.indexManager.Close(); err != nil {
		return errors.Wrap(err, "failed to close index manager")
	}

	if err := p.indexedGraph.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "failed to close graph")
	}

	if err := p.vfsManager.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "failed to close vfs manager")
	}

	if err := p.embeddingManager.Close(); err != nil {
		return errors.Wrap(err, "failed to close embedding manager")
	}

	return p.ds.Close()
}
