package project

import "github.com/greenboxal/agibootstrap/psidb/psi"

type AstNode interface {
	psi.Node

	GetLanguage() Language
	GetSourceFile() SourceFile

	IsAstNode()
}

type AstNodeBase struct {
	psi.NodeBase
}

func (n *AstNodeBase) IsAstNode() {}

func (n *AstNodeBase) GetLanguage() Language {
	return n.GetSourceFile().Language()
}

func (n *AstNodeBase) GetSourceFile() SourceFile {
	for parent := n.Parent(); parent != nil; parent = parent.Parent() {
		if sf, ok := parent.(SourceFile); ok {
			return sf
		}
	}

	return nil
}
