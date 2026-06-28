// Package debug provides opt-in, stderr-only diagnostic logging for the CLI.
// It is silent unless MOZEIDON_DEBUG is set to a truthy value, so it never
// pollutes stdout (which carries the JSON/command output consumers parse).
package debug

import (
	"fmt"
	"os"
)

// enabled is evaluated once: MOZEIDON_DEBUG set to anything other than "",
// "0", or "false" turns logging on.
var enabled = func() bool {
	switch os.Getenv("MOZEIDON_DEBUG") {
	case "", "0", "false":
		return false
	default:
		return true
	}
}()

// Enabled reports whether debug logging is on.
func Enabled() bool { return enabled }

// Logf writes a single diagnostic line to stderr when enabled; no-op otherwise.
func Logf(format string, args ...any) {
	if !enabled {
		return
	}
	fmt.Fprintf(os.Stderr, "[mozeidon-z] "+format+"\n", args...)
}
