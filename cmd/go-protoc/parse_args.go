package main

import "strings"

// ParseArgs is a basic flag parser that never fails on unrecognized input. It
// returns a slice of flag names that were set and a boolean indicating whether
// any non-flag arguments were present.
func ParseArgs(args []string) ([]string, bool) {
	var setFlags []string
	hasNonFlagArgs := false

	i := 0
	for i < len(args) {
		arg := args[i]

		// Check if it's a flag (starts with - or --)
		if len(arg) >= 2 && arg[0] == '-' {
			// Handle "--" terminator - everything after is treated as non-flag arguments
			if arg == "--" {
				if i+1 < len(args) {
					hasNonFlagArgs = true
				}
				break
			}

			flagName := extractFlagName(arg)
			if flagName != "" {
				setFlags = append(setFlags, flagName)

				// Skip next argument if it looks like a flag value
				// (doesn't start with - and we haven't seen -- yet)
				if i+1 < len(args) {
					nextArg := args[i+1]
					if !strings.HasPrefix(nextArg, "-") {
						// This could be a flag value, skip it
						i++
					}
				}
			}
		} else {
			// It's a non-flag argument
			hasNonFlagArgs = true
		}
		i++
	}

	return setFlags, hasNonFlagArgs
}

// extractFlagName extracts the flag name from a flag argument Handles both
// -flag, --flag, -flag=value, --flag=value formats
func extractFlagName(arg string) string {
	// Remove leading dashes
	name := arg
	if strings.HasPrefix(name, "--") {
		name = name[2:]
	} else if strings.HasPrefix(name, "-") {
		name = name[1:]
	}

	// Handle flag=value format
	if eqIndex := strings.Index(name, "="); eqIndex >= 0 {
		name = name[:eqIndex]
	}

	// Return empty string if invalid flag name
	if name == "" || strings.HasPrefix(name, "-") {
		return ""
	}

	return name
}
