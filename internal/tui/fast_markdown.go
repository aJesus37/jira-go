// internal/tui/fast_markdown.go
package tui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Styles for fast markdown rendering
	h1Style = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FF79C6")).
		MarginTop(1)

	h2Style = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#8BE9FD")).
		MarginTop(1)

	h3Style = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#50FA7B"))

	boldStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F8F8F2"))

	italicStyle = lipgloss.NewStyle().
			Italic(true).
			Foreground(lipgloss.Color("#F8F8F2"))

	codeStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#44475A")).
			Foreground(lipgloss.Color("#F8F8F2")).
			Padding(0, 1)

	codeBlockStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#282A36")).
			Foreground(lipgloss.Color("#F8F8F2")).
			Padding(1, 2)

	linkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8BE9FD")).
			Underline(true)

	blockquoteStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4")).
			BorderLeft(true).
			BorderForeground(lipgloss.Color("#6272A4")).
			PaddingLeft(1)

	listStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#BD93F9"))
)

// FastMarkdownRenderer is a lightweight markdown renderer for speed
type FastMarkdownRenderer struct {
	width uint
}

// NewFastMarkdownRenderer creates a new fast markdown renderer
func NewFastMarkdownRenderer(width uint) *FastMarkdownRenderer {
	return &FastMarkdownRenderer{width: width}
}

// SetWidth updates the renderer width
func (r *FastMarkdownRenderer) SetWidth(width uint) {
	r.width = width
}

// Render quickly renders markdown with basic formatting
func (r *FastMarkdownRenderer) Render(content string) (string, error) {
	if content == "" {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#808080")).Render("(no description)"), nil
	}

	lines := strings.Split(content, "\n")
	var result strings.Builder

	inCodeBlock := false
	var codeBlockContent strings.Builder

	for i, line := range lines {
		// Handle code blocks
		if strings.HasPrefix(line, "```") {
			if inCodeBlock {
				// End code block
				result.WriteString(codeBlockStyle.Render(codeBlockContent.String()))
				result.WriteString("\n")
				codeBlockContent.Reset()
				inCodeBlock = false
			} else {
				// Start code block
				inCodeBlock = true
			}
			continue
		}

		if inCodeBlock {
			codeBlockContent.WriteString(line)
			codeBlockContent.WriteString("\n")
			continue
		}

		// Render the line
		rendered := r.renderLine(line)
		result.WriteString(rendered)

		// Add newline if not last line
		if i < len(lines)-1 {
			result.WriteString("\n")
		}
	}

	return result.String(), nil
}

// renderLine renders a single line of markdown
func (r *FastMarkdownRenderer) renderLine(line string) string {
	// Headers
	if strings.HasPrefix(line, "# ") {
		return h1Style.Render(line[2:])
	}
	if strings.HasPrefix(line, "## ") {
		return h2Style.Render(line[3:])
	}
	if strings.HasPrefix(line, "### ") {
		return h3Style.Render(line[4:])
	}
	if strings.HasPrefix(line, "#### ") {
		return boldStyle.Render(line[5:])
	}

	// Horizontal rule
	if strings.TrimSpace(line) == "---" || strings.TrimSpace(line) == "***" {
		return strings.Repeat("─", int(r.width))
	}

	// Blockquote
	if strings.HasPrefix(line, "> ") {
		return blockquoteStyle.Render(line[2:])
	}

	// List items - preserve the bullet/number but style the content
	if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
		content := r.renderInline(line[2:])
		return listStyle.Render("• ") + content
	}
	if matched, _ := regexp.MatchString(`^\d+\.\s`, line); matched {
		parts := strings.SplitN(line, ". ", 2)
		if len(parts) == 2 {
			content := r.renderInline(parts[1])
			return listStyle.Render(parts[0]+". ") + content
		}
	}

	// Regular line with inline formatting
	return r.renderInline(line)
}

// renderInline handles inline formatting (bold, italic, code, links)
func (r *FastMarkdownRenderer) renderInline(text string) string {
	// Code spans (must be first to avoid processing content inside)
	text = renderCodeSpans(text)

	// Bold and italic
	text = renderBold(text)
	text = renderItalic(text)

	// Strikethrough
	text = renderStrikethrough(text)

	// Links
	text = renderLinks(text)

	return text
}

// renderCodeSpans handles inline code
func renderCodeSpans(text string) string {
	re := regexp.MustCompile("`([^`]+)`")
	return re.ReplaceAllStringFunc(text, func(match string) string {
		content := match[1 : len(match)-1]
		return codeStyle.Render(content)
	})
}

// renderBold handles **bold** and __bold__
func renderBold(text string) string {
	// **bold**
	re := regexp.MustCompile(`\*\*([^*]+)\*\*`)
	text = re.ReplaceAllStringFunc(text, func(match string) string {
		content := match[2 : len(match)-2]
		return boldStyle.Render(content)
	})

	// __bold__
	re = regexp.MustCompile(`__([^_]+)__`)
	return re.ReplaceAllStringFunc(text, func(match string) string {
		content := match[2 : len(match)-2]
		return boldStyle.Render(content)
	})
}

// renderItalic handles *italic* and _italic_
func renderItalic(text string) string {
	// *italic* (but not ** which is bold)
	re := regexp.MustCompile(`\*([^*]+)\*`)
	text = re.ReplaceAllStringFunc(text, func(match string) string {
		content := match[1 : len(match)-1]
		return italicStyle.Render(content)
	})

	// _italic_ (but not __ which is bold)
	re = regexp.MustCompile(`_([^_]+)_`)
	return re.ReplaceAllStringFunc(text, func(match string) string {
		content := match[1 : len(match)-1]
		return italicStyle.Render(content)
	})
}

// renderStrikethrough handles ~~strikethrough~~
func renderStrikethrough(text string) string {
	re := regexp.MustCompile(`~~([^~]+)~~`)
	return re.ReplaceAllStringFunc(text, func(match string) string {
		content := match[2 : len(match)-2]
		return lipgloss.NewStyle().Strikethrough(true).Render(content)
	})
}

// renderLinks handles [text](url)
func renderLinks(text string) string {
	re := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	return re.ReplaceAllStringFunc(text, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) == 3 {
			return linkStyle.Render(parts[1])
		}
		return match
	})
}

// RenderADF renders Atlassian Document Format as markdown (same interface as before)
func (r *FastMarkdownRenderer) RenderADF(ADF interface{}) (string, error) {
	// Convert ADF to markdown first
	markdown := adfToMarkdownFast(ADF)
	return r.Render(markdown)
}

// adfToMarkdownFast converts ADF to markdown string
func adfToMarkdownFast(ADF interface{}) string {
	if ADF == nil {
		return ""
	}

	// Handle string type (plain text description)
	if text, ok := ADF.(string); ok {
		return text
	}

	// Handle map type (ADF object)
	if doc, ok := ADF.(map[string]interface{}); ok {
		var result strings.Builder

		// Process content
		if content, ok := doc["content"].([]interface{}); ok {
			for _, node := range content {
				if nodeMap, ok := node.(map[string]interface{}); ok {
					result.WriteString(renderADFNodeFast(nodeMap))
					result.WriteString("\n")
				}
			}
		}

		return result.String()
	}

	return ""
}

// renderADFNodeFast renders a single ADF node
func renderADFNodeFast(node map[string]interface{}) string {
	nodeType, _ := node["type"].(string)

	switch nodeType {
	case "paragraph":
		return renderADFTextFast(node)
	case "heading":
		attrs, _ := node["attrs"].(map[string]interface{})
		level, _ := attrs["level"].(float64)
		prefix := strings.Repeat("#", int(level))
		return prefix + " " + renderADFTextFast(node)
	case "bulletList":
		return renderADFListFast(node, "bullet")
	case "orderedList":
		return renderADFListFast(node, "ordered")
	case "codeBlock":
		content := renderADFTextFast(node)
		return "```\n" + content + "\n```"
	case "blockquote":
		content := renderADFTextFast(node)
		lines := strings.Split(content, "\n")
		var result strings.Builder
		for _, line := range lines {
			result.WriteString("> ")
			result.WriteString(line)
			result.WriteString("\n")
		}
		return result.String()
	default:
		return renderADFTextFast(node)
	}
}

// renderADFTextFast extracts text from an ADF node
func renderADFTextFast(node map[string]interface{}) string {
	var result strings.Builder

	if content, ok := node["content"].([]interface{}); ok {
		for _, child := range content {
			if childMap, ok := child.(map[string]interface{}); ok {
				childType, _ := childMap["type"].(string)

				switch childType {
				case "text":
					text, _ := childMap["text"].(string)
					// Handle marks (bold, italic, etc.)
					if marks, ok := childMap["marks"].([]interface{}); ok {
						text = applyMarksFast(text, marks)
					}
					result.WriteString(text)
				case "hardBreak":
					result.WriteString("\n")
				case "inlineCard", "mention":
					// Handle special inline elements
					if attrs, ok := childMap["attrs"].(map[string]interface{}); ok {
						if text, ok := attrs["text"].(string); ok {
							result.WriteString(text)
						}
					}
				default:
					// Recursively render nested nodes
					result.WriteString(renderADFNodeFast(childMap))
				}
			}
		}
	}

	return result.String()
}

// renderADFListFast renders list nodes
func renderADFListFast(node map[string]interface{}, listType string) string {
	var result strings.Builder

	if content, ok := node["content"].([]interface{}); ok {
		for i, child := range content {
			if childMap, ok := child.(map[string]interface{}); ok {
				if childMap["type"] == "listItem" {
					prefix := "- "
					if listType == "ordered" {
						prefix = fmt.Sprintf("%d. ", i+1)
					}
					result.WriteString(prefix)
					result.WriteString(renderADFTextFast(childMap))
					result.WriteString("\n")
				}
			}
		}
	}

	return result.String()
}

// applyMarksFast applies formatting marks to text
func applyMarksFast(text string, marks []interface{}) string {
	for _, mark := range marks {
		if markMap, ok := mark.(map[string]interface{}); ok {
			markType, _ := markMap["type"].(string)
			switch markType {
			case "strong":
				text = "**" + text + "**"
			case "em":
				text = "*" + text + "*"
			case "code":
				text = "`" + text + "`"
			case "strike":
				text = "~~" + text + "~~"
			case "link":
				if attrs, ok := markMap["attrs"].(map[string]interface{}); ok {
					href, _ := attrs["href"].(string)
					text = fmt.Sprintf("[%s](%s)", text, href)
				}
			}
		}
	}
	return text
}
