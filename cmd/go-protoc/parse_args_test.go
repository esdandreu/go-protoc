package main

import (
	"reflect"
	"testing"
)

func TestParseArgs(t *testing.T) {
	testCases := map[string]struct {
		args           []string
		flags          []string
		hasNonFlagArgs bool
	}{
		"no args": {
			args:           []string{"--version", "--go_out=whatever"},
			flags:          []string{"version", "go_out"},
			hasNonFlagArgs: false,
		},
		"no flags": {
			args:           []string{"whatever"},
			flags:          nil,
			hasNonFlagArgs: true,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			flags, hasNonFlagArgs := ParseArgs(tc.args)
			if !reflect.DeepEqual(flags, tc.flags) {
				t.Errorf("Expected %v, got %v", tc.flags, flags)
			}
			if hasNonFlagArgs != tc.hasNonFlagArgs {
				t.Errorf("Expected %v, got %v", tc.hasNonFlagArgs, hasNonFlagArgs)
			}
		})
	}
}
