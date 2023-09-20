package jsonstream

import (
	"net/url"
	"strings"
)

type ParserPath struct{ s []string }

func (p ParserPath) Depth() int {
	return len(p.s)
}

func (p ParserPath) Path() []string {
	return append([]string{}, p.s...)
}

func (p ParserPath) Parent() ParserPath {
	if len(p.s) == 0 {
		return p
	}

	return ParserPath{s: p.s[:len(p.s)-1]}
}

func (p ParserPath) Push(s string) ParserPath {
	return ParserPath{s: append(p.s, s)}
}

func (p ParserPath) String() string {
	urlEncoded := make([]string, len(p.s))

	for i, s := range p.s {
		urlEncoded[i] = url.PathEscape(s)
	}

	return strings.Join(urlEncoded, "/")
}
