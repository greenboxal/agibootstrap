package pyparser

import "github.com/antlr4-go/antlr/v4"

type Python3ParserBase struct {
	*antlr.BaseParser

	version int
}

func (p *Python3ParserBase) SetVersion(version int) {
	p.version = version
}

func (p *Python3ParserBase) CheckVersion(version int) bool {
	return p.version == version
}
