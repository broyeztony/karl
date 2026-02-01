package interpreter

import "karl/shape"

// CodecValue carries format and optional shape-aware path mappings.
type CodecValue struct {
	Name     string
	Format   string
	Shape    *shape.Type
	Mappings []shape.CodecMapping
}

func (c *CodecValue) Type() ValueType { return CODEC }
func (c *CodecValue) Inspect() string {
	if c.Name == "" {
		return "<codec " + c.Format + ">"
	}
	return "<codec " + c.Name + ">"
}
