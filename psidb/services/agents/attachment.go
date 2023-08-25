package agents

import "github.com/greenboxal/agibootstrap/pkg/psi"

type Attachment struct {
	psi.NodeBase

	// The name of the attachment
	Name string `json:"name"`
}

var AttachmentType = psi.DefineNodeType[*Attachment]()

func NewAttachment(name string) *Attachment {
	att := &Attachment{}
	att.Init(att)

	return att
}

func (att *Attachment) PsiNodeName() string { return att.Name }
