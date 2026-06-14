# sflag

A minimal, opinionated library for struct-tagged CLI flags for Go. Zero deps, wraps stdlib `flag`.

## Example
```go
var flags struct {
  Verbose    bool
  Date       string  `help:"in YYYY-MM-DD format"`
  Range      string  `default:"7d" help:"Range of data"`
  Rate       float64 `short:"R"`
  MaxRetries int     `flag:"max" short:"" default:"3" help:"Max retries"`
}
args, err := sflag.Parse(&flags)
fmt.Println(flags.Date)
fmt.Println(args)
```

### Help example
<img width="801" height="322" alt="image" src="https://github.com/user-attachments/assets/e37b5721-e811-4400-a5fe-2f7576a30f1f" />

## Features

- **Auto names**: field `ApiKey` → `--api-key` (kebab-case)
- **Auto shorts**: first char of flag name. Conflicts silently skipped
- **Positional args**: returned from `Parse`, also `--` stops flag parsing, same as `flag` 
- **Colored help**: and plain text when piped (unix-based). [**`NO_COLOR`**](https://no-color.org/) respected
- **Stdlib `flag`** behavior: `-flag`, `--flag`, and `--flag=true` all work

## Tags

| Tag             | Default                  |
| -----           | ---------                |
| `flag:"name"`   | kebab-case of field name |
| `short:"x"`     | first char of long name  |
| `default:"val"` | zero value               |
| `help:"text"`   | empty                    |


## Options

```go
sflag.Parse(&cfg, sflag.Options{
  ProgramName:  "myapp", // app name shown in help, default: os.Args[0]
  NoAutoName:   true,    // disable kebab-case derivation, default: false
  NoAutoShort:  true,    // disable short derivation, default: false
  NoColor:      true,    // force no colors, default: false
})
```

## Supported Types (matches stdlib `flag`)

`string`, `int`, `int64`, `uint`, `uint64`, `bool`, `float64`, `time.Duration`
