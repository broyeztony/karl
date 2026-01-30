package interpreter

import (
	"fmt"
	"strings"
)

func builtinJSONPath(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "jsonPath expects value and path"}
	}
	segments, err := jsonPathSegments(args[1])
	if err != nil {
		return nil, &RuntimeError{Message: err.Error()}
	}
	if len(segments) == 0 {
		return args[0], nil
	}
	val, ok := jsonPathLookup(args[0], segments)
	if !ok {
		return NullValue, nil
	}
	return val, nil
}

type jsonPathSegmentKind int

const (
	jsonPathKey jsonPathSegmentKind = iota
	jsonPathIndex
)

type jsonPathSegment struct {
	kind  jsonPathSegmentKind
	key   string
	index int
}

func jsonPathSegments(path Value) ([]jsonPathSegment, error) {
	switch v := path.(type) {
	case *String:
		segments, err := parseJSONPath(v.Value)
		if err != nil {
			return nil, fmt.Errorf("jsonPath error: %s", err.Error())
		}
		return segments, nil
	case *Array:
		return parseJSONPathArray(v)
	default:
		return nil, fmt.Errorf("jsonPath expects path as string or array")
	}
}

func parseJSONPathArray(arr *Array) ([]jsonPathSegment, error) {
	segments := make([]jsonPathSegment, 0, len(arr.Elements))
	for _, el := range arr.Elements {
		switch v := el.(type) {
		case *String:
			segments = append(segments, jsonPathSegment{kind: jsonPathKey, key: v.Value})
		case *Char:
			segments = append(segments, jsonPathSegment{kind: jsonPathKey, key: v.Value})
		case *Integer:
			segments = append(segments, jsonPathSegment{kind: jsonPathIndex, index: int(v.Value)})
		default:
			return nil, fmt.Errorf("jsonPath path array must contain strings or integers")
		}
	}
	return segments, nil
}

func parseJSONPath(path string) ([]jsonPathSegment, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil, fmt.Errorf("jsonPath expects non-empty path")
	}
	i := 0
	if path[i] == '$' {
		i++
	}
	segments := []jsonPathSegment{}
	for i < len(path) {
		if path[i] == '.' {
			i++
			if i >= len(path) {
				return nil, fmt.Errorf("expected segment after '.'")
			}
		}
		if path[i] == '[' {
			seg, next, err := parseJSONPathBracket(path, i)
			if err != nil {
				return nil, err
			}
			segments = append(segments, seg)
			i = next
			continue
		}
		seg, next, err := parseJSONPathIdent(path, i)
		if err != nil {
			return nil, err
		}
		segments = append(segments, seg)
		i = next
	}
	return segments, nil
}

func parseJSONPathIdent(path string, i int) (jsonPathSegment, int, error) {
	if i >= len(path) || !isIdentStart(path[i]) {
		return jsonPathSegment{}, i, fmt.Errorf("expected identifier at %d", i)
	}
	start := i
	i++
	for i < len(path) && isIdentPart(path[i]) {
		i++
	}
	return jsonPathSegment{kind: jsonPathKey, key: path[start:i]}, i, nil
}

func parseJSONPathBracket(path string, i int) (jsonPathSegment, int, error) {
	if path[i] != '[' {
		return jsonPathSegment{}, i, fmt.Errorf("expected '[' at %d", i)
	}
	i++
	i = skipSpaces(path, i)
	if i >= len(path) {
		return jsonPathSegment{}, i, fmt.Errorf("unterminated '['")
	}
	switch path[i] {
	case '\'', '"':
		key, next, err := parseJSONPathString(path, i)
		if err != nil {
			return jsonPathSegment{}, i, err
		}
		i = skipSpaces(path, next)
		if i >= len(path) || path[i] != ']' {
			return jsonPathSegment{}, i, fmt.Errorf("expected ']' after string")
		}
		return jsonPathSegment{kind: jsonPathKey, key: key}, i + 1, nil
	default:
		idx, next, err := parseJSONPathIndex(path, i)
		if err != nil {
			return jsonPathSegment{}, i, err
		}
		i = skipSpaces(path, next)
		if i >= len(path) || path[i] != ']' {
			return jsonPathSegment{}, i, fmt.Errorf("expected ']' after index")
		}
		return jsonPathSegment{kind: jsonPathIndex, index: idx}, i + 1, nil
	}
}

func parseJSONPathString(path string, i int) (string, int, error) {
	quote := path[i]
	i++
	var b strings.Builder
	for i < len(path) {
		ch := path[i]
		if ch == '\\' {
			if i+1 >= len(path) {
				return "", i, fmt.Errorf("unterminated escape")
			}
			i++
			esc := path[i]
			switch esc {
			case 'n':
				b.WriteByte('\n')
			case 'r':
				b.WriteByte('\r')
			case 't':
				b.WriteByte('\t')
			default:
				b.WriteByte(esc)
			}
			i++
			continue
		}
		if ch == quote {
			return b.String(), i + 1, nil
		}
		b.WriteByte(ch)
		i++
	}
	return "", i, fmt.Errorf("unterminated string")
}

func parseJSONPathIndex(path string, i int) (int, int, error) {
	sign := 1
	if path[i] == '-' {
		sign = -1
		i++
	}
	if i >= len(path) || path[i] < '0' || path[i] > '9' {
		return 0, i, fmt.Errorf("invalid index")
	}
	val := 0
	for i < len(path) && path[i] >= '0' && path[i] <= '9' {
		val = val*10 + int(path[i]-'0')
		i++
	}
	return sign * val, i, nil
}

func skipSpaces(path string, i int) int {
	for i < len(path) {
		switch path[i] {
		case ' ', '\t', '\n', '\r':
			i++
		default:
			return i
		}
	}
	return i
}

func isIdentStart(ch byte) bool {
	return ch == '_' || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isIdentPart(ch byte) bool {
	return isIdentStart(ch) || (ch >= '0' && ch <= '9')
}

func jsonPathLookup(value Value, segments []jsonPathSegment) (Value, bool) {
	current := value
	for _, seg := range segments {
		switch seg.kind {
		case jsonPathKey:
			switch v := current.(type) {
			case *Object:
				next, ok := v.Pairs[seg.key]
				if !ok {
					return nil, false
				}
				current = next
			case *ModuleObject:
				if v.Env == nil {
					return nil, false
				}
				next, ok := v.Env.Snapshot()[seg.key]
				if !ok {
					return nil, false
				}
				current = next
			case *Map:
				if next, ok := v.Pairs[MapKey{Type: STRING, Value: seg.key}]; ok {
					current = next
					continue
				}
				if len(seg.key) == 1 {
					if next, ok := v.Pairs[MapKey{Type: CHAR, Value: seg.key}]; ok {
						current = next
						continue
					}
				}
				return nil, false
			default:
				return nil, false
			}
		case jsonPathIndex:
			arr, ok := current.(*Array)
			if !ok {
				return nil, false
			}
			if seg.index < 0 || seg.index >= len(arr.Elements) {
				return nil, false
			}
			current = arr.Elements[seg.index]
		}
	}
	return current, true
}
