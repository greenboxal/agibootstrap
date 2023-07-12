package filetypes

import "github.com/greenboxal/agibootstrap/pkg/psi/langs"

type FileType interface {
	GetName() string
	GetDescription() string
	GetIcon() string

	IsBinary() bool
	IsReadOnly() bool

	GetExtensions() []string
}

type LanguageFileType interface {
	FileType

	GetLanguage() langs.Language
}

type FileTypeBase struct {
	Name        string
	Description string
	Icon        string

	Binary   bool
	ReadOnly bool

	Extensions []string
}

func (l *FileTypeBase) GetName() string         { return l.Name }
func (l *FileTypeBase) GetDescription() string  { return l.Description }
func (l *FileTypeBase) GetExtensions() []string { return l.Extensions }
func (l *FileTypeBase) IsBinary() bool          { return l.Binary }
func (l *FileTypeBase) IsReadOnly() bool        { return l.ReadOnly }

type LanguageFileTypeBase struct {
	FileTypeBase

	Language langs.Language
}

func (l *LanguageFileTypeBase) GetLanguage() langs.Language { return l.Language }
