package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, dir, name, body string) string {
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestExitCodesAndNoLeak(t *testing.T) {
	d := t.TempDir()
	const secret = "SENTINEL_cli_8842"
	ex := write(t, d, "ex", "A=1\nB=2\n")
	ac := write(t, d, "ac", "A="+secret+"\nB=\nC="+secret+"\n")
	var o, e bytes.Buffer
	if c := run([]string{ex, ac}, &o, &e); c != 1 {
		t.Fatalf("code %d", c)
	}
	if strings.Contains(o.String()+e.String(), secret) {
		t.Fatal("leak")
	}
	if !strings.Contains(o.String(), "empty") || !strings.Contains(o.String(), "extra") {
		t.Fatal(o.String())
	}
	o.Reset()
	if c := run([]string{"-json", ex, ex}, &o, &e); c != 0 || strings.TrimSpace(o.String()) != "[]" {
		t.Fatalf("clean %d %q", c, o.String())
	}
}

func TestUsageAndReadError(t *testing.T) {
	var o, e bytes.Buffer
	if c := run([]string{"only-one"}, &o, &e); c != 2 {
		t.Fatalf("usage %d", c)
	}
	e.Reset()
	if c := run([]string{"/no/such/a", "/no/such/b"}, &o, &e); c != 2 || !strings.Contains(e.String(), "cannot read") {
		t.Fatalf("read %d %s", c, e.String())
	}
}

func TestOptionalFlagJSON(t *testing.T) {
	d := t.TempDir()
	ex := write(t, d, "ex", "A=1\nB=2\n")
	ac := write(t, d, "ac", "A=1\n")
	var o, e bytes.Buffer
	if c := run([]string{"-optional", "B", ex, ac}, &o, &e); c != 0 {
		t.Fatalf("code %d %s", c, o.String())
	}
}
