//go:build ignore

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/sahaj-b/sflag"
)

func main() {
	var flags struct {
		Verbose    bool
		Date       string        `help:"in YYYY-MM-DD format"`
		Range      string        `default:"7d" help:"Range of data"`
		Rate       float64       `short:"R"`
		MaxRetries int           `flag:"max" short:"" default:"3" help:"Max retries"`
		Timeout    time.Duration `positional:"" default:"30s" help:"Request timeout"`
		Files      []string      `positional:"" help:"Additional files"`
	}

	// Example with custom usage text
	if err := sflag.Parse(&flags, sflag.Options{
		ExtraUsage: "\nExamples:\n  myapp --format json file.txt",
	}); err != nil {
		os.Exit(1)
	}

	fmt.Printf("Range: %s\n", flags.Range)
}
