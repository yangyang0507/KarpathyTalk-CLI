package main

import (
	"flag"
	"reflect"
	"testing"
)

// newFS creates a FlagSet that mirrors the flags defined in runPost (the most
// flag-heavy command) so we can exercise splitArgs with realistic inputs.
func newPostFS() *flag.FlagSet {
	fs := flag.NewFlagSet("post", flag.ContinueOnError)
	fs.Int("limit", 20, "")
	fs.Bool("json", false, "")
	fs.Bool("markdown", false, "")
	fs.Bool("raw", false, "")
	fs.Int("revision", 0, "")
	return fs
}

func newUserFS() *flag.FlagSet {
	fs := flag.NewFlagSet("user", flag.ContinueOnError)
	fs.Bool("replies", false, "")
	fs.Int("limit", 20, "")
	fs.Int64("before", 0, "")
	fs.Bool("json", false, "")
	fs.Bool("markdown", false, "")
	return fs
}

func TestSplitArgs_Empty(t *testing.T) {
	fs := newPostFS()
	got := splitArgs(fs, []string{})
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestSplitArgs_OnlyPositional(t *testing.T) {
	fs := newPostFS()
	got := splitArgs(fs, []string{"42"})
	want := []string{"42"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestSplitArgs_BoolFlagOnly(t *testing.T) {
	fs := newPostFS()
	got := splitArgs(fs, []string{"--markdown"})
	if len(got) != 0 {
		t.Errorf("expected no positionals, got %v", got)
	}
	if v := flagBool(fs, "markdown"); !v {
		t.Error("--markdown should be true")
	}
}

// flagBool reads a bool flag value from a FlagSet.
func flagBool(fs *flag.FlagSet, name string) bool {
	f := fs.Lookup(name)
	return f != nil && f.Value.String() == "true"
}

// flagStr reads a string flag value from a FlagSet.
func flagStr(fs *flag.FlagSet, name string) string {
	f := fs.Lookup(name)
	if f == nil {
		return ""
	}
	return f.Value.String()
}

func TestSplitArgs_ValueFlagWithSpace(t *testing.T) {
	// --limit 50: "50" must stay with --limit, not become positional
	fs := newPostFS()
	got := splitArgs(fs, []string{"--limit", "50"})
	if len(got) != 0 {
		t.Errorf("expected no positionals, got %v", got)
	}
	if v := flagStr(fs, "limit"); v != "50" {
		t.Errorf("--limit: got %q, want %q", v, "50")
	}
}

func TestSplitArgs_PositionalBeforeFlag(t *testing.T) {
	// "kt post 42 --markdown"
	fs := newPostFS()
	got := splitArgs(fs, []string{"42", "--markdown"})
	want := []string{"42"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("positional: got %v, want %v", got, want)
	}
	if v := flagBool(fs, "markdown"); !v {
		t.Error("--markdown should be true")
	}
}

func TestSplitArgs_FlagBeforePositional(t *testing.T) {
	// "kt post --markdown 42"
	fs := newPostFS()
	got := splitArgs(fs, []string{"--markdown", "42"})
	want := []string{"42"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("positional: got %v, want %v", got, want)
	}
}

func TestSplitArgs_MixedWithValueFlag(t *testing.T) {
	// "kt user --limit 10 karpathy --json"
	fs := newUserFS()
	got := splitArgs(fs, []string{"--limit", "10", "karpathy", "--json"})
	want := []string{"karpathy"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("positional: got %v, want %v", got, want)
	}
	if v := flagStr(fs, "limit"); v != "10" {
		t.Errorf("--limit: got %q, want %q", v, "10")
	}
	if v := flagBool(fs, "json"); !v {
		t.Error("--json should be true")
	}
}

func TestSplitArgs_EqualSignForm(t *testing.T) {
	// --limit=50 (value embedded with =)
	fs := newPostFS()
	got := splitArgs(fs, []string{"--limit=50"})
	if len(got) != 0 {
		t.Errorf("expected no positionals, got %v", got)
	}
	if v := flagStr(fs, "limit"); v != "50" {
		t.Errorf("--limit: got %q, want %q", v, "50")
	}
}
