package interpreter

import "karl/shape"

// ShapeValue is a callable shape object imported from .shape files.
type ShapeValue struct {
	Name  string
	Shape *shape.Type
}

func (s *ShapeValue) Type() ValueType { return SHAPE }
func (s *ShapeValue) Inspect() string {
	if s.Name == "" {
		return "<shape>"
	}
	return "<shape " + s.Name + ">"
}
