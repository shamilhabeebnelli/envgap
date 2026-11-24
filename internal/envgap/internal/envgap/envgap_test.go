package envgap

import (
	"encoding/json"
	"strings"
	"testing"
)

func kinds(f []Finding) []string {
	var o []string
	for _, x := range f {
		o = append(o, x.Kind+":"+x.Key)
	}
	return o
}

func has(f []Finding, kind, key string) bool {
	for _, x := range f {
		if x.Kind == kind && x.Key == key {
			return true
		}
	}
	return false
}

func TestMissingExtraEmpty(t *testing.T) {
	ex := Parse("A=1\nB=2\nC=3\n")
	ac := Parse("A=1\nC=\nD=9\n")
	f := Compare(ex, ac, Options{})
	if !has(f, "missing", "B") || !has(f, "empty", "C") || !has(f, "extra", "D") || has(f, "missing", "A") {
		t.Fatalf("got %v", kinds(f))
	}
}

func TestCleanIsEmpty(t *testing.T) {
	if f := Compare(Parse("A=1\n"), Parse("A=real\n"), Options{}); len(f) != 0 {
		t.Fatalf("got %v", kinds(f))
	}
}

func TestPlaceholderOnlyWhenCopiedPlaceholder(t *testing.T) {
	ex := Parse("API=changeme\nPORT=3000\n")
	ac := Parse("API=changeme\nPORT=3000\n")
	f := Compare(ex, ac, Options{})
	if !has(f, "placeholder", "API") || has(f, "placeholder", "PORT") {
		t.Fatalf("got %v", kinds(f))
	}
}

func TestExportQuotesComments(t *testing.T) {
	f := Parse("export A=\"x # y\"\nB='z'\nC=plain # note\n# D=1\n\n")
	m := map[string]string{}
	for _, e := range f.Entries {
		m[e.Key] = e.Value
	}
	if m["A"] != "x # y" || m["B"] != "z" || m["C"] != "plain" || len(m) != 3 {
		t.Fatalf("got %v", m)
	}
}

func TestDuplicatesUseLastValueAndReportLines(t *testing.T) {
	ac := Parse("A=\nA=2\n")
	f := Compare(Parse("A=1\n"), ac, Options{})
	if !has(f, "duplicate", "A") || has(f, "empty", "A") {
		t.Fatalf("got %v", kinds(f))
	}
	for _, x := range f {
		if x.Kind == "duplicate" && (len(x.Lines) != 2 || x.Lines[0] != 1) {
			t.Fatalf("lines %v", x.Lines)
		}
	}
}

func TestInvalidAndUnclosed(t *testing.T) {
	f := Parse("not a pair\n1BAD=x\nOK=\"abc\nZ=1\n")
	if len(f.Invalid) != 2 || f.Invalid[0] != 1 || f.Invalid[1] != 2 {
		t.Fatalf("invalid %v", f.Invalid)
	}
	if len(f.Multiline) != 1 || f.Multiline[0] != 3 {
		t.Fatalf("multi %v", f.Multiline)
	}
}

func TestBOMCRLFSpaces(t *testing.T) {
	f := Parse("\ufeffA = 1\r\nB=2\r\n")
	if !f.HasBOM || !f.HasCRLF || len(f.SpaceAroundEq) != 1 || f.Entries[0].Key != "A" || f.Entries[0].Value != "1" {
		t.Fatalf("%+v", f)
	}
}

func TestOptionalAndIgnoreExtra(t *testing.T) {
	ex := Parse("A=1\nB=2\n")
	ac := Parse("A=1\nB=\nZ=1\n")
	f := Compare(ex, ac, Options{IgnoreExtra: true, Optional: map[string]bool{"B": true}})
	if len(f) != 0 {
		t.Fatalf("got %v", kinds(f))
	}
}

func TestNoValueEverInFindings(t *testing.T) {
	const s1, s2, s3 = "SENTINEL_ONE_91ab", "SENTINEL_TWO_77cd", "SENTINEL_THREE_55ef"
	ex := Parse("A=changeme_" + s1 + "\nB=1\nbad " + s3 + "\n")
	ac := Parse("A=changeme_" + s1 + "\nQ=" + s2 + "\nQ=" + s2 + "\nthis " + s3 + " line\nK=\"" + s3 + "\n")
	f := Compare(ex, ac, Options{})
	b, _ := json.Marshal(f)
	for _, s := range []string{s1, s2, s3} {
		if strings.Contains(string(b), s) {
			t.Fatalf("value leaked: %s", s)
		}
	}
	if len(f) == 0 {
		t.Fatal("expected findings")
	}
}
