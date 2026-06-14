package sflag

import (
	"flag"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var durationType = reflect.TypeFor[time.Duration]()

type flagDef struct {
	long     string
	short    string
	typeName string
	defStr   string
	help     string
}

type positionalDef struct {
	field      reflect.StructField
	fieldVal   reflect.Value
	help       string
	defStr     string
	isVariadic bool
}

func bindStruct(fs *flag.FlagSet, target any, o Options) ([]flagDef, []positionalDef, error) {
	rv := reflect.ValueOf(target)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		panic("sflag: Parse target must be a pointer to a struct")
	}

	var flags []flagDef
	var positionals []positionalDef
	usedNames := make(map[string]bool)
	structVal := rv.Elem()
	structType := structVal.Type()
	seenPositional, closedPositionals := false, false

	for i := range structType.NumField() {
		field := structType.Field(i)
		fieldVal := structVal.Field(i)
		if _, ok := field.Tag.Lookup("positional"); ok {
			if closedPositionals {
				return nil, nil, fmt.Errorf("sflag: positional field %s must be grouped with other positional fields", field.Name)
			}
			if len(positionals) > 0 && positionals[len(positionals)-1].isVariadic {
				return nil, nil, fmt.Errorf("sflag: variadic positional field %s must be the last positional field", positionals[len(positionals)-1].field.Name)
			}
			pos, err := newPositional(field, fieldVal)
			if err != nil {
				return nil, nil, err
			}
			positionals = append(positionals, pos)
			seenPositional = true
			continue
		}

		if seenPositional {
			closedPositionals = true
		}

		longName := field.Tag.Get("flag")
		if longName == "" && !o.NoAutoName {
			longName = toKebab(field.Name)
		}
		if longName == "" {
			continue
		}
		if err := ensureExported(field, fieldVal); err != nil {
			return nil, nil, err
		}
		if usedNames[longName] {
			return nil, nil, fmt.Errorf("sflag: duplicate flag name %q", longName)
		}
		usedNames[longName] = true

		shortName, hasShort := field.Tag.Lookup("short")
		switch {
		case hasShort && shortName == "":
		case hasShort:
			if usedNames[shortName] {
				return nil, nil, fmt.Errorf("sflag: duplicate flag name %q", shortName)
			}
			usedNames[shortName] = true
		case !o.NoAutoShort:
			shortName = firstRune(longName)
			if usedNames[shortName] {
				shortName = ""
			} else {
				usedNames[shortName] = true
			}
		}

		info, err := registerField(fs, field, fieldVal, longName, shortName)
		if err != nil {
			return nil, nil, err
		}
		flags = append(flags, info)
	}

	return flags, positionals, nil
}

func newPositional(field reflect.StructField, fieldVal reflect.Value) (positionalDef, error) {
	if err := ensureExported(field, fieldVal); err != nil {
		return positionalDef{}, err
	}

	pos := positionalDef{field: field, fieldVal: fieldVal, help: field.Tag.Get("help"), defStr: field.Tag.Get("default")}
	if field.Type.Kind() == reflect.Slice {
		pos.defStr = ""
		if field.Type.Elem().Kind() != reflect.String {
			return pos, fmt.Errorf("sflag: variadic positional field %s must be []string, got %s", field.Name, field.Type)
		}
		pos.isVariadic = true
		return pos, nil
	}

	if _, ok := typeName(field.Type, true); !ok {
		return pos, fmt.Errorf("sflag: field %s has unsupported positional type %s", field.Name, field.Type)
	}
	return pos, nil
}

func ensureExported(field reflect.StructField, fieldVal reflect.Value) error {
	if !fieldVal.CanAddr() || !fieldVal.CanInterface() {
		return fmt.Errorf("sflag: field %s must be exported", field.Name)
	}
	return nil
}

func bindPositionals(positionals []positionalDef, args []string) error {
	if len(positionals) == 0 {
		return nil
	}

	required := len(positionals)
	hasVariadic := positionals[required-1].isVariadic
	if hasVariadic {
		required--
	}

	if !hasVariadic && len(args) > required {
		return fmt.Errorf("sflag: unexpected extra positional argument: %s", args[required])
	}

	for i, pos := range positionals {
		if pos.isVariadic {
			if i < len(args) {
				pos.fieldVal.Set(reflect.ValueOf(args[i:]))
			} else {
				pos.fieldVal.Set(reflect.ValueOf([]string{}))
			}
			return nil
		}
		if i < len(args) {
			if err := assign(pos.fieldVal, args[i]); err != nil {
				return positionalError(pos, err)
			}
		} else if pos.defStr != "" {
			if err := assign(pos.fieldVal, pos.defStr); err != nil {
				return positionalError(pos, err)
			}
		} else {
			return fmt.Errorf("sflag: missing positional argument: %s", strings.ToUpper(pos.field.Name))
		}
	}
	return nil
}

func positionalError(pos positionalDef, err error) error {
	return fmt.Errorf("sflag: positional argument %s: %w", strings.ToUpper(pos.field.Name), err)
}

func registerField(fs *flag.FlagSet, field reflect.StructField, fieldVal reflect.Value, longName, shortName string) (flagDef, error) {
	name, ok := typeName(field.Type, false)
	info := flagDef{long: longName, short: shortName, typeName: name, defStr: field.Tag.Get("default"), help: field.Tag.Get("help")}
	if !ok {
		return info, fmt.Errorf("sflag: field %s has unsupported type %s", field.Name, field.Type)
	}
	if err := assignDefault(fieldVal, info.defStr); err != nil {
		return info, fmt.Errorf("sflag: invalid default %q for field %s: %w", info.defStr, field.Name, err)
	}

	fs.Var(scalarValue{fieldVal}, longName, info.help)
	if shortName != "" {
		fs.Var(scalarValue{fieldVal}, shortName, "")
	}
	return info, nil
}

type scalarValue struct{ v reflect.Value }

func (s scalarValue) Set(raw string) error { return assign(s.v, raw) }
func (s scalarValue) String() string { return fmt.Sprint(s.v.Interface()) }
func (s scalarValue) IsBoolFlag() bool { return s.v.Kind() == reflect.Bool }

func assignDefault(v reflect.Value, raw string) error {
	if raw == "" {
		v.Set(reflect.Zero(v.Type()))
		return nil
	}
	return assign(v, raw)
}

func assign(v reflect.Value, raw string) error {
	switch v.Kind() {
	case reflect.String:
		v.SetString(raw)
	case reflect.Int, reflect.Int64:
		if v.Type() == durationType {
			d, err := time.ParseDuration(raw)
			if err != nil {
				return fmt.Errorf("invalid duration %q: %w", raw, err)
			}
			v.SetInt(int64(d))
			return nil
		}
		n, err := strconv.ParseInt(raw, 10, v.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid %s %q: %w", v.Kind(), raw, err)
		}
		v.SetInt(n)
	case reflect.Uint, reflect.Uint64:
		n, err := strconv.ParseUint(raw, 10, v.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid %s %q: %w", v.Kind(), raw, err)
		}
		v.SetUint(n)
	case reflect.Float64:
		n, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return fmt.Errorf("invalid float %q: %w", raw, err)
		}
		v.SetFloat(n)
	case reflect.Bool:
		b, err := strconv.ParseBool(raw)
		if err != nil {
			return fmt.Errorf("invalid bool %q: %w", raw, err)
		}
		v.SetBool(b)
	}
	return nil
}

func typeName(t reflect.Type, includeBool bool) (string, bool) {
	switch t.Kind() {
	case reflect.String:
		return "string", true
	case reflect.Int:
		return "int", true
	case reflect.Int64:
		if t == durationType {
			return "duration", true
		}
		return "int64", true
	case reflect.Uint:
		return "uint", true
	case reflect.Uint64:
		return "uint64", true
	case reflect.Float64:
		return "float", true
	case reflect.Bool:
		if includeBool {
			return "bool", true
		}
		return "", true
	default:
		return "", false
	}
}

func toKebab(s string) string {
	runes := []rune(s)
	out := make([]rune, 0, len(runes))
	for i, r := range runes {
		if i > 0 && unicode.IsUpper(r) {
			prev := runes[i-1]
			nextIsLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])
			if unicode.IsLower(prev) || (unicode.IsUpper(prev) && nextIsLower) {
				out = append(out, '-')
			}
		}
		out = append(out, unicode.ToLower(r))
	}
	return string(out)
}

func firstRune(s string) string {
	for _, r := range s {
		return string(r)
	}
	return ""
}
