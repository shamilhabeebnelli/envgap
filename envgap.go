// Command envgap compares a dotenv example file with an actual dotenv file.
// It reports key names and line numbers only. It never returns values.
package main

import (
	"sort"
	"strings"
)

// Entry is one parsed KEY=VALUE line. Value stays inside this package's callers
// only for comparison; reports never include it.
type Entry struct {
	Key   string
	Value string
	Line  int
}

// File is a parsed dotenv file.
type File struct {
	Entries       []Entry
	Invalid       []int // line numbers that are not comments, blanks or KEY=VALUE
	Multiline     []int // lines that open a quote never closed on the same line
	HasBOM        bool
	HasCRLF       bool
	SpaceAroundEq []int
}

func validKey(k string) bool {
	if k == "" {
		return false
	}
	for i, r := range k {
		switch {
		case r == '_' || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z'):
		case r >= '0' && r <= '9' && i > 0:
		default:
			return false
		}
	}
	return true
}

// unquote returns the value with one layer of matching quotes removed.
// closed is false when an opening quote has no closing quote on the line.
func unquote(v string) (out string, closed bool) {
	if v == "" {
		return "", true
	}
	q := v[0]
	if q == '"' || q == '\'' {
		for i := 1; i < len(v); i++ {
			if q == '"' && v[i] == '\\' {
				i++
				continue
			}
			if v[i] == q {
				return v[1:i], true
			}
		}
		return v[1:], false
	}
	// unquoted: strip trailing " #comment"
	if i := strings.Index(v, " #"); i >= 0 {
		v = v[:i]
	}
	return strings.TrimSpace(v), true
}

// Parse reads dotenv text. It is deliberately small: no interpolation, and
// multi-line quoted values are flagged, not supported.
func Parse(text string) File {
	var f File
	if strings.HasPrefix(text, "\ufeff") {
		f.HasBOM = true
		text = strings.TrimPrefix(text, "\ufeff")
	}
	if strings.Contains(text, "\r\n") {
		f.HasCRLF = true
	}
	for i, raw := range strings.Split(text, "\n") {
		n := i + 1
		line := strings.TrimRight(raw, "\r")
		t := strings.TrimSpace(line)
		if t == "" || strings.HasPrefix(t, "#") {
			continue
		}
		t = strings.TrimPrefix(t, "export ")
		t = strings.TrimLeft(t, " \t")
		eq := strings.IndexByte(t, '=')
		if eq < 0 {
			f.Invalid = append(f.Invalid, n)
			continue
		}
		rawKey := t[:eq]
		key := strings.TrimSpace(rawKey)
		if !validKey(key) {
			f.Invalid = append(f.Invalid, n)
			continue
		}
		rest := t[eq+1:]
		if rawKey != key || strings.HasPrefix(rest, " ") || strings.HasPrefix(rest, "\t") {
			f.SpaceAroundEq = append(f.SpaceAroundEq, n)
		}
		val, closed := unquote(strings.TrimSpace(rest))
		if !closed {
			f.Multiline = append(f.Multiline, n)
		}
		f.Entries = append(f.Entries, Entry{Key: key, Value: val, Line: n})
	}
	return f
}

// Finding is one reportable item. No field ever holds a value.
type Finding struct {
	Kind  string `json:"kind"`
	Key   string `json:"key,omitempty"`
	Lines []int  `json:"lines,omitempty"`
	File  string `json:"file"` // "example" or "actual"
}

// Options tunes which findings count.
type Options struct {
	IgnoreExtra bool
	Optional    map[string]bool // keys the example may leave missing/empty in actual
}

func last(f File) map[string]Entry {
	m := map[string]Entry{}
	for _, e := range f.Entries {
		m[e.Key] = e
	}
	return m
}

func dupes(f File, which string) []Finding {
	lines := map[string][]int{}
	for _, e := range f.Entries {
		lines[e.Key] = append(lines[e.Key], e.Line)
	}
	var out []Finding
	for k, l := range lines {
		if len(l) > 1 {
			out = append(out, Finding{Kind: "duplicate", Key: k, Lines: l, File: which})
		}
	}
	return out
}

// Compare returns sorted findings.
func Compare(example, actual File, o Options) []Finding {
	var out []Finding
	ex, ac := last(example), last(actual)
	for _, e := range example.Entries {
		if ex[e.Key].Line != e.Line {
			continue // report each key once
		}
		a, ok := ac[e.Key]
		switch {
		case !ok:
			if !o.Optional[e.Key] {
				out = append(out, Finding{Kind: "missing", Key: e.Key, Lines: []int{e.Line}, File: "actual"})
			}
		case a.Value == "" && !o.Optional[e.Key]:
			out = append(out, Finding{Kind: "empty", Key: e.Key, Lines: []int{a.Line}, File: "actual"})
		case a.Value != "" && e.Value != "" && a.Value == e.Value && looksPlaceholder(e.Value):
			out = append(out, Finding{Kind: "placeholder", Key: e.Key, Lines: []int{a.Line}, File: "actual"})
		}
	}
	if !o.IgnoreExtra {
		for _, a := range actual.Entries {
			if ac[a.Key].Line != a.Line {
				continue
			}
			if _, ok := ex[a.Key]; !ok {
				out = append(out, Finding{Kind: "extra", Key: a.Key, Lines: []int{a.Line}, File: "actual"})
			}
		}
	}
	out = append(out, dupes(example, "example")...)
	out = append(out, dupes(actual, "actual")...)
	for _, p := range []struct {
		f     File
		which string
	}{{example, "example"}, {actual, "actual"}} {
		if len(p.f.Invalid) > 0 {
			out = append(out, Finding{Kind: "invalid-line", Lines: p.f.Invalid, File: p.which})
		}
		if len(p.f.Multiline) > 0 {
			out = append(out, Finding{Kind: "unclosed-quote", Lines: p.f.Multiline, File: p.which})
		}
		if len(p.f.SpaceAroundEq) > 0 {
			out = append(out, Finding{Kind: "space-around-equals", Lines: p.f.SpaceAroundEq, File: p.which})
		}
		if p.f.HasBOM {
			out = append(out, Finding{Kind: "bom", File: p.which})
		}
		if p.f.HasCRLF {
			out = append(out, Finding{Kind: "crlf", File: p.which})
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Kind != out[j].Kind {
			return out[i].Kind < out[j].Kind
		}
		if out[i].File != out[j].File {
			return out[i].File < out[j].File
		}
		return out[i].Key < out[j].Key
	})
	return out
}

// looksPlaceholder is true for example values nobody should deploy with.
func looksPlaceholder(v string) bool {
	l := strings.ToLower(v)
	for _, p := range []string{"changeme", "change_me", "change-me", "your_", "your-", "xxx", "todo", "replace", "example", "<", "secret", "password", "placeholder"} {
		if strings.Contains(l, p) {
			return true
		}
	}
	return false
}
