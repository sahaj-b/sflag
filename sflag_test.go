package sflag

import (
	"os"
	"testing"
	"time"
)

type TestConfig struct {
	Range  string  `flag:"range" short:"r" default:"7d" help:"Range of data to fetch"`
	Days   int     `flag:"days" short:"d" default:"0" help:"Number of days to fetch"`
	Full   bool    `flag:"full" short:"f" help:"Display full statistics"`
	Rate   float64 `flag:"rate" short:"R" default:"1.5" help:"Some rate"`
	Output string  `flag:"output" default:"json" help:"Output format"`
}

func TestParseDefaults(t *testing.T) {
	os.Args = []string{"test"}

	var cfg TestConfig
	err := Parse(&cfg)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Range != "7d" {
		t.Errorf("Range: got %q, want %q", cfg.Range, "7d")
	}
	if cfg.Days != 0 {
		t.Errorf("Days: got %d, want %d", cfg.Days, 0)
	}
	if cfg.Full != false {
		t.Errorf("Full: got %v, want %v", cfg.Full, false)
	}
	if cfg.Rate != 1.5 {
		t.Errorf("Rate: got %f, want %f", cfg.Rate, 1.5)
	}
	if cfg.Output != "json" {
		t.Errorf("Output: got %q, want %q", cfg.Output, "json")
	}
}

func TestParseWithFlags(t *testing.T) {
	os.Args = []string{"test", "--range", "30d", "-d", "5", "--full", "-R", "2.0", "--output", "csv", "file1.txt", "file2.txt"}

	var cfg TestConfig
	err := Parse(&cfg)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Range != "30d" {
		t.Errorf("Range: got %q, want %q", cfg.Range, "30d")
	}
	if cfg.Days != 5 {
		t.Errorf("Days: got %d, want %d", cfg.Days, 5)
	}
	if cfg.Full != true {
		t.Errorf("Full: got %v, want %v", cfg.Full, true)
	}
	if cfg.Rate != 2.0 {
		t.Errorf("Rate: got %f, want %f", cfg.Rate, 2.0)
	}
	if cfg.Output != "csv" {
		t.Errorf("Output: got %q, want %q", cfg.Output, "csv")
	}
}

func TestParseBoolVariants(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{"long equal true", []string{"test", "--full=true"}, true},
		{"long equal false", []string{"test", "--full=false"}, false},
		{"short standalone", []string{"test", "-f"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args
			var cfg TestConfig
			_ = Parse(&cfg)
			if cfg.Full != tt.want {
				t.Errorf("Full: got %v, want %v", cfg.Full, tt.want)
			}
		})
	}
}

type MinimalConfig struct {
	Name string `flag:"name" short:"n" default:"world" help:"Name to greet"`
}

func TestExtraArgsIgnored(t *testing.T) {
	os.Args = []string{"test", "-n", "alice", "arg1", "arg2", "arg3"}

	var cfg MinimalConfig
	err := Parse(&cfg)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Name != "alice" {
		t.Errorf("Name: got %q, want %q", cfg.Name, "alice")
	}
}

func TestDoubleDash(t *testing.T) {
	os.Args = []string{"test", "--name", "bob", "--", "--not-a-flag"}

	var cfg MinimalConfig
	err := Parse(&cfg)
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Name != "bob" {
		t.Errorf("Name: got %q, want %q", cfg.Name, "bob")
	}
}

type EmptyConfig struct{}

func TestEmptyStruct(t *testing.T) {
	os.Args = []string{"test", "just", "args"}

	var cfg EmptyConfig
	err := Parse(&cfg)
	if err != nil {
		t.Fatal(err)
	}
}

func TestProgramName(t *testing.T) {
	os.Args = []string{"myapp"}

	var cfg MinimalConfig
	_ = Parse(&cfg, Options{ProgramName: "custom"})

	if cfg.Name != "world" {
		t.Errorf("Name: got %q, want default %q", cfg.Name, "world")
	}
}

type AutoNameConfig struct {
	Range  string `short:"r" default:"7d" help:"Range of data"`
	Days   int    `short:"d" default:"0" help:"Number of days"`
	Full   bool   `short:"f" help:"Display full stats"`
	Output string `default:"text" help:"Output format"`
}

func TestAutoNameFromFieldName(t *testing.T) {
	os.Args = []string{"test", "--range", "30d", "-d", "5", "--full", "--output", "json"}

	var cfg AutoNameConfig
	_ = Parse(&cfg)

	if cfg.Range != "30d" {
		t.Errorf("Range: got %q, want %q", cfg.Range, "30d")
	}
	if cfg.Days != 5 {
		t.Errorf("Days: got %d, want %d", cfg.Days, 5)
	}
	if cfg.Full != true {
		t.Errorf("Full: got %v, want %v", cfg.Full, true)
	}
	if cfg.Output != "json" {
		t.Errorf("Output: got %q, want %q", cfg.Output, "json")
	}
}

type KebabAutoShortConfig struct {
	Range   string `short:"r" default:"7d" help:"Range"`
	Rate    string `help:"Rate"`                // auto short: r, conflicts with Range, skipped
	Retries int    `default:"3" help:"Retries"` // auto short: r, also skipped
	Output  string `short:"o" default:"text" help:"Output"`
}

func TestShortConflictSkips(t *testing.T) {
	os.Args = []string{"test", "--range", "a", "--rate", "b", "--retries", "5", "-o", "x"}

	var cfg KebabAutoShortConfig
	_ = Parse(&cfg)

	if cfg.Range != "a" {
		t.Errorf("Range: got %q", cfg.Range)
	}
	if cfg.Rate != "b" {
		t.Errorf("Rate: got %q", cfg.Rate)
	}
	if cfg.Retries != 5 {
		t.Errorf("Retries: got %d", cfg.Retries)
	}
	if cfg.Output != "x" {
		t.Errorf("Output: got %q", cfg.Output)
	}
}

type NoAutoConfig struct {
	Range string `flag:"range" short:"r" default:"7d" help:"Range"`
	Days  int    `flag:"days" short:"d" default:"0" help:"Days"`
}

func TestAutoNameDisabled(t *testing.T) {
	os.Args = []string{"test", "--range", "30d", "-d", "5"}

	var cfg NoAutoConfig
	_ = Parse(&cfg, Options{NoAutoName: true})

	if cfg.Range != "30d" {
		t.Errorf("Range: got %q, want %q", cfg.Range, "30d")
	}
	if cfg.Days != 5 {
		t.Errorf("Days: got %d, want %d", cfg.Days, 5)
	}
}

func TestAutoShortDisabled(t *testing.T) {
	os.Args = []string{"test", "--range", "30d"}

	var cfg AutoNameConfig
	_ = Parse(&cfg, Options{NoAutoShort: true})

	if cfg.Range != "30d" {
		t.Errorf("Range: got %q, want %q", cfg.Range, "30d")
	}
}

type ExtendedTypesConfig struct {
	ID      int64         `flag:"id" short:"i" default:"42" help:"ID number"`
	Port    uint          `flag:"port" short:"p" default:"8080" help:"Port number"`
	Size    uint64        `flag:"size" default:"1048576" help:"Size in bytes"`
	Timeout time.Duration `flag:"timeout" short:"t" default:"5s" help:"Timeout duration"`
}

func TestInt64(t *testing.T) {
	os.Args = []string{"test", "--id", "9999999999"}

	var cfg ExtendedTypesConfig
	_ = Parse(&cfg)

	if cfg.ID != 9999999999 {
		t.Errorf("ID: got %d, want 9999999999", cfg.ID)
	}
}

func TestInt64Default(t *testing.T) {
	os.Args = []string{"test"}

	var cfg ExtendedTypesConfig
	_ = Parse(&cfg)

	if cfg.ID != 42 {
		t.Errorf("ID: got %d, want 42", cfg.ID)
	}
}

func TestUint(t *testing.T) {
	os.Args = []string{"test", "-p", "9090"}

	var cfg ExtendedTypesConfig
	_ = Parse(&cfg)

	if cfg.Port != 9090 {
		t.Errorf("Port: got %d, want 9090", cfg.Port)
	}
}

func TestUintDefault(t *testing.T) {
	os.Args = []string{"test"}

	var cfg ExtendedTypesConfig
	_ = Parse(&cfg)

	if cfg.Port != 8080 {
		t.Errorf("Port: got %d, want 8080", cfg.Port)
	}
}

func TestUint64(t *testing.T) {
	os.Args = []string{"test", "--size", "9999999"}

	var cfg ExtendedTypesConfig
	_ = Parse(&cfg)

	if cfg.Size != 9999999 {
		t.Errorf("Size: got %d, want 9999999", cfg.Size)
	}
}

func TestUint64Default(t *testing.T) {
	os.Args = []string{"test"}

	var cfg ExtendedTypesConfig
	_ = Parse(&cfg)

	if cfg.Size != 1048576 {
		t.Errorf("Size: got %d, want 1048576", cfg.Size)
	}
}

func TestDuration(t *testing.T) {
	os.Args = []string{"test", "-t", "30s"}

	var cfg ExtendedTypesConfig
	_ = Parse(&cfg)

	if cfg.Timeout != 30*time.Second {
		t.Errorf("Timeout: got %v, want 30s", cfg.Timeout)
	}
}

func TestDurationDefault(t *testing.T) {
	os.Args = []string{"test"}

	var cfg ExtendedTypesConfig
	_ = Parse(&cfg)

	if cfg.Timeout != 5*time.Second {
		t.Errorf("Timeout: got %v, want 5s", cfg.Timeout)
	}
}

func TestDurationComplex(t *testing.T) {
	os.Args = []string{"test", "--timeout", "2m30s"}

	var cfg ExtendedTypesConfig
	_ = Parse(&cfg)

	if cfg.Timeout != 150*time.Second {
		t.Errorf("Timeout: got %v, want 2m30s", cfg.Timeout)
	}
}

func TestInvalidIntValue(t *testing.T) {
	os.Args = []string{"test", "--days", "abc"}

	var cfg TestConfig
	err := Parse(&cfg)
	if err == nil {
		t.Fatal("expected error for invalid int value, got nil")
	}
}

func TestInvalidUintValue(t *testing.T) {
	os.Args = []string{"test", "--port", "-1"}

	var cfg ExtendedTypesConfig
	err := Parse(&cfg)
	if err == nil {
		t.Fatal("expected error for negative uint value, got nil")
	}
}

func TestInvalidFloatValue(t *testing.T) {
	os.Args = []string{"test", "--rate", "not-a-float"}

	var cfg TestConfig
	err := Parse(&cfg)
	if err == nil {
		t.Fatal("expected error for invalid float value, got nil")
	}
}

func TestInvalidDurationValue(t *testing.T) {
	os.Args = []string{"test", "--timeout", "not-duration"}

	var cfg ExtendedTypesConfig
	err := Parse(&cfg)
	if err == nil {
		t.Fatal("expected error for invalid duration value, got nil")
	}
}

func TestUnknownFlag(t *testing.T) {
	os.Args = []string{"test", "--bogus"}

	var cfg MinimalConfig
	err := Parse(&cfg)
	if err == nil {
		t.Fatal("expected error for unknown flag, got nil")
	}
}

func TestDoubleDashTreatsRemainderAsPositional(t *testing.T) {
	os.Args = []string{"test", "--full", "--", "--full"}

	var cfg TestConfig
	err := Parse(&cfg)
	if err != nil {
		t.Fatal(err)
	}

	if !cfg.Full {
		t.Error("Full should be true from --full before --")
	}
}

func TestStringWithSpaces(t *testing.T) {
	os.Args = []string{"test", "--name", "hello world"}

	var cfg MinimalConfig
	_ = Parse(&cfg)

	if cfg.Name != "hello world" {
		t.Errorf("Name: got %q, want %q", cfg.Name, "hello world")
	}
}

func TestBoolDefaultTrue(t *testing.T) {
	type BoolConfig struct {
		Verbose bool `default:"true" help:"Verbose output"`
	}

	os.Args = []string{"test"}

	var cfg BoolConfig
	_ = Parse(&cfg)

	if !cfg.Verbose {
		t.Error("Verbose should default to true")
	}
}

func TestBoolDefaultTrueOverridden(t *testing.T) {
	type BoolConfig struct {
		Verbose bool `default:"true" help:"Verbose output"`
	}

	os.Args = []string{"test", "--verbose=false"}

	var cfg BoolConfig
	_ = Parse(&cfg)

	if cfg.Verbose {
		t.Error("Verbose should be false after override")
	}
}

func TestNoColorOption(t *testing.T) {
	os.Args = []string{"test"}

	var cfg MinimalConfig
	_ = Parse(&cfg, Options{NoColor: true})

	if cfg.Name != "world" {
		t.Errorf("Name: got %q, want default %q", cfg.Name, "world")
	}
}

func TestNegativeInt(t *testing.T) {
	os.Args = []string{"test", "--days", "-5"}

	var cfg TestConfig
	_ = Parse(&cfg)

	if cfg.Days != -5 {
		t.Errorf("Days: got %d, want -5", cfg.Days)
	}
}

func TestFloat64ZeroDefault(t *testing.T) {
	type FloatConfig struct {
		Rate float64 `default:"0" help:"Rate"`
	}

	os.Args = []string{"test"}

	var cfg FloatConfig
	_ = Parse(&cfg)

	if cfg.Rate != 0.0 {
		t.Errorf("Rate: got %f, want 0.0", cfg.Rate)
	}
}

func TestStringDefaultEmpty(t *testing.T) {
	os.Args = []string{"test"}

	var cfg TestConfig
	_ = Parse(&cfg)

	// Output has default "json", Range has default "7d"
	if cfg.Output != "json" {
		t.Errorf("Output: got %q, want %q", cfg.Output, "json")
	}
}

func TestInvalidDefaultReturnsError(t *testing.T) {
	tests := []struct {
		name string
		cfg  any
	}{
		{"int", &struct {
			Count int `default:"abc"`
		}{}},
		{"bool", &struct {
			Enabled bool `default:"maybe"`
		}{}},
		{"duration", &struct {
			Timeout time.Duration `default:"soon"`
		}{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = []string{"test"}
			err := Parse(tt.cfg)
			if err == nil {
				t.Fatal("expected invalid default error, got nil")
			}
		})
	}
}

func TestUnsupportedFieldTypeReturnsError(t *testing.T) {
	type Config struct {
		Values []string
	}

	os.Args = []string{"test"}
	var cfg Config
	err := Parse(&cfg)
	if err == nil {
		t.Fatal("expected unsupported type error, got nil")
	}
}

func TestUnexportedFieldReturnsError(t *testing.T) {
	type Config struct {
		name string `flag:"name"` //nolint:unused // intentionally unexported to trigger ensureExported error
	}

	os.Args = []string{"test"}
	var cfg Config
	err := Parse(&cfg)
	if err == nil {
		t.Fatal("expected unexported field error, got nil")
	}
}

func TestDuplicateFlagNamesReturnError(t *testing.T) {
	type Config struct {
		First  string `flag:"name"`
		Second string `flag:"name"`
	}

	os.Args = []string{"test"}
	var cfg Config
	err := Parse(&cfg)
	if err == nil {
		t.Fatal("expected duplicate flag error, got nil")
	}
}

func TestDuplicateShortNamesReturnError(t *testing.T) {
	type Config struct {
		First  string `short:"n"`
		Second string `short:"n"`
	}

	os.Args = []string{"test"}
	var cfg Config
	err := Parse(&cfg)
	if err == nil {
		t.Fatal("expected duplicate short flag error, got nil")
	}
}

func TestLongAndShortNamesShareNamespace(t *testing.T) {
	type Config struct {
		First  string `flag:"name"`
		Second string `short:"name"`
	}

	os.Args = []string{"test"}
	var cfg Config
	err := Parse(&cfg)
	if err == nil {
		t.Fatal("expected duplicate flag namespace error, got nil")
	}
}

func TestEmptyShortTagDisablesAutoShort(t *testing.T) {
	type Config struct {
		Range  string `short:"" default:"7d" help:"Range"`
		Output string `default:"text" help:"Output"`
	}

	// -r should fail, --range should work, -o should work for Output
	os.Args = []string{"test", "-r", "bad"}
	var cfg Config
	err := Parse(&cfg)
	if err == nil {
		t.Fatal("expected error for disabled short flag -r")
	}

	// -o should work since Output has no short tag and auto-short is on
	os.Args = []string{"test", "-o", "csv"}
	var cfg2 Config
	err = Parse(&cfg2)
	if err != nil {
		t.Fatal(err)
	}
	if cfg2.Output != "csv" {
		t.Errorf("Output: got %q, want csv", cfg2.Output)
	}
}

func TestUnicodeAutoNameAndShort(t *testing.T) {
	type Config struct {
		Éclair string
	}

	os.Args = []string{"test", "-é", "vanilla"}
	var cfg Config
	err := Parse(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Éclair != "vanilla" {
		t.Errorf("Éclair: got %q, want vanilla", cfg.Éclair)
	}
}

// --- Positional args tests ---

func TestPositionalBasic(t *testing.T) {
	type Config struct {
		Source string `positional:"" help:"input file"`
		Target string `positional:"" help:"output file"`
	}

	os.Args = []string{"test", "src.txt", "dst.txt"}
	var cfg Config
	err := Parse(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Source != "src.txt" {
		t.Errorf("Source: got %q, want src.txt", cfg.Source)
	}
	if cfg.Target != "dst.txt" {
		t.Errorf("Target: got %q, want dst.txt", cfg.Target)
	}
}

func TestPositionalWithFlags(t *testing.T) {
	type Config struct {
		Verbose bool   `flag:"verbose" short:"v" help:"Verbose"`
		Source  string `positional:"" help:"input"`
		Target  string `positional:"" help:"output"`
	}

	os.Args = []string{"test", "-v", "in.txt", "out.txt"}
	var cfg Config
	err := Parse(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Verbose {
		t.Error("Verbose should be true")
	}
	if cfg.Source != "in.txt" {
		t.Errorf("Source: got %q, want in.txt", cfg.Source)
	}
	if cfg.Target != "out.txt" {
		t.Errorf("Target: got %q, want out.txt", cfg.Target)
	}
}

func TestPositionalVariadic(t *testing.T) {
	type Config struct {
		Source string   `positional:"" help:"input"`
		Files  []string `positional:"" help:"additional files"`
	}

	os.Args = []string{"test", "src.txt", "a.txt", "b.txt", "c.txt"}
	var cfg Config
	err := Parse(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Source != "src.txt" {
		t.Errorf("Source: got %q, want src.txt", cfg.Source)
	}
	if len(cfg.Files) != 3 {
		t.Fatalf("Files len: got %d, want 3", len(cfg.Files))
	}
	if cfg.Files[0] != "a.txt" || cfg.Files[1] != "b.txt" || cfg.Files[2] != "c.txt" {
		t.Errorf("Files: got %v", cfg.Files)
	}
}

func TestPositionalVariadicEmpty(t *testing.T) {
	type Config struct {
		Source string   `positional:"" help:"input"`
		Files  []string `positional:"" help:"additional files"`
	}

	os.Args = []string{"test", "src.txt"}
	var cfg Config
	err := Parse(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Source != "src.txt" {
		t.Errorf("Source: got %q, want src.txt", cfg.Source)
	}
	if len(cfg.Files) != 0 {
		t.Errorf("Files: got %v, want empty", cfg.Files)
	}
}

func TestPositionalVariadicOnly(t *testing.T) {
	type Config struct {
		Files []string `positional:"" help:"files"`
	}

	os.Args = []string{"test", "a.txt", "b.txt"}
	var cfg Config
	err := Parse(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Files) != 2 {
		t.Fatalf("Files len: got %d, want 2", len(cfg.Files))
	}
	if cfg.Files[0] != "a.txt" || cfg.Files[1] != "b.txt" {
		t.Errorf("Files: got %v", cfg.Files)
	}
}

func TestPositionalInt(t *testing.T) {
	type Config struct {
		Port int `positional:"" help:"port"`
	}

	os.Args = []string{"test", "8080"}
	var cfg Config
	err := Parse(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 8080 {
		t.Errorf("Port: got %d, want 8080", cfg.Port)
	}
}

func TestPositionalBool(t *testing.T) {
	type Config struct {
		Force bool `positional:"" help:"force"`
	}

	os.Args = []string{"test", "true"}
	var cfg Config
	err := Parse(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Force {
		t.Error("Force should be true")
	}
}

func TestPositionalDuration(t *testing.T) {
	type Config struct {
		Timeout time.Duration `positional:"" help:"timeout"`
	}

	os.Args = []string{"test", "30s"}
	var cfg Config
	err := Parse(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("Timeout: got %v, want 30s", cfg.Timeout)
	}
}

func TestPositionalFloat64(t *testing.T) {
	type Config struct {
		Rate float64 `positional:"" help:"rate"`
	}

	os.Args = []string{"test", "1.5"}
	var cfg Config
	err := Parse(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Rate != 1.5 {
		t.Errorf("Rate: got %f, want 1.5", cfg.Rate)
	}
}

func TestPositionalMissingError(t *testing.T) {
	type Config struct {
		Source string `positional:"" help:"input"`
		Target string `positional:"" help:"output"`
	}

	os.Args = []string{"test", "only-one.txt"}
	var cfg Config
	err := Parse(&cfg)
	if err == nil {
		t.Fatal("expected error for missing positional, got nil")
	}
}

func TestPositionalTooManyError(t *testing.T) {
	type Config struct {
		Source string `positional:"" help:"input"`
	}

	os.Args = []string{"test", "a.txt", "b.txt"}
	var cfg Config
	err := Parse(&cfg)
	// This should still work - extra args go to the returned slice
	// Actually no, with positional fields we consume them.
	// But there's no variadic, so extra args should error.
	if err == nil {
		t.Fatal("expected error for too many positional args, got nil")
	}
}

func TestPositionalFlagIgnored(t *testing.T) {
	type Config struct {
		Source string `flag:"source" positional:"" help:"input"`
	}

	// --source should NOT work, positional should consume args
	os.Args = []string{"test", "file.txt"}
	var cfg Config
	err := Parse(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Source != "file.txt" {
		t.Errorf("Source: got %q, want file.txt", cfg.Source)
	}
}

func TestPositionalNonContiguousError(t *testing.T) {
	type Config struct {
		Source string `positional:"" help:"input"`
		Extra  string `flag:"extra" help:"extra flag"`
		Target string `positional:"" help:"output"`
	}

	os.Args = []string{"test", "a", "b"}
	var cfg Config
	err := Parse(&cfg)
	if err == nil {
		t.Fatal("expected error for non-contiguous positionals, got nil")
	}
}

func TestPositionalVariadicNotLastError(t *testing.T) {
	type Config struct {
		Files []string `positional:"" help:"files"`
		Target string  `positional:"" help:"target"`
	}

	os.Args = []string{"test", "a", "b"}
	var cfg Config
	err := Parse(&cfg)
	if err == nil {
		t.Fatal("expected error for variadic not last, got nil")
	}
}

func TestPositionalUnsupportedTypeError(t *testing.T) {
	type Config struct {
		Values []int `positional:"" help:"values"`
	}

	os.Args = []string{"test", "1", "2"}
	var cfg Config
	err := Parse(&cfg)
	if err == nil {
		t.Fatal("expected error for unsupported positional type, got nil")
	}
}

func TestPositionalInvalidIntError(t *testing.T) {
	type Config struct {
		Port int `positional:"" help:"port"`
	}

	os.Args = []string{"test", "not-a-port"}
	var cfg Config
	err := Parse(&cfg)
	if err == nil {
		t.Fatal("expected error for invalid int positional, got nil")
	}
}

func TestPositionalInvalidBoolError(t *testing.T) {
	type Config struct {
		Force bool `positional:"" help:"force"`
	}

	os.Args = []string{"test", "maybe"}
	var cfg Config
	err := Parse(&cfg)
	if err == nil {
		t.Fatal("expected error for invalid bool positional, got nil")
	}
}

func TestPositionalWithDoubleDash(t *testing.T) {
	type Config struct {
		Source string `positional:"" help:"input"`
	}

	os.Args = []string{"test", "--", "file.txt"}
	var cfg Config
	err := Parse(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Source != "file.txt" {
		t.Errorf("Source: got %q, want file.txt", cfg.Source)
	}
}



func TestPositionalDefault(t *testing.T) {
	type Config struct {
		Output string `positional:"" default:"output.mp4" help:"output file"`
	}

	os.Args = []string{"test"}
	var cfg Config
	err := Parse(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Output != "output.mp4" {
		t.Errorf("Output: got %q, want output.mp4", cfg.Output)
	}
}

func TestPositionalDefaultOverridden(t *testing.T) {
	type Config struct {
		Output string `positional:"" default:"output.mp4" help:"output file"`
	}

	os.Args = []string{"test", "custom.mov"}
	var cfg Config
	err := Parse(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Output != "custom.mov" {
		t.Errorf("Output: got %q, want custom.mov", cfg.Output)
	}
}

func TestPositionalDefaultWithFlags(t *testing.T) {
	type Config struct {
		Verbose bool   `flag:"verbose" short:"v" help:"Verbose"`
		Source  string `positional:"" help:"input"`
		Output  string `positional:"" default:"out.txt" help:"output"`
	}

	// Only provide Source, Output should use default
	os.Args = []string{"test", "-v", "in.txt"}
	var cfg Config
	err := Parse(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Verbose {
		t.Error("Verbose should be true")
	}
	if cfg.Source != "in.txt" {
		t.Errorf("Source: got %q, want in.txt", cfg.Source)
	}
	if cfg.Output != "out.txt" {
		t.Errorf("Output: got %q, want out.txt", cfg.Output)
	}
}

func TestPositionalDefaultInt(t *testing.T) {
	type Config struct {
		Port int `positional:"" default:"8080" help:"port"`
	}

	os.Args = []string{"test"}
	var cfg Config
	err := Parse(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 8080 {
		t.Errorf("Port: got %d, want 8080", cfg.Port)
	}
}

func TestPositionalMissingNoDefaultError(t *testing.T) {
	type Config struct {
		Source string `positional:"" help:"input"`
	}

	os.Args = []string{"test"}
	var cfg Config
	err := Parse(&cfg)
	if err == nil {
		t.Fatal("expected error for missing positional without default, got nil")
	}
}
