package project

type FileType interface {
	GetName() string
	GetDescription() string
	GetIcon() string

	IsBinary() bool
	IsReadOnly() bool

	GetExtensions() []string

	String() error
}

type LanguageFileType interface {
	FileType

	GetLanguage() Language
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
func (l *FileTypeBase) GetIcon() string         { return l.Icon }
func (l *FileTypeBase) GetExtensions() []string { return l.Extensions }
func (l *FileTypeBase) IsBinary() bool          { return l.Binary }
func (l *FileTypeBase) IsReadOnly() bool        { return l.ReadOnly }

func (l *FileTypeBase) String() string { return l.Name }

type LanguageFileTypeBase struct {
	FileTypeBase

	Language Language
}

func (l *LanguageFileTypeBase) String() error {
	//TODO implement me
	panic("implement me")
}

func (l *LanguageFileTypeBase) GetLanguage() Language { return l.Language }
