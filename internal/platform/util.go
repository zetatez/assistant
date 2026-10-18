package platform

import (
	"os"
	"regexp"
	"strings"
	"time"
)

var macRe = regexp.MustCompile(`(?i)[0-9a-f]{2}([-:])[0-9a-f]{2}([-:][0-9a-f]{2}){4}`)

// ExtractMAC returns the first MAC address found in s, normalized to colons.
// Returns "" when no address is present.
func ExtractMAC(s string) string {
	m := macRe.FindString(s)
	if m == "" {
		return ""
	}
	return strings.ReplaceAll(m, "-", ":")
}

// shellQuote single-quotes s for safe interpolation into a bash command line.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// CommandLine joins argv into a shell command line with each argument quoted.
func CommandLine(argv []string) string {
	quoted := make([]string, len(argv))
	for i, a := range argv {
		quoted[i] = shellQuote(a)
	}
	return strings.Join(quoted, " ")
}

// Quote single-quotes s for interpolation into a shell command line.
func Quote(s string) string { return shellQuote(s) }

// appleQuote escapes s so it can be embedded inside an AppleScript string
// literal (double-quoted).
func appleQuote(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

// appleList renders a string slice as an AppleScript list literal.
func appleList(items []string) string {
	quoted := make([]string, 0, len(items))
	for _, it := range items {
		quoted = append(quoted, `"`+appleQuote(it)+`"`)
	}
	return "{" + strings.Join(quoted, ", ") + "}"
}

// writeTempList writes each item on its own line and returns the file path.
// Used for feeding option lists to interactive pickers.
func writeTempList(pattern string, items []string) (string, error) {
	f, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := f.WriteString(strings.Join(items, "\n")); err != nil {
		return "", err
	}
	if _, err := f.WriteString("\n"); err != nil {
		return "", err
	}
	return f.Name(), nil
}

// normalizeTTL converts a duration to whole seconds, defaulting to 3s.
func normalizeTTL(ttl time.Duration) int {
	if ttl <= 0 {
		return 3
	}
	return int(ttl.Seconds())
}
