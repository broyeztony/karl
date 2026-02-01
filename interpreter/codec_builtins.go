package interpreter

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"karl/shape"
)

func builtinDecode(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "decode expects 2 arguments: text, codec"}
	}
	text, ok := args[0].(*String)
	if !ok {
		return nil, &RuntimeError{Message: "decode expects string as first argument"}
	}
	codec, ok := args[1].(*CodecValue)
	if !ok {
		return nil, &RuntimeError{Message: "decode expects codec as second argument"}
	}

	var val Value
	var err error
	switch codec.Format {
	case "json":
		val, err = builtinDecodeJSON(nil, []Value{text})
	case "yaml":
		val, err = decodeYAMLValue(text.Value)
	default:
		return nil, &RuntimeError{Message: "unknown codec format: " + codec.Format}
	}
	if err != nil {
		return nil, err
	}
	return applyCodecMappings(val, codec, true)
}

func builtinJsonCodec(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 0 {
		return nil, &RuntimeError{Message: "JsonCodec expects no arguments"}
	}
	return &CodecValue{Name: "Json", Format: "json"}, nil
}

func builtinYamlCodec(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 0 {
		return nil, &RuntimeError{Message: "YamlCodec expects no arguments"}
	}
	return &CodecValue{Name: "Yaml", Format: "yaml"}, nil
}

func builtinEncode(_ *Evaluator, args []Value) (Value, error) {
	if len(args) != 2 {
		return nil, &RuntimeError{Message: "encode expects 2 arguments: value, codec"}
	}
	codec, ok := args[1].(*CodecValue)
	if !ok {
		return nil, &RuntimeError{Message: "encode expects codec as second argument"}
	}
	mapped, err := applyCodecMappings(args[0], codec, false)
	if err != nil {
		return nil, err
	}

	switch codec.Format {
	case "json":
		return builtinEncodeJSON(nil, []Value{mapped})
	case "yaml":
		out, err := encodeYAML(mapped, 0)
		if err != nil {
			return nil, err
		}
		return &String{Value: strings.TrimSuffix(out, "\n")}, nil
	default:
		return nil, &RuntimeError{Message: "unknown codec format: " + codec.Format}
	}
}

func applyCodecMappings(value Value, codec *CodecValue, decode bool) (Value, error) {
	if codec == nil || len(codec.Mappings) == 0 {
		if decode {
			return value, nil
		}
		return cloneValue(value), nil
	}
	target := value
	if !decode {
		target = cloneValue(value)
	}
	if codec.Shape != nil && codec.Shape.Kind == shape.KindArray {
		arr, ok := target.(*Array)
		if !ok {
			return nil, recoverableError("codec", "codec expects array root")
		}
		for _, el := range arr.Elements {
			if err := applyCodecMappingsToObject(el, codec.Mappings, decode); err != nil {
				return nil, err
			}
		}
		return target, nil
	}
	if err := applyCodecMappingsToObject(target, codec.Mappings, decode); err != nil {
		return nil, err
	}
	return target, nil
}

func applyCodecMappingsToObject(root Value, mappings []shape.CodecMapping, decode bool) error {
	for _, mapping := range mappings {
		if decode && !mapping.Decode {
			continue
		}
		if !decode && !mapping.Encode {
			continue
		}
		from := mapping.InternalPath
		to := mapping.ExternalPath
		if decode {
			from = mapping.ExternalPath
			to = mapping.InternalPath
		}
		if err := renamePath(root, from, to); err != nil {
			return err
		}
	}
	return nil
}

func renamePath(root Value, fromPath []string, toPath []string) error {
	cur := root
	for i := 0; i < len(fromPath)-1; i++ {
		obj, ok := cur.(*Object)
		if !ok {
			return nil
		}
		next, ok := obj.Pairs[fromPath[i]]
		if !ok {
			return nil
		}
		cur = next
	}
	obj, ok := cur.(*Object)
	if !ok {
		return nil
	}
	fromKey := fromPath[len(fromPath)-1]
	toKey := toPath[len(toPath)-1]
	val, ok := obj.Pairs[fromKey]
	if !ok {
		return nil
	}
	if existing, exists := obj.Pairs[toKey]; exists && toKey != fromKey {
		_ = existing
		return recoverableError("codec", fmt.Sprintf("codec key collision at %q", strings.Join(toPath, ".")))
	}
	obj.Pairs[toKey] = val
	if toKey != fromKey {
		delete(obj.Pairs, fromKey)
	}
	return nil
}

func cloneValue(value Value) Value {
	switch v := value.(type) {
	case *Object:
		pairs := make(map[string]Value, len(v.Pairs))
		for k, item := range v.Pairs {
			pairs[k] = cloneValue(item)
		}
		return &Object{Pairs: pairs, Shape: v.Shape}
	case *Array:
		out := make([]Value, 0, len(v.Elements))
		for _, el := range v.Elements {
			out = append(out, cloneValue(el))
		}
		return &Array{Elements: out}
	case *Map:
		out := make(map[MapKey]Value, len(v.Pairs))
		for k, val := range v.Pairs {
			out[k] = cloneValue(val)
		}
		return &Map{Pairs: out}
	default:
		return value
	}
}

func encodeYAML(value Value, indent int) (string, error) {
	pad := strings.Repeat(" ", indent)
	switch v := value.(type) {
	case *Object:
		if len(v.Pairs) == 0 {
			return "{}\n", nil
		}
		keys := make([]string, 0, len(v.Pairs))
		for k := range v.Pairs {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var b strings.Builder
		for _, k := range keys {
			val := v.Pairs[k]
			scalar, ok, err := yamlScalar(val)
			if err != nil {
				return "", err
			}
			key := yamlKey(k)
			if ok {
				b.WriteString(pad)
				b.WriteString(key)
				b.WriteString(": ")
				b.WriteString(scalar)
				b.WriteString("\n")
				continue
			}
			nested, err := encodeYAML(val, indent+2)
			if err != nil {
				return "", err
			}
			b.WriteString(pad)
			b.WriteString(key)
			b.WriteString(":\n")
			b.WriteString(nested)
		}
		return b.String(), nil
	case *Array:
		var b strings.Builder
		for _, el := range v.Elements {
			scalar, ok, err := yamlScalar(el)
			if err != nil {
				return "", err
			}
			if ok {
				b.WriteString(pad)
				b.WriteString("- ")
				b.WriteString(scalar)
				b.WriteString("\n")
				continue
			}
			nested, err := encodeYAML(el, indent+2)
			if err != nil {
				return "", err
			}
			b.WriteString(pad)
			b.WriteString("-\n")
			b.WriteString(nested)
		}
		return b.String(), nil
	default:
		scalar, ok, err := yamlScalar(v)
		if err != nil {
			return "", err
		}
		if !ok {
			return "", &RuntimeError{Message: "yaml encoder cannot encode value"}
		}
		return pad + scalar + "\n", nil
	}
}

func yamlScalar(value Value) (string, bool, error) {
	switch v := value.(type) {
	case *String:
		return strconv.Quote(v.Value), true, nil
	case *Char:
		return strconv.Quote(v.Value), true, nil
	case *Integer:
		return fmt.Sprintf("%d", v.Value), true, nil
	case *Float:
		return fmt.Sprintf("%g", v.Value), true, nil
	case *Boolean:
		if v.Value {
			return "true", true, nil
		}
		return "false", true, nil
	case *Null:
		return "null", true, nil
	case *Unit:
		return "null", true, nil
	default:
		return "", false, nil
	}
}

func yamlKey(key string) string {
	if key == "" {
		return strconv.Quote(key)
	}
	for i := 0; i < len(key); i++ {
		ch := key[i]
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' || ch == '-' {
			continue
		}
		return strconv.Quote(key)
	}
	return key
}

type yamlLine struct {
	line   int
	indent int
	text   string
}

func decodeYAMLValue(input string) (Value, error) {
	lines := collectYAMLLines(input)
	if len(lines) == 0 {
		return &Object{Pairs: map[string]Value{}}, nil
	}
	val, next, err := parseYAMLBlock(lines, 0, lines[0].indent)
	if err != nil {
		return nil, recoverableError("decode", "yaml decode error: "+err.Error())
	}
	if next != len(lines) {
		return nil, recoverableError("decode", "yaml decode error: unexpected trailing content")
	}
	return val, nil
}

func collectYAMLLines(input string) []yamlLine {
	raw := strings.Split(input, "\n")
	out := make([]yamlLine, 0, len(raw))
	for i, line := range raw {
		line = strings.TrimRight(line, "\r")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		indent := 0
		for indent < len(line) && line[indent] == ' ' {
			indent++
		}
		out = append(out, yamlLine{line: i + 1, indent: indent, text: strings.TrimSpace(line)})
	}
	return out
}

func parseYAMLBlock(lines []yamlLine, idx int, indent int) (Value, int, error) {
	if idx >= len(lines) {
		return nil, idx, fmt.Errorf("unexpected end of input")
	}
	if lines[idx].indent < indent {
		return nil, idx, fmt.Errorf("unexpected indentation on line %d", lines[idx].line)
	}
	if strings.HasPrefix(lines[idx].text, "-") {
		return parseYAMLArray(lines, idx, indent)
	}
	return parseYAMLObject(lines, idx, indent)
}

func parseYAMLObject(lines []yamlLine, idx int, indent int) (Value, int, error) {
	pairs := map[string]Value{}
	for idx < len(lines) {
		ln := lines[idx]
		if ln.indent < indent {
			break
		}
		if ln.indent > indent {
			return nil, idx, fmt.Errorf("unexpected indentation on line %d", ln.line)
		}
		if strings.HasPrefix(ln.text, "-") {
			return nil, idx, fmt.Errorf("mixed mapping and sequence on line %d", ln.line)
		}
		key, raw, hasValue, err := parseYAMLKeyValue(ln.text)
		if err != nil {
			return nil, idx, fmt.Errorf("line %d: %v", ln.line, err)
		}
		if hasValue {
			scalar, err := parseYAMLScalar(raw)
			if err != nil {
				return nil, idx, fmt.Errorf("line %d: %v", ln.line, err)
			}
			pairs[key] = scalar
			idx++
			continue
		}
		idx++
		if idx >= len(lines) || lines[idx].indent <= indent {
			return nil, idx, fmt.Errorf("line %d: expected nested block for key %q", ln.line, key)
		}
		childIndent := lines[idx].indent
		nested, next, err := parseYAMLBlock(lines, idx, childIndent)
		if err != nil {
			return nil, idx, err
		}
		pairs[key] = nested
		idx = next
	}
	return &Object{Pairs: pairs}, idx, nil
}

func parseYAMLArray(lines []yamlLine, idx int, indent int) (Value, int, error) {
	items := make([]Value, 0)
	for idx < len(lines) {
		ln := lines[idx]
		if ln.indent < indent {
			break
		}
		if ln.indent > indent {
			return nil, idx, fmt.Errorf("unexpected indentation on line %d", ln.line)
		}
		if !strings.HasPrefix(ln.text, "-") {
			return nil, idx, fmt.Errorf("mixed sequence and mapping on line %d", ln.line)
		}
		item := strings.TrimSpace(strings.TrimPrefix(ln.text, "-"))
		if item == "" {
			idx++
			if idx >= len(lines) || lines[idx].indent <= indent {
				return nil, idx, fmt.Errorf("line %d: expected nested block for sequence item", ln.line)
			}
			childIndent := lines[idx].indent
			nested, next, err := parseYAMLBlock(lines, idx, childIndent)
			if err != nil {
				return nil, idx, err
			}
			items = append(items, nested)
			idx = next
			continue
		}

		if key, raw, hasValue, err := parseYAMLKeyValue(item); err == nil && hasValue {
			val, err := parseYAMLScalar(raw)
			if err != nil {
				return nil, idx, fmt.Errorf("line %d: %v", ln.line, err)
			}
			items = append(items, &Object{Pairs: map[string]Value{key: val}})
			idx++
			continue
		}

		val, err := parseYAMLScalar(item)
		if err != nil {
			return nil, idx, fmt.Errorf("line %d: %v", ln.line, err)
		}
		items = append(items, val)
		idx++
	}
	return &Array{Elements: items}, idx, nil
}

func parseYAMLKeyValue(text string) (key string, raw string, hasValue bool, err error) {
	idx := strings.Index(text, ":")
	if idx < 0 {
		return "", "", false, fmt.Errorf("expected ':'")
	}
	keyPart := strings.TrimSpace(text[:idx])
	if keyPart == "" {
		return "", "", false, fmt.Errorf("empty key")
	}
	if strings.HasPrefix(keyPart, "\"") {
		k, err := strconv.Unquote(keyPart)
		if err != nil {
			return "", "", false, fmt.Errorf("invalid quoted key")
		}
		keyPart = k
	}
	rawPart := strings.TrimSpace(text[idx+1:])
	if rawPart == "" {
		return keyPart, "", false, nil
	}
	return keyPart, rawPart, true, nil
}

func parseYAMLScalar(raw string) (Value, error) {
	switch raw {
	case "null":
		return NullValue, nil
	case "true":
		return &Boolean{Value: true}, nil
	case "false":
		return &Boolean{Value: false}, nil
	}
	if strings.HasPrefix(raw, "\"") {
		str, err := strconv.Unquote(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid string literal")
		}
		return &String{Value: str}, nil
	}
	if i, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return &Integer{Value: i}, nil
	}
	if strings.ContainsAny(raw, ".eE") {
		if f, err := strconv.ParseFloat(raw, 64); err == nil {
			return &Float{Value: f}, nil
		}
	}
	return &String{Value: raw}, nil
}
