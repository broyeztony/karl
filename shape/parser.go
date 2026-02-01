package shape

import (
	"fmt"
	"strconv"
	"strings"
)

// Parse parses a .shape file content into a single Shape.
// It fails if the file declares multiple shapes.
func Parse(content string) (*Shape, error) {
	file, err := ParseFile(content)
	if err != nil {
		return nil, err
	}
	if len(file.Shapes) != 1 {
		return nil, fmt.Errorf("shape parse error: expected exactly one shape")
	}
	return file.Shapes[0], nil
}

// ParseFile parses a .shape file content into a File that can contain
// multiple top-level shapes.
func ParseFile(content string) (*File, error) {
	lines := collectLines(content)
	if len(lines) == 0 {
		return nil, fmt.Errorf("shape parse error: empty file")
	}

	file := &File{
		Shapes:      []*Shape{},
		ByName:      map[string]*Shape{},
		Codecs:      []*CodecSpec{},
		ByCodecName: map[string]*CodecSpec{},
	}

	for i := 0; i < len(lines); {
		ln := lines[i]
		if ln.indent != 0 {
			return nil, lineError(ln, "unexpected indentation")
		}

		switch {
		case strings.HasPrefix(ln.text, "codec "):
			codec, err := parseCodecHeader(ln.text)
			if err != nil {
				return nil, lineError(ln, err.Error())
			}
			if _, exists := file.ByCodecName[codec.Name]; exists {
				return nil, lineError(ln, "duplicate codec name: "+codec.Name)
			}
			if _, exists := file.ByName[codec.Name]; exists {
				return nil, lineError(ln, "codec name conflicts with shape name: "+codec.Name)
			}
			if _, ok := file.ByName[codec.ShapeName]; !ok {
				return nil, lineError(ln, "unknown shape in codec: "+codec.ShapeName)
			}
			i++
			for i < len(lines) && lines[i].indent > 0 {
				mapLine := lines[i]
				if mapLine.indent != 4 {
					return nil, lineError(mapLine, "unexpected indentation")
				}
				mapping, err := parseCodecMapping(mapLine.text)
				if err != nil {
					return nil, lineError(mapLine, err.Error())
				}
				codec.Mappings = append(codec.Mappings, mapping)
				i++
			}
			file.Codecs = append(file.Codecs, codec)
			file.ByCodecName[codec.Name] = codec
		default:
			text := ln.text
			if strings.HasPrefix(text, "shape ") {
				text = strings.TrimSpace(strings.TrimPrefix(text, "shape "))
			}
			name, typeStr, err := splitDecl(text)
			if err != nil {
				return nil, lineError(ln, err.Error())
			}
			if !isIdent(name) {
				return nil, lineError(ln, "shape name must be a valid identifier")
			}
			if _, exists := file.ByName[name]; exists {
				return nil, lineError(ln, "duplicate shape name: "+name)
			}
			if _, exists := file.ByCodecName[name]; exists {
				return nil, lineError(ln, "shape name conflicts with codec name: "+name)
			}
			rootType, err := parseType(typeStr)
			if err != nil {
				return nil, lineError(ln, err.Error())
			}
			current := &Shape{Name: name, Type: rootType}
			file.Shapes = append(file.Shapes, current)
			file.ByName[name] = current

			contexts := []context{}
			if obj := ObjectType(rootType); obj != nil {
				contexts = append(contexts, context{indent: 4, target: obj})
			}
			i++
			for i < len(lines) && lines[i].indent > 0 {
				fieldLine := lines[i]
				if len(contexts) == 0 {
					return nil, lineError(fieldLine, "field declared outside of an object shape")
				}
				for len(contexts) > 0 && fieldLine.indent < contexts[len(contexts)-1].indent {
					contexts = contexts[:len(contexts)-1]
				}
				if len(contexts) == 0 || fieldLine.indent != contexts[len(contexts)-1].indent {
					return nil, lineError(fieldLine, "unexpected indentation")
				}
				ctx := contexts[len(contexts)-1]
				field, fieldType, err := parseField(fieldLine.text)
				if err != nil {
					return nil, lineError(fieldLine, err.Error())
				}
				field.Type = fieldType
				ctx.target.Fields = append(ctx.target.Fields, field)
				if obj := ObjectType(fieldType); obj != nil {
					contexts = append(contexts, context{indent: fieldLine.indent + 4, target: obj})
				}
				i++
			}
		}
	}

	if err := resolveFile(file); err != nil {
		return nil, err
	}
	return file, nil
}

type context struct {
	indent int
	target *Type
}

type parsedLine struct {
	num    int
	indent int
	text   string
}

func collectLines(content string) []parsedLine {
	lines := []parsedLine{}
	raw := strings.Split(content, "\n")
	for i, line := range raw {
		line = strings.TrimRight(line, "\r")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			continue
		}
		indent := countIndent(line)
		text := strings.TrimSpace(line)
		lines = append(lines, parsedLine{num: i + 1, indent: indent, text: text})
	}
	return lines
}

func countIndent(line string) int {
	count := 0
	for i := 0; i < len(line); i++ {
		switch line[i] {
		case ' ':
			count++
		case '\t':
			count += 4
		default:
			return count
		}
	}
	return count
}

func splitDecl(text string) (string, string, error) {
	parts := strings.SplitN(text, ":", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("expected ':' in declaration")
	}
	name := strings.TrimSpace(parts[0])
	typeStr := strings.TrimSpace(parts[1])
	if name == "" || typeStr == "" {
		return "", "", fmt.Errorf("invalid declaration")
	}
	return name, typeStr, nil
}

func parseCodecHeader(text string) (*CodecSpec, error) {
	parts := strings.Fields(text)
	if len(parts) != 3 || parts[0] != "codec" {
		return nil, fmt.Errorf("codec header must be: codec <Name> <ShapeName>")
	}
	name := parts[1]
	shapeName := parts[2]
	if !isIdent(name) {
		return nil, fmt.Errorf("codec name must be a valid identifier")
	}
	if !isIdent(shapeName) {
		return nil, fmt.Errorf("codec shape name must be a valid identifier")
	}
	return &CodecSpec{
		Name:      name,
		Format:    strings.ToLower(name),
		ShapeName: shapeName,
		Mappings:  []CodecMapping{},
	}, nil
}

func parseCodecMapping(text string) (CodecMapping, error) {
	type opSpec struct {
		op     string
		decode bool
		encode bool
	}
	ops := []opSpec{
		{op: "<-->", decode: true, encode: true},
		{op: "<->", decode: true, encode: true},
		{op: "<-", decode: true, encode: false},
		{op: "->", decode: false, encode: true},
	}
	for _, op := range ops {
		if idx := strings.Index(text, op.op); idx >= 0 {
			left := strings.TrimSpace(text[:idx])
			right := strings.TrimSpace(text[idx+len(op.op):])
			if left == "" || right == "" {
				return CodecMapping{}, fmt.Errorf("invalid codec mapping")
			}
			internal, err := parsePath(left, true)
			if err != nil {
				return CodecMapping{}, fmt.Errorf("invalid internal path: %v", err)
			}
			external, err := parsePath(right, false)
			if err != nil {
				return CodecMapping{}, fmt.Errorf("invalid external path: %v", err)
			}
			if len(internal) != len(external) {
				return CodecMapping{}, fmt.Errorf("codec mapping paths must have same depth")
			}
			return CodecMapping{
				InternalPath: internal,
				ExternalPath: external,
				Decode:       op.decode,
				Encode:       op.encode,
			}, nil
		}
	}
	return CodecMapping{}, fmt.Errorf("codec mapping must use one of: <--> <-> <- ->")
}

func parsePath(text string, internal bool) ([]string, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("empty path")
	}
	segments := []string{}
	for i := 0; i < len(text); {
		seg, next, err := parsePathSegment(text, i, internal)
		if err != nil {
			return nil, err
		}
		segments = append(segments, seg)
		i = next
		if i >= len(text) {
			break
		}
		if text[i] != '.' {
			return nil, fmt.Errorf("unexpected character %q", text[i])
		}
		i++
		if i >= len(text) {
			return nil, fmt.Errorf("path cannot end with '.'")
		}
	}
	return segments, nil
}

func parsePathSegment(text string, i int, internal bool) (string, int, error) {
	if i >= len(text) {
		return "", i, fmt.Errorf("missing segment")
	}
	if text[i] == '"' {
		j := i + 1
		for j < len(text) {
			if text[j] == '\\' {
				j += 2
				continue
			}
			if text[j] == '"' {
				raw := text[i : j+1]
				val, err := strconv.Unquote(raw)
				if err != nil {
					return "", i, fmt.Errorf("invalid quoted segment")
				}
				if internal && !isIdent(val) {
					return "", i, fmt.Errorf("internal segments must be identifiers")
				}
				return val, j + 1, nil
			}
			j++
		}
		return "", i, fmt.Errorf("unterminated quoted segment")
	}
	j := i
	for j < len(text) && text[j] != '.' {
		if text[j] == ' ' || text[j] == '\t' {
			return "", i, fmt.Errorf("spaces require quoted segments")
		}
		j++
	}
	seg := text[i:j]
	if seg == "" {
		return "", i, fmt.Errorf("empty segment")
	}
	if !isIdent(seg) {
		if internal {
			return "", i, fmt.Errorf("internal segments must be identifiers")
		}
		return "", i, fmt.Errorf("external non-identifier segments must be quoted")
	}
	return seg, j, nil
}

func parseField(text string) (Field, *Type, error) {
	left, typeStr, err := splitDecl(text)
	if err != nil {
		return Field{}, nil, err
	}
	required := true
	if strings.HasPrefix(left, "+") {
		left = strings.TrimSpace(left[1:])
		required = true
	} else if strings.HasPrefix(left, "-") {
		left = strings.TrimSpace(left[1:])
		required = false
	}

	name := left
	alias := ""
	if idx := strings.Index(left, "("); idx >= 0 {
		if !strings.HasSuffix(left, ")") {
			return Field{}, nil, fmt.Errorf("alias must end with ')' ")
		}
		name = strings.TrimSpace(left[:idx])
		alias = strings.TrimSpace(left[idx+1 : len(left)-1])
	}
	if name == "" || !isIdent(name) {
		return Field{}, nil, fmt.Errorf("field name must be a valid identifier")
	}
	fieldType, err := parseType(typeStr)
	if err != nil {
		return Field{}, nil, err
	}
	return Field{Name: name, Alias: alias, Required: required, Type: fieldType}, fieldType, nil
}

func parseType(text string) (*Type, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("missing type")
	}
	if strings.HasSuffix(text, "[]") {
		base := strings.TrimSpace(strings.TrimSuffix(text, "[]"))
		elem, err := parseType(base)
		if err != nil {
			return nil, err
		}
		return &Type{Kind: KindArray, Elem: elem}, nil
	}
	if strings.Contains(text, "|") {
		return nil, fmt.Errorf("union types are not supported yet")
	}
	switch text {
	case "string":
		return &Type{Kind: KindString}, nil
	case "int":
		return &Type{Kind: KindInt}, nil
	case "float":
		return &Type{Kind: KindFloat}, nil
	case "bool":
		return &Type{Kind: KindBool}, nil
	case "null":
		return &Type{Kind: KindNull}, nil
	case "any":
		return &Type{Kind: KindAny}, nil
	case "object":
		return &Type{Kind: KindObject, Fields: []Field{}}, nil
	default:
		if isIdent(text) {
			return &Type{Kind: KindRef, RefName: text}, nil
		}
		return nil, fmt.Errorf("unknown type: %s", text)
	}
}

func isIdent(name string) bool {
	if name == "" {
		return false
	}
	for i := 0; i < len(name); i++ {
		ch := name[i]
		if i == 0 {
			if !(ch == '_' || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')) {
				return false
			}
			continue
		}
		if !(ch == '_' || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9')) {
			return false
		}
	}
	return true
}

func lineError(line parsedLine, msg string) error {
	return fmt.Errorf("shape parse error at line %d: %s", line.num, msg)
}
