package goctionutil

import (
	"fmt"
	"strings"
	"unicode"
)

// GenerateGoctionTemplate generates a template for a new goction
func GenerateGoctionTemplate(name string) string {
	return fmt.Sprintf(`package main

import (
	"fmt"
	"strings"
)

// %s is a goction that does something
func %s(args ...string) (string, error) {
	// TODO: Implement your goction logic here
	return fmt.Sprintf("Goction %s executed with args: %%s", strings.Join(args, ", ")), nil
}
`, TitleCase(name), TitleCase(name), name)
}

// TitleCase converts a string to title case
func TitleCase(s string) string {
	return strings.Title(strings.ToLower(s))
}

// CamelCase converts a string to camel case
func CamelCase(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}

	n := strings.Builder{}
	n.Grow(len(s))
	capNext := true
	for i, v := range []rune(s) {
		if unicode.IsLetter(v) || unicode.IsDigit(v) {
			if capNext {
				n.WriteRune(unicode.ToUpper(v))
				capNext = false
			} else {
				n.WriteRune(v)
			}
			continue
		}
		if v == '_' || v == ' ' || v == '-' {
			capNext = true
		}
		if i == 0 {
			capNext = false
		}
	}
	return n.String()
}

// ValidateGoctionName checks if a goction name is valid
func ValidateGoctionName(name string) error {
	if name == "" {
		return fmt.Errorf("goction name cannot be empty")
	}
	if !unicode.IsLetter(rune(name[0])) {
		return fmt.Errorf("goction name must start with a letter")
	}
	for _, char := range name {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '_' {
			return fmt.Errorf("goction name can only contain letters, digits, and underscores")
		}
	}
	return nil
}

// FormatDuration formats a duration in a human-readable way
func FormatDuration(d float64) string {
	if d < 1e-6 {
		return fmt.Sprintf("%.2f ns", d*1e9)
	} else if d < 1e-3 {
		return fmt.Sprintf("%.2f µs", d*1e6)
	} else if d < 1 {
		return fmt.Sprintf("%.2f ms", d*1e3)
	}
	return fmt.Sprintf("%.2f s", d)
}
