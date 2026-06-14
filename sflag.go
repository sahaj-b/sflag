package sflag

import (
	"flag"
	"os"
	"path/filepath"
)

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

// Parse registers flags from struct tags, parses os.Args[1:], and binds
// positional arguments to struct fields. target must be a pointer to a struct.
//
// Struct tags:
//   - flag:"name"       – long flag name (default: kebab-case of field name)
//   - short:"x"         – short flag (default: first char of long name, skipped on conflict)
//   - default:"val"     – default value as string (default: zero value)
//   - help:"text"       – help description
//   - positional:""     – marks field as positional arg (field order = arg order)
//
// Supported field types: string, int, int64, uint, uint64, bool, float64, time.Duration.
// Positional fields also support []string (must be last positional, captures remaining args).
func Parse(target any, opts ...Options) error {
	return ParseArgs(target, os.Args[1:], opts...)
}

// ParseArgs is Parse with explicit args instead of os.Args[1:].
func ParseArgs(target any, args []string, opts ...Options) error {
	var o Options
	if len(opts) > 0 {
		o = opts[0]
	}

	progName := o.ProgramName
	if progName == "" {
		progName = filepath.Base(os.Args[0])
	}
	useColor := !o.NoColor && os.Getenv("NO_COLOR") == ""

	fs := flag.NewFlagSet(progName, flag.ContinueOnError)
	fs.Usage = func() {}

	flags, positionals, err := bindStruct(fs, target, o)
	if err != nil {
		return err
	}

	err = fs.Parse(args)
	if err == flag.ErrHelp {
		setColors(useColor && isWriterTerminal(os.Stdout))
		showHelp(os.Stdout, progName, flags, positionals)
		os.Exit(0)
	}
	if err != nil {
		setColors(useColor && isWriterTerminal(os.Stderr))
		showHelp(os.Stderr, progName, flags, positionals)
		return err
	}
	if err := bindPositionals(positionals, fs.Args()); err != nil {
		setColors(useColor && isWriterTerminal(os.Stderr))
		showHelp(os.Stderr, progName, flags, positionals)
		return err
	}
	return nil
}
