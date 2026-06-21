package sflag

import (
	"fmt"
	"io"
	"strings"
)

var (
	cBold   string
	cGreen  string
	cBlue   string
	cYellow string
	cReset  string
)

func setColors(on bool) {
	if !on {
		cBold, cGreen, cBlue, cYellow, cReset = "", "", "", "", ""
		return
	}
	cBold = "\x1b[1m"
	cGreen = "\x1b[32m"
	cBlue = "\x1b[34m"
	cYellow = "\x1b[33m"
	cReset = "\x1b[0m"
}

func showHelp(w io.Writer, prog string, flags []FlagInfo, positionals []positionalDef, opts Options) {
	usage := prog
	if len(flags) > 0 {
		usage += " [options]"
	}
	for _, p := range positionals {
		switch {
		case p.info.IsVariadic:
			usage += " " + p.info.Name + "..."
		case p.info.DefStr != "":
			usage += " [" + p.info.Name + "]"
		default:
			usage += " " + p.info.Name
		}
	}
	fmt.Fprintf(w, "%sUsage:%s %s\n\n", cBold, cReset, usage) //nolint:errcheck

	if len(flags) > 0 {
		rows := make([]helpRow, 0, len(flags)+1)
		for _, f := range flags {
			label := flagLabel(f)
			if f.TypeName != "" {
				label += " " + cBlue + f.TypeName + cReset
			}
			rows = append(rows, helpRow{label: label, help: f.Help, def: f.DefStr})
		}
		rows = append(rows, helpRow{label: cGreen + "-h, --help" + cReset, help: "Display help information"})
		printRows(w, "Options", rows)
	}

	if len(positionals) > 0 {
		if len(flags) > 0 {
			fmt.Fprintln(w) //nolint:errcheck
		}
		rows := make([]helpRow, 0, len(positionals))
		for _, p := range positionals {
			rows = append(rows, helpRow{label: positionalLabel(p.info), help: p.info.Help, def: p.info.DefStr})
		}
		printRows(w, "Arguments", rows)
	}

	if opts.ExtraUsage != "" {
		fmt.Fprintln(w)                  //nolint:errcheck
		fmt.Fprintln(w, opts.ExtraUsage) //nolint:errcheck
	}
}

type helpRow struct {
	label string
	help  string
	def   string
}

func printRows(w io.Writer, title string, rows []helpRow) {
	fmt.Fprintf(w, "%s%s:%s\n", cBold, title, cReset) //nolint:errcheck

	width := 0
	for _, row := range rows {
		if n := len(stripAnsi(row.label)); n > width {
			width = n
		}
	}
	width += 2

	for _, row := range rows {
		help := row.help
		if row.def != "" {
			if help != "" {
				help += " "
			}
			help += cYellow + "(default: " + row.def + ")" + cReset
		}
		fmt.Fprintf(w, "  %s%s%s\n", row.label, strings.Repeat(" ", width-len(stripAnsi(row.label))), help) //nolint:errcheck
	}
}

func positionalLabel(p PositionalInfo) string {
	name := "<" + p.Name + ">"
	if p.IsVariadic {
		return cGreen + name + "..." + cReset
	}
	if p.TypeName != "" {
		return cGreen + name + cReset + " " + cBlue + p.TypeName + cReset
	}
	return cGreen + name + cReset
}

func flagLabel(f FlagInfo) string {
	if f.Short != "" {
		return cGreen + fmt.Sprintf("-%s, --%s", f.Short, f.Long) + cReset
	}
	return cGreen + fmt.Sprintf("    --%s", f.Long) + cReset
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
