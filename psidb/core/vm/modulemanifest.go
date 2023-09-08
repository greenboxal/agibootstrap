package vm

import "github.com/invopop/jsonschema"

type ModuleManifestFile struct {
	Name string `json:"name,omitempty"`
	Hash string `json:"hash,omitempty"`
}

type ModuleManifest struct {
	Name string `json:"name,omitempty"`

	Files []ModuleManifestFile `json:"files,omitempty"`

	Schema *jsonschema.Schema `json:"schema,omitempty"`
}
