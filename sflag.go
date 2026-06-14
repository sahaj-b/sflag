package sflag

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var (
	cBold   string
	cGreen  string
	cBlue   string
	cYellow string
	cReset  string
)

func setColors(on bool) {
	if on {
		cBold = "\x1b[1m"
		cGreen = "\x1b[32m"
		cBlue = "\x1b[34m"
		cYellow = "\x1b[33m"
		cReset = "\x1b[0m"
	} else {
		cBold = ""
		cGreen = ""
		cBlue = ""
		cYellow = ""
		cReset = ""
	}
}

type flagDef struct {
	long     string
	short    string
	typeName string
	defStr   string
	help     string
}

// Options configures flag parsing behavior.
type Options struct {
	// ProgramName is shown in the usage message. Default: os.Args[0].
	ProgramName string

	// NoAutoName disables kebab-case name derivation from field names.
	// e.g., without it ApiKey → --api-key. Default: false.
	NoAutoName bool

	// NoAutoShort disables short flag derivation from first char.
	// Skipped if it conflicts with an already-registered short. Default: false.
	NoAutoShort bool

	// NoColor disables colored help output. Default: false (auto-detect).
	NoColor bool
}

// Parse registers flags from struct tags, parses os.Args[1:], and returns
// any remaining positional arguments. target must be a pointer to a struct.
//
// Struct tags:
//   - flag:"name"   – long flag name (default: kebab-case of field name)
//   - short:"x"     – short flag (default: first char of long name, skipped on conflict)
//   - default:"val" – default value as string (default: zero value)
//   - help:"text"   – help description
//
// Supported field types: string, int, int64, uint, uint64, bool, float64, time.Duration.
func Parse(target any, opts ...Options) ([]string, error) {
	rv := reflect.ValueOf(target)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		panic("sflag: Parse target must be a pointer to a struct")
	}

	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}

	doAutoName := !o.NoAutoName
	doAutoShort := !o.NoAutoShort

	useColor := !o.NoColor && os.Getenv("NO_COLOR") == ""

	progName := o.ProgramName
	if progName == "" {
		progName = filepath.Base(os.Args[0])
	}

	fs := flag.NewFlagSet(progName, flag.ContinueOnError)
	var definedFlags []flagDef
	usedNames := make(map[string]bool)

	structVal := rv.Elem()
	structType := structVal.Type()

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldVal := structVal.Field(i)

		longName := field.Tag.Get("flag")
		if longName == "" && doAutoName {
			longName = toKebab(field.Name)
		}
		if longName == "" {
			continue
		}
		if !fieldVal.CanAddr() || !fieldVal.CanInterface() {
			return nil, fmt.Errorf("sflag: field %s must be exported", field.Name)
		}
		if usedNames[longName] {
			return nil, fmt.Errorf("sflag: duplicate flag name %q", longName)
		}
		usedNames[longName] = true

		shortName, hasShort := field.Tag.Lookup("short")
		if hasShort {
			if shortName == "" {
				// explicit `short:""` → no short flag
			} else if usedNames[shortName] {
				return nil, fmt.Errorf("sflag: duplicate flag name %q", shortName)
			} else {
				usedNames[shortName] = true
			}
		} else if doAutoShort {
			first := firstRune(longName)
			if !usedNames[first] {
				shortName = first
				usedNames[first] = true
			}
		}

		help := field.Tag.Get("help")
		def := field.Tag.Get("default")

		info, err := registerField(fs, field, fieldVal, longName, shortName, def, help)
		if err != nil {
			return nil, err
		}
		definedFlags = append(definedFlags, info)
	}

	fs.Usage = func() {} // we handle output ourselves

	err := fs.Parse(os.Args[1:])
	if err == flag.ErrHelp {
		setColors(useColor && isWriterTerminal(os.Stdout))
		showHelp(os.Stdout, progName, definedFlags)
		os.Exit(0)
	}
	if err != nil {
		setColors(useColor && isWriterTerminal(os.Stderr))
		showHelp(os.Stderr, progName, definedFlags)
	}
	return fs.Args(), err
}

func registerField(fs *flag.FlagSet, field reflect.StructField, fieldVal reflect.Value, longName, shortName, def, help string) (flagDef, error) {
	typeName := typeNameFor(field.Type)
	info := flagDef{long: longName, short: shortName, typeName: typeName, defStr: def, help: help}

	switch field.Type.Kind() {
	case reflect.String:
		ptr := fieldVal.Addr().Interface().(*string)
		fs.StringVar(ptr, longName, def, help)
		registerShort(fs, shortName, func() { fs.StringVar(ptr, shortName, def, "") })

	case reflect.Int:
		defInt, err := parseDefault(def, strconv.Atoi)
		if err != nil {
			return info, defaultError(field.Name, def, err)
		}
		ptr := fieldVal.Addr().Interface().(*int)
		fs.IntVar(ptr, longName, defInt, help)
		registerShort(fs, shortName, func() { fs.IntVar(ptr, shortName, defInt, "") })

	case reflect.Int64:
		if field.Type == reflect.TypeFor[time.Duration]() {
			defDur, err := parseDefault(def, time.ParseDuration)
			if err != nil {
				return info, defaultError(field.Name, def, err)
			}
			ptr := fieldVal.Addr().Interface().(*time.Duration)
			fs.DurationVar(ptr, longName, defDur, help)
			registerShort(fs, shortName, func() { fs.DurationVar(ptr, shortName, defDur, "") })
		} else {
			defInt64, err := parseDefault(def, func(s string) (int64, error) { return strconv.ParseInt(s, 10, 64) })
			if err != nil {
				return info, defaultError(field.Name, def, err)
			}
			ptr := fieldVal.Addr().Interface().(*int64)
			fs.Int64Var(ptr, longName, defInt64, help)
			registerShort(fs, shortName, func() { fs.Int64Var(ptr, shortName, defInt64, "") })
		}

	case reflect.Uint:
		defUint64, err := parseDefault(def, func(s string) (uint64, error) { return strconv.ParseUint(s, 10, 0) })
		if err != nil {
			return info, defaultError(field.Name, def, err)
		}
		ptr := fieldVal.Addr().Interface().(*uint)
		fs.UintVar(ptr, longName, uint(defUint64), help)
		registerShort(fs, shortName, func() { fs.UintVar(ptr, shortName, uint(defUint64), "") })

	case reflect.Uint64:
		defUint64, err := parseDefault(def, func(s string) (uint64, error) { return strconv.ParseUint(s, 10, 64) })
		if err != nil {
			return info, defaultError(field.Name, def, err)
		}
		ptr := fieldVal.Addr().Interface().(*uint64)
		fs.Uint64Var(ptr, longName, defUint64, help)
		registerShort(fs, shortName, func() { fs.Uint64Var(ptr, shortName, defUint64, "") })

	case reflect.Bool:
		defBool, err := parseDefault(def, strconv.ParseBool)
		if err != nil {
			return info, defaultError(field.Name, def, err)
		}
		ptr := fieldVal.Addr().Interface().(*bool)
		fs.BoolVar(ptr, longName, defBool, help)
		registerShort(fs, shortName, func() { fs.BoolVar(ptr, shortName, defBool, "") })

	case reflect.Float64:
		defFloat, err := parseDefault(def, func(s string) (float64, error) { return strconv.ParseFloat(s, 64) })
		if err != nil {
			return info, defaultError(field.Name, def, err)
		}
		ptr := fieldVal.Addr().Interface().(*float64)
		fs.Float64Var(ptr, longName, defFloat, help)
		registerShort(fs, shortName, func() { fs.Float64Var(ptr, shortName, defFloat, "") })

	default:
		return info, fmt.Errorf("sflag: field %s has unsupported type %s", field.Name, field.Type)
	}

	return info, nil
}

func registerShort(fs *flag.FlagSet, name string, register func()) {
	if name == "" {
		return
	}
	register()
	if f := fs.Lookup(name); f != nil {
		f.Usage = ""
	}
}

func parseDefault[T any](def string, parse func(string) (T, error)) (T, error) {
	var zero T
	if def == "" {
		return zero, nil
	}
	return parse(def)
}

func defaultError(fieldName, def string, err error) error {
	return fmt.Errorf("sflag: invalid default %q for field %s: %w", def, fieldName, err)
}

func typeNameFor(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int:
		return "int"
	case reflect.Int64:
		if t == reflect.TypeFor[time.Duration]() {
			return "duration"
		}
		return "int64"
	case reflect.Uint:
		return "uint"
	case reflect.Uint64:
		return "uint64"
	case reflect.Float64:
		return "float"
	default:
		return ""
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

func showHelp(w io.Writer, prog string, flags []flagDef) {
	fmt.Fprintf(w, "%sUsage:%s %s [options] [args]\n\n", cBold, cReset, prog)
	fmt.Fprintf(w, "%sOptions:%s\n", cBold, cReset)

	maxW := 0
	for _, f := range flags {
		label := flagLabel(f)
		if f.typeName != "" {
			label += " " + f.typeName
		}
		if n := len(stripAnsi(label)); n > maxW {
			maxW = n
		}
	}
	maxW += 2

	for _, f := range flags {
		label := flagLabel(f)
		if f.typeName != "" {
			label += " " + cBlue + f.typeName + cReset
		}
		plainLen := len(stripAnsi(label))
		padding := strings.Repeat(" ", maxW-plainLen)
		helpStr := f.help
		if f.defStr != "" {
			helpStr += " " + cYellow + "(default: " + f.defStr + ")" + cReset
		}
		fmt.Fprintf(w, "  %s%s%s\n", label, padding, helpStr)
	}

	hLabel := cGreen + "-h, --help" + cReset
	padding := strings.Repeat(" ", maxW-len(stripAnsi(hLabel)))
	fmt.Fprintf(w, "  %s%sDisplay help information\n", hLabel, padding)
}

func flagLabel(f flagDef) string {
	if f.short != "" {
		return cGreen + fmt.Sprintf("-%s, --%s", f.short, f.long) + cReset
	}
	return cGreen + fmt.Sprintf("    --%s", f.long) + cReset
}

func stripAnsi(s string) string {
	var out strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			for i < len(s) && s[i] != 'm' {
				i++
			}
			continue
		}
		out.WriteByte(s[i])
	}
	return out.String()
}
