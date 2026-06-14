package sflag

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"
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
	cBold = "\x1b[1m"
	cGreen = "\x1b[32m"
	cBlue = "\x1b[34m"
	cYellow = "\x1b[33m"
	cReset = "\x1b[0m"
}

func resetColors() {
	cBold = ""
	cGreen = ""
	cBlue = ""
	cYellow = ""
	cReset = ""
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
//   - default:"val" – default value as string
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

	if !o.NoColor && os.Getenv("NO_COLOR") == "" {
		initColors()
	}

	progName := o.ProgramName
	if progName == "" {
		progName = filepath.Base(os.Args[0])
	}

	fs := flag.NewFlagSet(progName, flag.ContinueOnError)
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

	fs.Usage = func() {} // we handle output ourselves

	err := fs.Parse(os.Args[1:])
	if err == flag.ErrHelp {
		if !o.NoColor && os.Getenv("NO_COLOR") == "" && isWriterTerminal(os.Stdout) {
			initColors()
		} else {
			resetColors()
		}
		showHelp(os.Stdout, progName, definedFlags)
		os.Exit(0)
	}
	if err != nil {
		if !o.NoColor && os.Getenv("NO_COLOR") == "" && isWriterTerminal(os.Stderr) {
			initColors()
		} else {
			resetColors()
		}
		showHelp(os.Stderr, progName, definedFlags)
	}
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

	case reflect.Int64:
		if field.Type == reflect.TypeOf(time.Duration(0)) {
			var defDur time.Duration
			if def != "" {
				defDur, _ = time.ParseDuration(def)
			}
			ptr := fieldVal.Addr().Interface().(*time.Duration)
			fs.DurationVar(ptr, longName, defDur, help)
			if shortName != "" {
				fs.DurationVar(ptr, shortName, defDur, "")
				clearShortHelp(fs, shortName)
			}
		} else {
			var defInt64 int64
			if def != "" {
				fmt.Sscanf(def, "%d", &defInt64)
			}
			ptr := fieldVal.Addr().Interface().(*int64)
			fs.Int64Var(ptr, longName, defInt64, help)
			if shortName != "" {
				fs.Int64Var(ptr, shortName, defInt64, "")
				clearShortHelp(fs, shortName)
			}
		}

	case reflect.Uint:
		var defUint uint64
		if def != "" {
			fmt.Sscanf(def, "%d", &defUint)
		}
		ptr := fieldVal.Addr().Interface().(*uint)
		fs.UintVar(ptr, longName, uint(defUint), help)
		if shortName != "" {
			fs.UintVar(ptr, shortName, uint(defUint), "")
			clearShortHelp(fs, shortName)
		}

	case reflect.Uint64:
		var defUint64 uint64
		if def != "" {
			fmt.Sscanf(def, "%d", &defUint64)
		}
		ptr := fieldVal.Addr().Interface().(*uint64)
		fs.Uint64Var(ptr, longName, defUint64, help)
		if shortName != "" {
			fs.Uint64Var(ptr, shortName, defUint64, "")
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
	case reflect.Int:
		return "int"
	case reflect.Int64:
		if t == reflect.TypeOf(time.Duration(0)) {
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

func showHelp(w io.Writer, prog string, flags []flagDef) {
	fmt.Fprintf(w, "%sUsage:%s %s [options] [args]\n\n", cBold, cReset, prog)
	fmt.Fprintf(w, "%sOptions:%s\n", cBold, cReset)

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
		fmt.Fprintf(w, "  %s%s%s\n", label, padding, helpStr)
	}

	hLabel := cGreen + "-h, --help" + cReset
	padding := strings.Repeat(" ", maxW-len("-h, --help"))
	fmt.Fprintf(w, "  %s%sDisplay help information\n", hLabel, padding)
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
