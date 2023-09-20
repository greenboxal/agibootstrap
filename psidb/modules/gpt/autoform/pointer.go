package autoform

import (
	"reflect"
	"strings"

	"github.com/go-errors/errors"
)

type JsonPointer struct {
	referenceTokens []string
}

func NewJsonPointer(jsonPointerString string) (p JsonPointer, err error) {
	if len(jsonPointerString) == 0 {
		// Keep referenceTokens nil
		return
	}
	if jsonPointerString[0] != '/' {
		return p, errors.New("JSON pointer must be empty or start with a \"/\"")
	}

	p.referenceTokens = strings.Split(jsonPointerString[1:], "/")

	return
}

func (p JsonPointer) Len() int {
	return len(p.referenceTokens)
}

func (p JsonPointer) Elements() []string {
	return p.referenceTokens
}

func (p JsonPointer) String() string {
	return "/" + strings.Join(p.referenceTokens, "/")
}

func (p JsonPointer) Parent() JsonPointer {
	if len(p.referenceTokens) == 0 {
		return p
	}

	return JsonPointer{referenceTokens: p.referenceTokens[:len(p.referenceTokens)-1]}
}

func (p JsonPointer) Last() string {
	if len(p.referenceTokens) == 0 {
		return ""
	}

	return p.referenceTokens[len(p.referenceTokens)-1]
}

var stringType = reflect.TypeOf((*string)(nil)).Elem()
var anyType = reflect.TypeOf((*any)(nil)).Elem()
var anySlice = reflect.SliceOf(anyType)
var anyMap = reflect.MapOf(stringType, anyType)
