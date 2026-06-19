package api

import (
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// MarkdownToADF converts markdown text to Atlassian Document Format JSON
func MarkdownToADF(markdown string) map[string]interface{} {
	source := []byte(markdown)
	md := goldmark.New()
	doc := md.Parser().Parse(text.NewReader(source))

	content := walkNodes(doc, source)
	return map[string]interface{}{
		"type":    "doc",
		"version": 1,
		"content": content,
	}
}

func walkNodes(n ast.Node, source []byte) []map[string]interface{} {
	var result []map[string]interface{}
	for n != nil {
		switch n.Kind() {
		case ast.KindDocument, ast.KindTextBlock:
			result = append(result, walkChildren(n, source)...)
		case ast.KindParagraph:
			result = append(result, walkParagraph(n, source))
		case ast.KindHeading:
			result = append(result, walkHeading(n, source))
		case ast.KindList:
			result = append(result, walkList(n, source)...)
		case ast.KindFencedCodeBlock:
			result = append(result, walkCodeBlock(n, source))
		case ast.KindCodeBlock:
			result = append(result, walkCodeBlock(n, source))
		case ast.KindBlockquote:
			result = append(result, walkBlockquote(n, source))
		case ast.KindThematicBreak:
			result = append(result, map[string]interface{}{"type": "rule"})
		}
		n = n.NextSibling()
	}
	return result
}

func walkChildren(n ast.Node, source []byte) []map[string]interface{} {
	var result []map[string]interface{}
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		switch c.Kind() {
		case ast.KindParagraph:
			result = append(result, walkParagraph(c, source))
		case ast.KindHeading:
			result = append(result, walkHeading(c, source))
		case ast.KindList:
			result = append(result, walkList(c, source)...)
		case ast.KindFencedCodeBlock:
			result = append(result, walkCodeBlock(c, source))
		case ast.KindCodeBlock:
			result = append(result, walkCodeBlock(c, source))
		case ast.KindBlockquote:
			result = append(result, walkBlockquote(c, source))
		case ast.KindThematicBreak:
			result = append(result, map[string]interface{}{"type": "rule"})
		case ast.KindTextBlock:
			result = append(result, walkChildren(c, source)...)
		}
	}
	return result
}

func walkParagraph(n ast.Node, source []byte) map[string]interface{} {
	return map[string]interface{}{
		"type":    "paragraph",
		"content": walkInlineContent(n, source),
	}
}

func walkHeading(n ast.Node, source []byte) map[string]interface{} {
	level := 1
	if h, ok := n.(*ast.Heading); ok {
		level = h.Level
	}
	return map[string]interface{}{
		"type": "heading",
		"attrs": map[string]interface{}{
			"level": level,
		},
		"content": walkInlineContent(n, source),
	}
}

func walkList(n ast.Node, source []byte) []map[string]interface{} {
	list, ok := n.(*ast.List)
	if !ok {
		return nil
	}

	listType := "bulletList"
	if list.Start > 0 {
		listType = "orderedList"
	}

	var items []map[string]interface{}
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if c.Kind() == ast.KindListItem {
			items = append(items, walkListItem(c, source))
		}
	}
	return []map[string]interface{}{
		{
			"type":    listType,
			"content": items,
		},
	}
}

func walkListItem(n ast.Node, source []byte) map[string]interface{} {
	content := walkChildren(n, source)
	if content == nil {
		content = []map[string]interface{}{}
	}
	return map[string]interface{}{
		"type":    "listItem",
		"content": content,
	}
}

func walkCodeBlock(n ast.Node, source []byte) map[string]interface{} {
	lines := n.Lines()
	var textBuilder strings.Builder
	if lines != nil {
		for i := 0; i < lines.Len(); i++ {
			seg := lines.At(i)
			textBuilder.Write(seg.Value(source))
		}
	}

	result := map[string]interface{}{
		"type": "codeBlock",
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": textBuilder.String(),
			},
		},
	}

	if fcb, ok := n.(*ast.FencedCodeBlock); ok && fcb.Info != nil {
		infoSeg := fcb.Info.Segment
		lang := infoSeg.Value(source)
		if len(lang) > 0 {
			result["attrs"] = map[string]interface{}{
				"language": strings.TrimSpace(string(lang)),
			}
		}
	}
	return result
}

func walkBlockquote(n ast.Node, source []byte) map[string]interface{} {
	return map[string]interface{}{
		"type":    "blockquote",
		"content": walkChildren(n, source),
	}
}

func walkInlineContent(n ast.Node, source []byte) []map[string]interface{} {
	var content []map[string]interface{}
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		switch c.Kind() {
		case ast.KindText:
			textStr := string(c.(*ast.Text).Segment.Value(source))
			content = append(content, buildTextNode(textStr, collectMarks(c)))
		case ast.KindString:
			textStr := string(c.(*ast.String).Value)
			content = append(content, buildTextNode(textStr, collectMarks(c)))
		case ast.KindCodeSpan:
			textStr := string(c.Text(source))
			content = append(content, buildTextNode(textStr, []map[string]interface{}{{"type": "code"}}))
		case ast.KindEmphasis:
			content = append(content, walkInlineContent(c, source)...)
		case ast.KindLink:
			link := c.(*ast.Link)
			linkContent := walkInlineContent(c, source)
			linkMarks := []map[string]interface{}{
				{
					"type": "link",
					"attrs": map[string]interface{}{
						"href": string(link.Destination),
					},
				},
			}
			for i := range linkContent {
				if text, ok := linkContent[i]["text"]; ok {
					linkContent[i] = buildTextNode(text.(string), linkMarks)
				}
			}
			content = append(content, linkContent...)
		case ast.KindAutoLink:
			link := c.(*ast.AutoLink)
			url := string(link.URL(source))
			content = append(content, buildTextNode(url, []map[string]interface{}{
				{
					"type": "link",
					"attrs": map[string]interface{}{
						"href": url,
					},
				},
			}))
		case ast.KindImage:
			// Images are not well-supported in ADF; skip
		}
	}
	return content
}

func collectMarks(n ast.Node) []map[string]interface{} {
	var marks []map[string]interface{}
	for parent := n.Parent(); parent != nil; parent = parent.Parent() {
		switch parent.Kind() {
		case ast.KindEmphasis:
			emp := parent.(*ast.Emphasis)
			if emp.Level == 2 {
				marks = append(marks, map[string]interface{}{"type": "strong"})
			} else if emp.Level == 1 {
				marks = append(marks, map[string]interface{}{"type": "em"})
			}
		default:
			return marks
		}
	}
	return marks
}

func buildTextNode(text string, marks []map[string]interface{}) map[string]interface{} {
	node := map[string]interface{}{
		"type": "text",
		"text": text,
	}
	if len(marks) > 0 {
		// Deduplicate marks (same type can appear from nested emphasis)
		seen := make(map[string]bool)
		var unique []map[string]interface{}
		for _, m := range marks {
			t := m["type"].(string)
			if !seen[t] {
				seen[t] = true
				unique = append(unique, m)
			}
		}
		node["marks"] = unique
	}
	return node
}
