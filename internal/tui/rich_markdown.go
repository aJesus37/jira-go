// internal/tui/rich_markdown.go
package tui

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extAST "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// RichMarkdownRenderer provides full markdown support with syntax highlighting
type RichMarkdownRenderer struct {
	width       uint
	buffer      bytes.Buffer
	highlighter *SimpleHighlighter
	bgColor     string // Background color for code blocks (hex format)
}

// NewRichMarkdownRenderer creates a new rich markdown renderer
func NewRichMarkdownRenderer(width uint) *RichMarkdownRenderer {
	return &RichMarkdownRenderer{
		width:       width,
		highlighter: NewSimpleHighlighter(),
		bgColor:     "#1a1a2e", // Default dark background
	}
}

// SetBackgroundColor sets the background color for code blocks and inline code
func (r *RichMarkdownRenderer) SetBackgroundColor(color string) {
	r.bgColor = color
}

// SetWidth updates the renderer width
func (r *RichMarkdownRenderer) SetWidth(width uint) {
	r.width = width
}

// Render renders markdown with full feature support
func (r *RichMarkdownRenderer) Render(content string) (string, error) {
	if content == "" {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#6c6c6c")).Render("(no description)"), nil
	}

	r.buffer.Reset()

	// Parse markdown with goldmark
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.Table,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)

	doc := md.Parser().Parse(text.NewReader([]byte(content)))

	// Walk the AST and render
	err := r.renderNode(doc, []byte(content), 0)
	if err != nil {
		// Fallback to fast renderer on error
		fast := NewFastMarkdownRenderer(r.width)
		return fast.Render(content)
	}

	return r.buffer.String(), nil
}

// renderNode recursively renders AST nodes
func (r *RichMarkdownRenderer) renderNode(n ast.Node, source []byte, depth int) error {
	if n == nil {
		return nil
	}

	switch node := n.(type) {
	case *ast.Document:
		return r.renderChildren(node, source, depth)

	case *ast.Heading:
		r.renderHeading(node, source)
		return nil

	case *ast.Paragraph:
		r.renderParagraph(node, source)
		return nil

	case *ast.Text:
		r.renderText(node, source)
		return nil

	case *ast.String:
		r.buffer.WriteString(string(node.Value))
		return nil

	case *ast.CodeBlock:
		r.renderCodeBlock(node, source)
		return nil

	case *ast.FencedCodeBlock:
		r.renderFencedCodeBlock(node, source)
		return nil

	case *ast.CodeSpan:
		r.renderCodeSpan(node, source)
		return nil

	case *ast.Emphasis:
		r.renderEmphasis(node, source)
		return nil

	case *ast.List:
		r.renderList(node, source, depth)
		return nil

	case *ast.ListItem:
		r.renderListItem(node, source, depth)
		return nil

	case *ast.Blockquote:
		r.renderBlockquote(node, source, depth)
		return nil

	case *ast.ThematicBreak:
		r.buffer.WriteString(strings.Repeat("─", int(r.width)) + "\n")
		return nil

	case *ast.Link:
		r.renderLink(node, source)
		return nil

	case *ast.AutoLink:
		r.renderAutoLink(node, source)
		return nil

	case *ast.Image:
		r.renderImage(node, source)
		return nil

	case *extAST.Table:
		r.renderTable(node, source)
		return nil

	case *extAST.TableHeader:
		// Handled in renderTable
		return nil

	case *extAST.TableRow:
		// Handled in renderTable
		return nil

	case *extAST.TableCell:
		// Handled in renderTable
		return nil

	case *ast.RawHTML:
		// Skip raw HTML
		return nil

	case *ast.HTMLBlock:
		// Skip HTML blocks
		return nil

	default:
		return r.renderChildren(node, source, depth)
	}
}

// renderChildren renders all child nodes
func (r *RichMarkdownRenderer) renderChildren(n ast.Node, source []byte, depth int) error {
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if err := r.renderNode(c, source, depth); err != nil {
			return err
		}
	}
	return nil
}

// renderHeading renders a heading
func (r *RichMarkdownRenderer) renderHeading(n *ast.Heading, source []byte) {
	text := r.extractText(n, source)

	var style lipgloss.Style
	switch n.Level {
	case 1:
		style = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ff7edb")).MarginTop(1) // Aura pink
	case 2:
		style = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7ee7ff")).MarginTop(1) // Aura cyan
	case 3:
		style = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7ee787")) // Aura green
	default:
		style = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#a277ff")) // Aura purple
	}

	r.buffer.WriteString(style.Render(text) + "\n")
}

// renderParagraph renders a paragraph
func (r *RichMarkdownRenderer) renderParagraph(n *ast.Paragraph, source []byte) {
	text := r.extractText(n, source)
	if text != "" {
		r.buffer.WriteString(text + "\n")
	}
}

// renderText renders plain text
func (r *RichMarkdownRenderer) renderText(n *ast.Text, source []byte) {
	segment := n.Segment
	text := string(segment.Value(source))
	// Don't replace newlines here, handle them based on context
	r.buffer.WriteString(text)
}

// renderCodeBlock renders a code block with syntax highlighting
func (r *RichMarkdownRenderer) renderCodeBlock(n *ast.CodeBlock, source []byte) {
	// Get code from all lines
	var code bytes.Buffer
	for i := 0; i < n.Lines().Len(); i++ {
		line := n.Lines().At(i)
		code.Write(line.Value(source))
	}
	r.renderHighlightedCode(code.String(), "", false)
}

// renderFencedCodeBlock renders a fenced code block with syntax highlighting
func (r *RichMarkdownRenderer) renderFencedCodeBlock(n *ast.FencedCodeBlock, source []byte) {
	language := string(n.Language(source))

	// Get code from all lines
	var code bytes.Buffer
	for i := 0; i < n.Lines().Len(); i++ {
		line := n.Lines().At(i)
		code.Write(line.Value(source))
	}

	r.renderHighlightedCode(code.String(), language, true)
}

// renderHighlightedCode renders code with custom syntax highlighting
func (r *RichMarkdownRenderer) renderHighlightedCode(code, language string, isFenced bool) {
	if code == "" {
		return
	}

	// Trim trailing newline
	code = strings.TrimSuffix(code, "\n")

	// Use custom highlighter (adds foreground ANSI codes)
	highlighted := r.highlighter.Highlight(code, language)

	// Use configurable background color
	bgColor := r.bgColor
	if bgColor == "" {
		bgColor = "#1a1a2e"
	}
	// Parse hex color to RGB
	var rVal, gVal, bVal int
	fmt.Sscanf(bgColor, "#%02x%02x%02x", &rVal, &gVal, &bVal)
	ansiBG := fmt.Sprintf("\x1b[48;2;%d;%d;%dm", rVal, gVal, bVal)
	ansiReset := "\x1b[0m"

	// Calculate padding width (viewer width minus margins)
	// We want the block to span the full width
	blockWidth := int(r.width) - 4 // Account for margins
	if blockWidth < 20 {
		blockWidth = 20
	}

	// Build code block
	var block strings.Builder

	// Add top padding line (empty line with background)
	block.WriteString(ansiBG)
	block.WriteString(strings.Repeat(" ", blockWidth))
	block.WriteString(ansiReset)
	block.WriteString("\n")

	// Process each line
	lines := strings.Split(highlighted, "\n")
	for _, line := range lines {
		// Start with background
		block.WriteString(ansiBG)
		block.WriteString("  ") // Left padding

		// Add the highlighted line
		block.WriteString(line)

		// Pad the rest of the line with spaces to fill width
		// Note: We need to pad to visual width, not byte length
		// Since line may contain ANSI codes, we count visible columns
		lineLen := visualLen(line)
		paddingNeeded := blockWidth - 2 - lineLen // 2 for left padding
		if paddingNeeded > 0 {
			block.WriteString(strings.Repeat(" ", paddingNeeded))
		}

		// Right padding already included in the fill
		block.WriteString(ansiReset)
		block.WriteString("\n")
	}

	// Add bottom padding line (empty line with background)
	block.WriteString(ansiBG)
	block.WriteString(strings.Repeat(" ", blockWidth))
	block.WriteString(ansiReset)
	block.WriteString("\n")

	r.buffer.WriteString(block.String())
}

// visualLen calculates the visible length of a string (excluding ANSI codes)
func visualLen(s string) int {
	// Count visible characters, skipping ANSI escape sequences
	visLen := 0
	inEscape := false
	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' {
			inEscape = true
		} else if inEscape {
			if s[i] == 'm' {
				inEscape = false
			}
		} else {
			visLen++
		}
	}
	return visLen
}

// renderPlainCodeBlock renders a code block without syntax highlighting
func (r *RichMarkdownRenderer) renderPlainCodeBlock(code string) {
	// Use configurable background color, default to Aura theme
	bgColor := r.bgColor
	if bgColor == "" {
		bgColor = "#1a1a2e"
	}
	// Parse hex color to RGB
	var rVal, gVal, bVal int
	fmt.Sscanf(bgColor, "#%02x%02x%02x", &rVal, &gVal, &bVal)
	ansiBG := fmt.Sprintf("\x1b[48;2;%d;%d;%dm", rVal, gVal, bVal)
	// Slightly lighter foreground for code
	ansiFG := "\x1b[38;2;237;236;238m"
	ansiReset := "\x1b[0m"

	// Calculate padding width
	blockWidth := int(r.width) - 4
	if blockWidth < 20 {
		blockWidth = 20
	}

	var block strings.Builder

	// Top padding line
	block.WriteString(ansiBG)
	block.WriteString(strings.Repeat(" ", blockWidth))
	block.WriteString(ansiReset)
	block.WriteString("\n")

	// Code lines
	lines := strings.Split(code, "\n")
	for _, line := range lines {
		block.WriteString(ansiBG)
		block.WriteString(ansiFG)
		block.WriteString("  ")
		block.WriteString(line)

		// Pad to full width
		lineLen := visualLen(line)
		paddingNeeded := blockWidth - 2 - lineLen
		if paddingNeeded > 0 {
			block.WriteString(strings.Repeat(" ", paddingNeeded))
		}

		block.WriteString(ansiReset)
		block.WriteString("\n")
	}

	// Bottom padding line
	block.WriteString(ansiBG)
	block.WriteString(strings.Repeat(" ", blockWidth))
	block.WriteString(ansiReset)
	block.WriteString("\n")

	r.buffer.WriteString(block.String())
}

// renderCodeSpan renders inline code
func (r *RichMarkdownRenderer) renderCodeSpan(n *ast.CodeSpan, source []byte) {
	// Get text from the code span
	var code bytes.Buffer
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if text, ok := c.(*ast.Text); ok {
			code.Write(text.Segment.Value(source))
		}
	}

	// Use configurable background color for inline code
	bgColor := r.bgColor
	if bgColor == "" {
		bgColor = "#1a1a2e"
	}
	// Parse hex color to RGB
	var rVal, gVal, bVal int
	fmt.Sscanf(bgColor, "#%02x%02x%02x", &rVal, &gVal, &bVal)
	ansiBG := fmt.Sprintf("\x1b[48;2;%d;%d;%dm", rVal, gVal, bVal)
	ansiFG := "\x1b[38;2;248;248;242m" // Soft white
	ansiReset := "\x1b[0m"

	r.buffer.WriteString(" ") // Left padding
	r.buffer.WriteString(ansiBG)
	r.buffer.WriteString(ansiFG)
	r.buffer.WriteString(code.String())
	r.buffer.WriteString(ansiReset)
	r.buffer.WriteString(" ") // Right padding
}

// renderEmphasis renders italic or bold text based on level
// Level 1 = italic, Level 2 = bold
func (r *RichMarkdownRenderer) renderEmphasis(n *ast.Emphasis, source []byte) {
	text := r.extractText(n, source)

	var style lipgloss.Style
	if n.Level >= 2 {
		// Bold
		style = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F8F8F2"))
	} else {
		// Italic
		style = lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("#F8F8F2"))
	}

	r.buffer.WriteString(style.Render(text))
}

// renderList renders a list
func (r *RichMarkdownRenderer) renderList(n *ast.List, source []byte, depth int) {
	// Lists are handled by their children (ListItems)
	r.renderChildren(n, source, depth+1)
}

// renderListItem renders a list item
func (r *RichMarkdownRenderer) renderListItem(n *ast.ListItem, source []byte, depth int) {
	indent := strings.Repeat("  ", depth)

	// Check if parent is ordered list
	bullet := "•"
	if parent := n.Parent(); parent != nil {
		if list, ok := parent.(*ast.List); ok {
			if list.IsOrdered() {
				// Find the index of this item
				idx := 0
				for sibling := parent.FirstChild(); sibling != nil; sibling = sibling.NextSibling() {
					if sibling == n {
						break
					}
					idx++
				}
				bullet = fmt.Sprintf("%d.", list.Start+idx)
			}
		}
	}

	bulletStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#a277ff"))

	// Write bullet and first line
	r.buffer.WriteString(indent + bulletStyle.Render(bullet) + " ")

	// Track if we need a newline before nested content
	hasNestedList := false
	firstParagraph := true

	// Render the content
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		switch node := c.(type) {
		case *ast.Paragraph:
			if !firstParagraph {
				r.buffer.WriteString("\n" + indent + "  ")
			}
			r.renderInlineChildren(node, source)
			firstParagraph = false
		case *ast.List:
			// Nested list - ensure newline and proper depth
			hasNestedList = true
			r.buffer.WriteString("\n")
			r.renderList(node, source, depth+1)
		default:
			r.renderNode(c, source, depth)
		}
	}

	if !hasNestedList {
		r.buffer.WriteString("\n")
	}
}

// renderInlineChildren renders only inline children
func (r *RichMarkdownRenderer) renderInlineChildren(n ast.Node, source []byte) {
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		r.renderNode(c, source, 0)
	}
}

// renderBlockquote renders a blockquote
func (r *RichMarkdownRenderer) renderBlockquote(n *ast.Blockquote, source []byte, depth int) {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6c6c6c")).
		BorderLeft(true).
		BorderForeground(lipgloss.Color("#a277ff")).
		PaddingLeft(1)

	var content strings.Builder
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if p, ok := c.(*ast.Paragraph); ok {
			text := r.extractText(p, source)
			content.WriteString(text + "\n")
		}
	}

	r.buffer.WriteString(style.Render(content.String()) + "\n")
}

// renderLink renders a link
func (r *RichMarkdownRenderer) renderLink(n *ast.Link, source []byte) {
	text := r.extractText(n, source)
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7ee7ff")).
		Underline(true)

	r.buffer.WriteString(style.Render(text))
}

// renderAutoLink renders an auto link
func (r *RichMarkdownRenderer) renderAutoLink(n *ast.AutoLink, source []byte) {
	label := string(n.Label(source))
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7ee7ff")).
		Underline(true)

	r.buffer.WriteString(style.Render(label))
}

// renderImage renders an image (shows alt text)
func (r *RichMarkdownRenderer) renderImage(n *ast.Image, source []byte) {
	text := r.extractText(n, source)
	if text == "" {
		text = "[image]"
	}
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#ffca85")).
		Italic(true)

	r.buffer.WriteString(style.Render("[" + text + "]"))
}

// renderTable renders a markdown table using lipgloss table component
func (r *RichMarkdownRenderer) renderTable(n *extAST.Table, source []byte) {
	// Collect headers
	var headers []string
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if header, ok := c.(*extAST.TableHeader); ok {
			for cell := header.FirstChild(); cell != nil; cell = cell.NextSibling() {
				if tc, ok := cell.(*extAST.TableCell); ok {
					text := r.extractText(tc, source)
					headers = append(headers, text)
				}
			}
		}
	}

	// Collect rows
	var rows [][]string
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if row, ok := c.(*extAST.TableRow); ok {
			var rowData []string
			for cell := row.FirstChild(); cell != nil; cell = cell.NextSibling() {
				if tc, ok := cell.(*extAST.TableCell); ok {
					text := r.extractText(tc, source)
					rowData = append(rowData, text)
				}
			}
			if len(rowData) > 0 {
				rows = append(rows, rowData)
			}
		}
	}

	// Use lipgloss table component
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#6c6c6c"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == 0 {
				// Header row
				return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ff7edb"))
			}
			return lipgloss.NewStyle()
		}).
		Headers(headers...).
		Rows(rows...)

	r.buffer.WriteString("\n" + t.Render() + "\n\n")
}

// extractText extracts plain text from a node
func (r *RichMarkdownRenderer) extractText(n ast.Node, source []byte) string {
	var buf bytes.Buffer

	ast.Walk(n, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch node := n.(type) {
		case *ast.Text:
			buf.Write(node.Segment.Value(source))
		case *ast.String:
			buf.Write(node.Value)
		}
		return ast.WalkContinue, nil
	})

	return buf.String()
}

// RenderADF renders Atlassian Document Format as markdown
func (r *RichMarkdownRenderer) RenderADF(ADF interface{}) (string, error) {
	// Convert ADF to markdown first
	markdown := adfToMarkdownRich(ADF)
	return r.Render(markdown)
}

// adfToMarkdownRich converts ADF to markdown string
func adfToMarkdownRich(ADF interface{}) string {
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
					result.WriteString(renderADFNodeRich(nodeMap))
					result.WriteString("\n")
				}
			}
		}

		return result.String()
	}

	return ""
}

// renderADFNodeRich renders a single ADF node
func renderADFNodeRich(node map[string]interface{}) string {
	nodeType, _ := node["type"].(string)

	switch nodeType {
	case "paragraph":
		return renderADFTextRich(node)
	case "heading":
		attrs, _ := node["attrs"].(map[string]interface{})
		level, _ := attrs["level"].(float64)
		prefix := strings.Repeat("#", int(level))
		return prefix + " " + renderADFTextRich(node)
	case "bulletList":
		return renderADFListRich(node, "bullet")
	case "orderedList":
		return renderADFListRich(node, "ordered")
	case "codeBlock":
		content := renderADFTextRich(node)
		return "```\n" + content + "\n```"
	case "blockquote":
		content := renderADFTextRich(node)
		lines := strings.Split(content, "\n")
		var result strings.Builder
		for _, line := range lines {
			result.WriteString("> ")
			result.WriteString(line)
			result.WriteString("\n")
		}
		return result.String()
	default:
		return renderADFTextRich(node)
	}
}

// renderADFTextRich extracts text from an ADF node
func renderADFTextRich(node map[string]interface{}) string {
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
						text = applyMarksRich(text, marks)
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
					result.WriteString(renderADFNodeRich(childMap))
				}
			}
		}
	}

	return result.String()
}

// renderADFListRich renders list nodes
func renderADFListRich(node map[string]interface{}, listType string) string {
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
					result.WriteString(renderADFTextRich(childMap))
					result.WriteString("\n")
				}
			}
		}
	}

	return result.String()
}

// applyMarksRich applies formatting marks to text
func applyMarksRich(text string, marks []interface{}) string {
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
