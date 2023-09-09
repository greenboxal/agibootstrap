package vm

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/go-errors/errors"
	"github.com/invopop/jsonschema"

	"github.com/greenboxal/agibootstrap/psidb/psi"
)

type PackageFileManifest struct {
	Name string `json:"name"`
	Hash string `json:"hash"`
}

type ModuleManifest struct {
	Name       string             `json:"name"`
	Entrypoint string             `json:"entrypoint"`
	Schema     *jsonschema.Schema `json:"schema,omitempty"`
}

type PackageManifest struct {
	Name    string                `json:"name"`
	Files   []PackageFileManifest `json:"files,omitempty"`
	Modules []ModuleManifest      `json:"modules,omitempty"`
}

type IPackage interface {
	Register(ctx context.Context) error
}

var PackageInterface = psi.DefineNodeInterface[IPackage]()

type Package struct {
	psi.NodeBase

	Name     string            `json:"name,omitempty"`
	Manifest PackageManifest   `json:"manifest"`
	Files    map[string]string `json:"files,omitempty"`

	Registered bool `json:"registered"`
}

var PackageType = psi.DefineNodeType[*Package](
	psi.WithInterfaceFromNode(PackageInterface),
)

func (p *Package) PsiNodeName() string {
	if p.Name != "" {
		return p.Name
	}

	return p.Manifest.Name
}

func (p *Package) OnUpdate(ctx context.Context) error {
	if !p.Registered {
		for _, fileManifest := range p.Manifest.Files {
			file := p.Files[fileManifest.Name]

			if file == "" {
				return fmt.Errorf("file %s not found", fileManifest.Name)
			}

			if fileManifest.Hash != "" {
				parsed, err := hex.DecodeString(fileManifest.Hash)

				if err != nil {
					return errors.Errorf("failed to parse file hash: %w", err)
				}

				h := sha256.New()
				h.Write([]byte(file))
				actual := h.Sum(nil)

				if !bytes.Equal(parsed, actual) {
					return fmt.Errorf("file %s hash mismatch", fileManifest.Name)
				}
			}
		}

		if err := p.doRegister(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (p *Package) doRegister(ctx context.Context) error {
	if p.Registered {
		return nil
	}

	for _, modManifest := range p.Manifest.Modules {
		entrypoint := p.Files[modManifest.Entrypoint]

		if entrypoint == "" {
			return fmt.Errorf("entrypoint %s not found", modManifest.Entrypoint)
		}

		mod := NewModule(modManifest.Name, entrypoint)
		mod.SetParent(p)

		if err := mod.Update(ctx); err != nil {
			return err
		}

		if err := mod.Register(ctx); err != nil {
			return err
		}
	}

	p.Registered = true
	p.Invalidate()

	return nil
}

func (p *Package) Register(ctx context.Context) error {
	if err := p.doRegister(ctx); err != nil {
		return err
	}

	return p.Update(ctx)
}
