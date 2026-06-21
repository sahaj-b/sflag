# sflag

A minimal, opinionated library for struct-tagged CLI flags for Go. Zero deps, wraps stdlib `flag`.

## Example
```go
var flags struct {
  Verbose    bool
  Date       string        `help:"in YYYY-MM-DD format"`
  Range      string        `default:"7d" help:"Range of data"`
  Rate       float64       `short:"R"`
  MaxRetries int           `flag:"max" short:"" default:"3" help:"Max retries"`
  Timeout    time.Duration `positional:"" default:"30s" help:"Request timeout"`
  Files      []string      `positional:"" help:"Additional files"`
}
err := sflag.Parse(&flags)
fmt.Println(flags.Date)         
```

### Help example
<img width="847" height="471" alt="image" src="https://github.com/user-attachments/assets/7119cc66-d967-4c61-9586-21a69b84de92" />

## Features

- **Auto names**: field `ApiKey` → `--api-key` (kebab-case)
- **Auto shorts**: first char of flag name. Conflicts silently skipped
- **Positional args**: `positional:""` tag binds CLI positionals to struct fields; `--` stops flag parsing, same as `flag`
- **Colored help**: and plain text when piped (unix-based). [**`NO_COLOR`**](https://no-color.org/) respected
- **Stdlib `flag`** behavior: `-myflag`, `--myflag`, and `--myflag=true` all work

## Tags

| Tag             | Default                       |
| -----           | ---------                     |
| `flag:"name"`   | kebab-case of field name      |
| `short:"x"`     | first char of long name       |
| `default:"val"` | zero value                    |
| `help:"text"`   | empty                         |
| `positional:""` | marks field as positional arg |


## Positional args (`positional:""`) Rules
- Must be grouped together (no non-positionals between them)
- Only the last positional field can be `[]string` (variadic)
- `flag:""` tags on positional fields are ignored
- `default:"val"` in positional fields make them optional (wrapped in `[]` in help)
- `default:"val"` is ignored for **variadic**(`[]string`) positionals

## Parsing

- `Parse` (`Parse(target any, opts ...Options)`) reads `os.Args[1:]`
- `ParseArgs` (`ParseArgs(target any, args []string, opts ...Options)`) accepts explicit args
- `-h` / `--help` prints usage and calls `os.Exit(0)`
- invalid flags prints the error and usage, then returns an error (does not exit)

## Options
```go
sflag.Parse(&cfg, sflag.Options{
  ProgramName:  "myapp", // app name shown in help, default: os.Args[0]
  NoAutoName:   true,    // disable kebab-case derivation, default: false
  NoAutoShort:  true,    // disable short derivation, default: false
  NoColor:      true,    // force no colors, default: false
  ExtraUsage:   "\nExamples:\n  myapp --format json file.txt", // extra text appended to help
})
```

## Supported Types (matches stdlib `flag`)

`string`, `int`, `int64`, `uint`, `uint64`, `bool`, `float64`, `time.Duration`, `[]string` (positional only)
