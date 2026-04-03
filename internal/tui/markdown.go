// internal/tui/markdown.go
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// MarkdownRenderer handles markdown rendering with Glow styles
type MarkdownRenderer struct {
	renderer *glamour.TermRenderer
	width    uint
}

// NewMarkdownRenderer creates a new markdown renderer
func NewMarkdownRenderer(width uint) (*MarkdownRenderer, error) {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(int(width)),
	)
	if err != nil {
		return nil, fmt.Errorf("creating markdown renderer: %w", err)
	}

	return &MarkdownRenderer{
		renderer: renderer,
		width:    width,
	}, nil
}

// Render renders markdown content to styled string
func (m *MarkdownRenderer) Render(content string) (string, error) {
	if content == "" {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#808080")).Render("(no description)"), nil
	}

	// Trim whitespace
	content = strings.TrimSpace(content)

	// Render markdown
	rendered, err := m.renderer.Render(content)
	if err != nil {
		// If rendering fails, return plain text
		return content, nil
	}

	return rendered, nil
}

// RenderADF renders Atlassian Document Format as markdown
func (m *MarkdownRenderer) RenderADF(ADF interface{}) (string, error) {
	// Convert ADF to markdown-like text
	// This is a simplified conversion
	markdown := m.adfToMarkdown(ADF)
	return m.Render(markdown)
}

// adfToMarkdown converts ADF to markdown string
func (m *MarkdownRenderer) adfToMarkdown(ADF interface{}) string {
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
					result.WriteString(m.renderADFNode(nodeMap))
					result.WriteString("\n")
				}
			}
		}

		return result.String()
	}

	return ""
}

// renderADFNode renders a single ADF node
func (m *MarkdownRenderer) renderADFNode(node map[string]interface{}) string {
	nodeType, _ := node["type"].(string)

	switch nodeType {
	case "paragraph":
		return m.renderADFText(node)
	case "heading":
		attrs, _ := node["attrs"].(map[string]interface{})
		level, _ := attrs["level"].(float64)
		prefix := strings.Repeat("#", int(level))
		return prefix + " " + m.renderADFText(node)
	case "bulletList":
		return m.renderADFList(node, "bullet")
	case "orderedList":
		return m.renderADFList(node, "ordered")
	case "codeBlock":
		content := m.renderADFText(node)
		return "```\n" + content + "\n```"
	case "blockquote":
		content := m.renderADFText(node)
		lines := strings.Split(content, "\n")
		var result strings.Builder
		for _, line := range lines {
			result.WriteString("> ")
			result.WriteString(line)
			result.WriteString("\n")
		}
		return result.String()
	default:
		return m.renderADFText(node)
	}
}

// renderADFText extracts text from an ADF node
func (m *MarkdownRenderer) renderADFText(node map[string]interface{}) string {
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
						text = m.applyMarks(text, marks)
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
					result.WriteString(m.renderADFNode(childMap))
				}
			}
		}
	}

	return result.String()
}

// renderADFList renders list nodes
func (m *MarkdownRenderer) renderADFList(node map[string]interface{}, listType string) string {
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
					result.WriteString(m.renderADFText(childMap))
					result.WriteString("\n")
				}
			}
		}
	}

	return result.String()
}

// applyMarks applies formatting marks to text
func (m *MarkdownRenderer) applyMarks(text string, marks []interface{}) string {
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

// SetWidth updates the renderer width
func (m *MarkdownRenderer) SetWidth(width uint) {
	m.width = width
	// Recreate renderer with new width
	m.renderer, _ = glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(int(width)),
	)
}
