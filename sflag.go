package sflag

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"unicode"
)

var (
	cBold  string
	cGreen string
	cBlue  string
	cYellow string
	cReset string
	parsedArgs []string
)

func initColors() {
	if isTerminal() {
		cBold = "\x1b[1m"
		cGreen = "\x1b[32m"
		cBlue = "\x1b[34m"
		cYellow = "\x1b[33m"
		cReset = "\x1b[0m"
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

	// AutoName derives flag names from field names in kebab-case.
	// e.g., ApiKey → --api-key. Default: true.
	AutoName *bool

	// AutoShort derives short flags from the first char of the long name.
	// Skipped if it conflicts with an already-registered short. Default: true.
	AutoShort *bool

	// NoColor disables colored help output. Default: false (auto-detect).
	NoColor bool
}

func optBool(p *bool) bool {
	if p == nil {
		return true
	}
	return *p
}

// Parse registers flags from struct tags, parses os.Args[1:], and returns
// any remaining positional arguments. target must be a pointer to a struct.
//
// Struct tags:
//   - flag:"name"   – long flag name (default: kebab-case of field name)
//   - short:"x"     – short flag (default: first char of long name, skipped on conflict)
//   - default:"val" – default value as string
//   - help:"text"   – help description
//
// Supported field types: string, int, bool, float64.
func Parse(target any, opts ...Options) ([]string, error) {
	rv := reflect.ValueOf(target)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		panic("sflag: Parse target must be a pointer to a struct")
	}

	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}

	doAutoName := optBool(o.AutoName)
	doAutoShort := optBool(o.AutoShort)

	if !o.NoColor && os.Getenv("NO_COLOR") == "" && isTerminal() {
		initColors()
	}

	progName := o.ProgramName
	if progName == "" {
		progName = filepath.Base(os.Args[0])
	}

	fs := flag.NewFlagSet(progName, flag.ExitOnError)
	var definedFlags []flagDef
	usedShorts := make(map[rune]bool)

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

		shortName := field.Tag.Get("short")
		if shortName != "" {
			usedShorts[rune(shortName[0])] = true
		} else if doAutoShort {
			first := rune(longName[0])
			if !usedShorts[first] {
				shortName = string(first)
				usedShorts[first] = true
			}
		}

		help := field.Tag.Get("help")
		def := field.Tag.Get("default")

		info := registerField(fs, field, fieldVal, longName, shortName, def, help)
		definedFlags = append(definedFlags, info)
	}

	fs.Usage = func() {
		showHelp(fs, progName, definedFlags)
	}

	err := fs.Parse(os.Args[1:])
	parsedArgs = fs.Args()
	return parsedArgs, err
}

// Args returns the positional arguments from the last Parse call.
func Args() []string {
	return parsedArgs
}

func registerField(fs *flag.FlagSet, field reflect.StructField, fieldVal reflect.Value, longName, shortName, def, help string) flagDef {
	typeName := typeNameFor(field.Type)
	info := flagDef{long: longName, short: shortName, typeName: typeName, defStr: def, help: help}

	switch field.Type.Kind() {
	case reflect.String:
		ptr := fieldVal.Addr().Interface().(*string)
		fs.StringVar(ptr, longName, def, help)
		if shortName != "" {
			fs.StringVar(ptr, shortName, def, "")
			clearShortHelp(fs, shortName)
		}

	case reflect.Int:
		defInt := 0
		if def != "" {
			fmt.Sscanf(def, "%d", &defInt)
		}
		ptr := fieldVal.Addr().Interface().(*int)
		fs.IntVar(ptr, longName, defInt, help)
		if shortName != "" {
			fs.IntVar(ptr, shortName, defInt, "")
			clearShortHelp(fs, shortName)
		}

	case reflect.Bool:
		defBool := def == "true"
		ptr := fieldVal.Addr().Interface().(*bool)
		fs.BoolVar(ptr, longName, defBool, help)
		if shortName != "" {
			fs.BoolVar(ptr, shortName, defBool, "")
			clearShortHelp(fs, shortName)
		}

	case reflect.Float64:
		defFloat := 0.0
		if def != "" {
			fmt.Sscanf(def, "%f", &defFloat)
		}
		ptr := fieldVal.Addr().Interface().(*float64)
		fs.Float64Var(ptr, longName, defFloat, help)
		if shortName != "" {
			fs.Float64Var(ptr, shortName, defFloat, "")
			clearShortHelp(fs, shortName)
		}
	}

	return info
}

func clearShortHelp(fs *flag.FlagSet, name string) {
	if f := fs.Lookup(name); f != nil {
		f.Usage = ""
	}
}

func typeNameFor(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int64:
		return "int"
	case reflect.Float64:
		return "float"
	default:
		return ""
	}
}

func toKebab(s string) string {
	var out []rune
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			prev := rune(s[i-1])
			if unicode.IsLower(prev) ||
				(unicode.IsUpper(prev) && i+1 < len(s) && unicode.IsLower(rune(s[i+1]))) {
				out = append(out, '-')
			}
		}
		out = append(out, unicode.ToLower(r))
	}
	return string(out)
}

func showHelp(_ *flag.FlagSet, prog string, flags []flagDef) {
	fmt.Fprintf(os.Stderr, "%sUsage:%s %s [options] [args]\n\n", cBold, cReset, prog)
	fmt.Fprintf(os.Stderr, "%sOptions:%s\n", cBold, cReset)

	maxW := 0
	for _, f := range flags {
		w := len(flagLabelPlain(f))
		if f.typeName != "" {
			w += 1 + len(f.typeName)
		}
		if w > maxW {
			maxW = w
		}
	}
	maxW += 2

	for _, f := range flags {
		label := flagLabelColored(f)
		if f.typeName != "" {
			label += " " + cBlue + f.typeName + cReset
		}
		plainLen := len(flagLabelPlain(f))
		if f.typeName != "" {
			plainLen += 1 + len(f.typeName)
		}
		padding := strings.Repeat(" ", maxW-plainLen)
		helpStr := f.help
		if f.defStr != "" {
			helpStr += " " + cYellow + "(default: " + f.defStr + ")" + cReset
		}
		fmt.Fprintf(os.Stderr, "  %s%s%s\n", label, padding, helpStr)
	}

	hLabel := cGreen + "-h, --help" + cReset
	padding := strings.Repeat(" ", maxW-len("-h, --help"))
	fmt.Fprintf(os.Stderr, "  %s%sDisplay help information\n", hLabel, padding)
}

func flagLabelPlain(f flagDef) string {
	if f.short != "" {
		return fmt.Sprintf("-%s, --%s", f.short, f.long)
	}
	return fmt.Sprintf("    --%s", f.long)
}

func flagLabelColored(f flagDef) string {
	if f.short != "" {
		return cGreen + fmt.Sprintf("-%s, --%s", f.short, f.long) + cReset
	}
	return cGreen + fmt.Sprintf("    --%s", f.long) + cReset
}
