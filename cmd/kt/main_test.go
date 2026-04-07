package main

import (
	"reflect"
	"testing"
)

func TestSplitArgs(t *testing.T) {
	cases := []struct {
		name       string
		args       []string
		positional []string
		flags      []string
	}{
		{
			name:       "empty",
			args:       []string{},
			positional: nil,
			flags:      nil,
		},
		{
			name:       "only positional",
			args:       []string{"karpathy"},
			positional: []string{"karpathy"},
			flags:      nil,
		},
		{
			// "20" doesn't start with '-', so it's classified as positional.
			// flag.FlagSet.Parse handles the value pairing internally.
			name:       "flag with value",
			args:       []string{"--json", "--limit", "20"},
			positional: []string{"20"},
			flags:      []string{"--json", "--limit"},
		},
		{
			name:       "positional before flags",
			args:       []string{"42", "--markdown"},
			positional: []string{"42"},
			flags:      []string{"--markdown"},
		},
		{
			name:       "flags before positional",
			args:       []string{"--markdown", "42"},
			positional: []string{"42"},
			flags:      []string{"--markdown"},
		},
		{
			name:       "mixed order",
			args:       []string{"--limit", "10", "karpathy", "--json"},
			positional: []string{"10", "karpathy"},
			flags:      []string{"--limit", "--json"},
		},
		{
			name:       "single dash flag",
			args:       []string{"-json"},
			positional: nil,
			flags:      []string{"-json"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotPos, gotFlags := splitArgs(tc.args)
			if !reflect.DeepEqual(gotPos, tc.positional) {
				t.Errorf("positional: got %v, want %v", gotPos, tc.positional)
			}
			if !reflect.DeepEqual(gotFlags, tc.flags) {
				t.Errorf("flags: got %v, want %v", gotFlags, tc.flags)
			}
		})
	}
}
