package promptml

import "github.com/greenboxal/aip/aip-controller/pkg/collective/msn"

type Message struct {
	Container

	From string
	To   string
	Role msn.Role
}
