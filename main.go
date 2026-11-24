package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, out, errw io.Writer) int {
	fs := flag.NewFlagSet("envgap", flag.ContinueOnError)
	fs.SetOutput(errw)
	asJSON := fs.Bool("json", false, "print findings as JSON")
	ignoreExtra := fs.Bool("ignore-extra", false, "do not report keys only in the actual file")
	optional := fs.String("optional", "", "comma-separated keys allowed to be missing or empty")
	fs.Usage = func() {
		fmt.Fprintln(errw, "usage: envgap [flags] EXAMPLE_FILE ACTUAL_FILE\n\nReports key names and line numbers only. Values are never printed.\nExit: 0 clean, 1 findings, 2 usage or read error.")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 2 {
		fs.Usage()
		return 2
	}
	read := func(p string) (File, bool) {
		b, err := os.ReadFile(p)
		if err != nil {
			// Report the OS error kind, never file content.
			fmt.Fprintf(errw, "envgap: cannot read %s\n", p)
			return File{}, false
		}
		return Parse(string(b)), true
	}
	ex, ok1 := read(fs.Arg(0))
	ac, ok2 := read(fs.Arg(1))
	if !ok1 || !ok2 {
		return 2
	}
	opt := Options{IgnoreExtra: *ignoreExtra, Optional: map[string]bool{}}
	for _, k := range strings.Split(*optional, ",") {
		if k = strings.TrimSpace(k); k != "" {
			opt.Optional[k] = true
		}
	}
	fnd := Compare(ex, ac, opt)
	if *asJSON {
		if fnd == nil {
			fnd = []Finding{}
		}
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		enc.Encode(fnd)
	} else if len(fnd) == 0 {
		fmt.Fprintln(out, "ok: no differences found")
	} else {
		for _, f := range fnd {
			name := f.Key
			if name == "" {
				name = "-"
			}
			ls := ""
			if len(f.Lines) > 0 {
				ls = fmt.Sprintf(" line %s", joinInts(f.Lines))
			}
			fmt.Fprintf(out, "%-19s %-8s %s%s\n", f.Kind, f.File, name, ls)
		}
		fmt.Fprintf(out, "%d finding(s)\n", len(fnd))
	}
	if len(fnd) > 0 {
		return 1
	}
	return 0
}

func joinInts(a []int) string {
	s := make([]string, len(a))
	for i, v := range a {
		s[i] = fmt.Sprint(v)
	}
	return strings.Join(s, ",")
}
