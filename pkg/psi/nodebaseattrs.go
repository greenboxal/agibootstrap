package psi

func (n *NodeBase) Attributes() map[string]interface{} {
	return n.attributes.Map()
}

func (n *NodeBase) SetAttribute(key string, value any) {
	n.attributes.Set(key, value)
}

func (n *NodeBase) GetAttribute(key string) (value any, ok bool) {
	return n.attributes.Get(key)
}

func (n *NodeBase) RemoveAttribute(key string) (value any, ok bool) {
	value, ok = n.attributes.Get(key)

	if !ok {
		return value, false
	}

	n.attributes.Remove(key)

	return
}
