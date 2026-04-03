// internal/tui/simple_highlighter.go
package tui

import (
	"regexp"
	"strings"
)

// ANSI color codes (foreground only, no background)
// Aura theme inspired colors
const (
	ansiReset   = "\x1b[0m"
	ansiResetFG = "\x1b[39m"               // Reset foreground only (preserve background)
	ansiGray    = "\x1b[38;2;108;108;108m" // Comments - muted gray
	ansiPink    = "\x1b[38;2;255;121;198m" // Keywords - soft pink
	ansiCyan    = "\x1b[38;2;139;233;253m" // Types - bright cyan
	ansiGreen   = "\x1b[38;2;80;250;123m"  // Functions/functions - vibrant green
	ansiPurple  = "\x1b[38;2;189;147;249m" // Numbers - soft purple
	ansiYellow  = "\x1b[38;2;241;250;140m" // Strings - warm yellow
	ansiOrange  = "\x1b[38;2;255;184;108m" // Special - soft orange
	ansiWhite   = "\x1b[38;2;248;248;242m" // Default - soft white
)

// SimpleHighlighter provides basic syntax highlighting without external dependencies
type SimpleHighlighter struct {
	// Language-specific patterns
	patterns map[string][]TokenPattern
}

// TokenPattern defines a token type and its regex pattern
type TokenPattern struct {
	Pattern   *regexp.Regexp
	ANSICode  string
	TokenType string
}

// NewSimpleHighlighter creates a new simple highlighter with predefined patterns
func NewSimpleHighlighter() *SimpleHighlighter {
	h := &SimpleHighlighter{
		patterns: make(map[string][]TokenPattern),
	}

	// Go patterns
	h.patterns["go"] = []TokenPattern{
		// Comments - must be first to avoid matching keywords inside comments
		{
			Pattern:  regexp.MustCompile(`(?m)//.*$`),
			ANSICode: ansiGray,
		},
		{
			Pattern:  regexp.MustCompile(`(?s)/\*.*?\*/`),
			ANSICode: ansiGray,
		},
		// Strings
		{
			Pattern:  regexp.MustCompile(`"(?:[^"\\]|\\.)*"`),
			ANSICode: ansiYellow,
		},
		{
			Pattern:  regexp.MustCompile("`(?:[^`]|\\`)*`"),
			ANSICode: ansiYellow,
		},
		// Keywords
		{
			Pattern:  regexp.MustCompile(`\b(?:package|import|func|var|const|type|struct|interface|map|chan|go|defer|return|if|else|for|range|switch|case|default|break|continue|goto|fallthrough|select)\b`),
			ANSICode: ansiPink,
		},
		// Built-in types
		{
			Pattern:  regexp.MustCompile(`\b(?:string|int|int8|int16|int32|int64|uint|uint8|uint16|uint32|uint64|float32|float64|bool|byte|rune|error|uintptr|complex64|complex128)\b`),
			ANSICode: ansiCyan,
		},
		// Built-in functions
		{
			Pattern:  regexp.MustCompile(`\b(?:append|cap|close|complex|copy|delete|imag|len|make|new|panic|print|println|real|recover)\b`),
			ANSICode: ansiGreen,
		},
		// Numbers
		{
			Pattern:  regexp.MustCompile(`\b(?:0x[0-9a-fA-F]+|0[0-7]*|[1-9][0-9]*(?:\.[0-9]+)?(?:[eE][+-]?[0-9]+)?)\b`),
			ANSICode: ansiPurple,
		},
		// Booleans and nil
		{
			Pattern:  regexp.MustCompile(`\b(?:true|false|nil)\b`),
			ANSICode: ansiPurple,
		},
	}

	// JSON patterns
	h.patterns["json"] = []TokenPattern{
		// Keys - strings followed by colon including the colon (cyan)
		// Pattern matches: "key":
		{
			Pattern:  regexp.MustCompile(`"(?:[^"\\]|\\.)*"\s*:`),
			ANSICode: ansiCyan,
		},
		// String values - all other strings (yellow)
		{
			Pattern:  regexp.MustCompile(`"(?:[^"\\]|\\.)*"`),
			ANSICode: ansiYellow,
		},
		// Numbers
		{
			Pattern:  regexp.MustCompile(`-?(?:0|[1-9]\d*)(?:\.\d+)?(?:[eE][+-]?\d+)?`),
			ANSICode: ansiPurple,
		},
		// Boolean and null
		{
			Pattern:  regexp.MustCompile(`\b(?:true|false|null)\b`),
			ANSICode: ansiPink,
		},
	}

	// YAML patterns
	h.patterns["yaml"] = []TokenPattern{
		// Comments
		{
			Pattern:  regexp.MustCompile(`(?m)#.*$`),
			ANSICode: ansiGray,
		},
		// Strings
		{
			Pattern:  regexp.MustCompile(`"(?:[^"\\]|\\.)*"`),
			ANSICode: ansiYellow,
		},
		{
			Pattern:  regexp.MustCompile("`(?:[^`]|\\`)*`"),
			ANSICode: ansiYellow,
		},
		// Keys (before colon)
		{
			Pattern:  regexp.MustCompile(`(?m)^\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*:`),
			ANSICode: ansiCyan,
		},
		// Numbers
		{
			Pattern:  regexp.MustCompile(`\b(?:0x[0-9a-fA-F]+|0[0-7]*|[1-9][0-9]*(?:\.[0-9]+)?(?:[eE][+-]?[0-9]+)?)\b`),
			ANSICode: ansiPurple,
		},
		// Booleans
		{
			Pattern:  regexp.MustCompile(`\b(?:true|false|yes|no|on|off)\b`),
			ANSICode: ansiPink,
		},
	}

	return h
}

// Highlight highlights code for a given language
func (h *SimpleHighlighter) Highlight(code, language string) string {
	if code == "" {
		return ""
	}

	// Map language names
	langMap := map[string]string{
		"go":     "go",
		"golang": "go",
		"json":   "json",
		"yaml":   "yaml",
		"yml":    "yaml",
	}

	lang := language
	if mapped, ok := langMap[language]; ok {
		lang = mapped
	}

	patterns, ok := h.patterns[lang]
	if !ok {
		// No patterns for this language, return plain text
		return code
	}

	return h.highlightWithPatterns(code, patterns)
}

// highlightWithPatterns applies highlighting patterns to code using ANSI codes
type highlightRange struct {
	start    int
	end      int
	ansiCode string
}

func (h *SimpleHighlighter) highlightWithPatterns(code string, patterns []TokenPattern) string {
	var ranges []highlightRange

	// Find all matches for each pattern
	for _, pattern := range patterns {
		matches := pattern.Pattern.FindAllStringIndex(code, -1)
		for _, match := range matches {
			if match == nil {
				continue
			}
			// Check if this range overlaps with existing ranges
			overlaps := false
			for _, r := range ranges {
				if match[0] < r.end && match[1] > r.start {
					overlaps = true
					break
				}
			}
			if !overlaps {
				ranges = append(ranges, highlightRange{
					start:    match[0],
					end:      match[1],
					ansiCode: pattern.ANSICode,
				})
			}
		}
	}

	// Sort ranges by start position
	for i := 0; i < len(ranges); i++ {
		for j := i + 1; j < len(ranges); j++ {
			if ranges[j].start < ranges[i].start {
				ranges[i], ranges[j] = ranges[j], ranges[i]
			}
		}
	}

	// Build the highlighted string with ANSI codes
	var result strings.Builder
	lastEnd := 0
	for _, r := range ranges {
		// Add plain text before this range
		if r.start > lastEnd {
			result.WriteString(code[lastEnd:r.start])
		}
		// Add highlighted text with ANSI codes (foreground only, no background)
		text := code[r.start:r.end]
		result.WriteString(r.ansiCode)
		result.WriteString(text)
		result.WriteString(ansiResetFG) // Reset foreground only, preserve background
		lastEnd = r.end
	}
	// Add remaining plain text
	if lastEnd < len(code) {
		result.WriteString(code[lastEnd:])
	}

	return result.String()
}
