package issue

import (
	"errors"
	"fmt"
	"strings"
)

// ReplaceOnce replaces exactly one occurrence of old with new in text.
// Returns an error if old is empty, not found, or found multiple times.
// The new string can be empty to delete the matched text.
func ReplaceOnce(text, old, new string) (string, error) {
	if old == "" {
		return "", errors.New("old text cannot be empty")
	}
	count := strings.Count(text, old)
	if count == 0 {
		return "", errors.New("text not found in body")
	}
	if count > 1 {
		return "", fmt.Errorf("text found %d times in body (must be unique)", count)
	}
	return strings.Replace(text, old, new, 1), nil
}

// CheckItem finds an unchecked checkbox line (- [ ]) matching substr
// (case-insensitive) and checks it. Returns error if no match or ambiguous.
func CheckItem(text, substr string) (string, error) {
	return toggleCheckbox(text, substr, false)
}

// UncheckItem finds a checked checkbox line (- [x]) matching substr
// (case-insensitive) and unchecks it. Returns error if no match or ambiguous.
func UncheckItem(text, substr string) (string, error) {
	return toggleCheckbox(text, substr, true)
}

func toggleCheckbox(text, substr string, uncheck bool) (string, error) {
	if substr == "" {
		return "", errors.New("search text cannot be empty")
	}

	var fromPrefix, toPrefix, stateLabel string
	if uncheck {
		fromPrefix = "- [x] "
		toPrefix = "- [ ] "
		stateLabel = "checked"
	} else {
		fromPrefix = "- [ ] "
		toPrefix = "- [x] "
		stateLabel = "unchecked"
	}

	lines := strings.Split(text, "\n")
	lowerSubstr := strings.ToLower(substr)
	var matches []int

	for i, line := range lines {
		trimmed := strings.TrimLeft(line, " \t")
		if strings.HasPrefix(trimmed, fromPrefix) && strings.Contains(strings.ToLower(trimmed), lowerSubstr) {
			matches = append(matches, i)
		}
	}

	if len(matches) == 0 {
		return "", fmt.Errorf("no %s item matching %q", stateLabel, substr)
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("%d %s items match %q (must be unique)", len(matches), stateLabel, substr)
	}

	idx := matches[0]
	line := lines[idx]
	trimmed := strings.TrimLeft(line, " \t")
	indent := line[:len(line)-len(trimmed)]
	lines[idx] = indent + toPrefix + trimmed[len(fromPrefix):]

	return strings.Join(lines, "\n"), nil
}

// HasIncompleteChecklist returns true if text contains at least one
// unchecked checkbox (- [ ]) line. Handles optional leading whitespace.
func HasIncompleteChecklist(text string) bool {
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimLeft(line, " \t")
		if strings.HasPrefix(trimmed, "- [ ] ") {
			return true
		}
	}
	return false
}

// AppendWithSeparator appends addition to text with a blank line separator.
// If text is empty, returns addition without separator.
// If addition is empty, returns text unchanged (no-op).
func AppendWithSeparator(text, addition string) string {
	if addition == "" {
		return text
	}
	if text == "" {
		return addition
	}
	// Ensure single newline separator
	text = strings.TrimRight(text, "\n")
	return text + "\n\n" + addition
}
