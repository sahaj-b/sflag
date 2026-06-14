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

func showHelp(w io.Writer, prog string, flags []flagDef, positionals []positionalDef) {
	usage := prog
	if len(flags) > 0 {
		usage += " [options]"
	}
	for _, p := range positionals {
		name := strings.ToUpper(p.field.Name)
		switch {
		case p.isVariadic:
			usage += " " + name + "..."
		case p.defStr != "":
			usage += " [" + name + "]"
		default:
			usage += " " + name
		}
	}
	fmt.Fprintf(w, "%sUsage:%s %s\n\n", cBold, cReset, usage)

	if len(flags) > 0 {
		rows := make([]helpRow, 0, len(flags)+1)
		for _, f := range flags {
			label := flagLabel(f)
			if f.typeName != "" {
				label += " " + cBlue + f.typeName + cReset
			}
			rows = append(rows, helpRow{label: label, help: f.help, def: f.defStr})
		}
		rows = append(rows, helpRow{label: cGreen + "-h, --help" + cReset, help: "Display help information"})
		printRows(w, "Options", rows)
	}

	if len(positionals) > 0 {
		if len(flags) > 0 {
			fmt.Fprintln(w)
		}
		rows := make([]helpRow, 0, len(positionals))
		for _, p := range positionals {
			rows = append(rows, helpRow{label: positionalLabel(p), help: p.help, def: p.defStr})
		}
		printRows(w, "Arguments", rows)
	}
}

type helpRow struct {
	label string
	help  string
	def   string
}

func printRows(w io.Writer, title string, rows []helpRow) {
	fmt.Fprintf(w, "%s%s:%s\n", cBold, title, cReset)

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
		fmt.Fprintf(w, "  %s%s%s\n", row.label, strings.Repeat(" ", width-len(stripAnsi(row.label))), help)
	}
}

func positionalLabel(p positionalDef) string {
	name := "<" + strings.ToUpper(p.field.Name) + ">"
	if p.isVariadic {
		return cGreen + name + "..." + cReset
	}
	if typeLabel, _ := typeName(p.field.Type, true); typeLabel != "" {
		return cGreen + name + cReset + " " + cBlue + typeLabel + cReset
	}
	return cGreen + name + cReset
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
