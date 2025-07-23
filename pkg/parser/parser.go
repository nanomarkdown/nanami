/*
Copyright 2025 rivst.
This file is part of nanami.

nanami is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

nanami is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with nanami. If not, see <https://www.gnu.org/licenses/>.
*/

package parser

import (
	"fmt"
	"strings"

	"github.com/nanomarkdown/nanami/pkg/ast"
	stringUtil "github.com/nanomarkdown/nanami/pkg/common/strings"
)

var globalBib *ast.Webography

func ParseFile(lines []string) (*ast.Document, error) {
	globalBib = new(ast.Webography)
	globalBib.LoadFromFile("webography")

	doc := &ast.Document{
		Title:   "",
		Content: []ast.Node{},
		Cases:   []ast.CaseNode{},
		NoNLP:   false,
	}

	i := 0

	for i < len(lines) {
		line := strings.TrimSpace(lines[i])

		if strings.HasPrefix(line, "title:") {
			doc.Title = strings.TrimSpace(line[6:])
			i++
		} else if line == "!nlp" {
			doc.NoNLP = true
			i++
		} else if line == "content {" {
			i++
			break
		} else if line == "" {
			i++
		} else {
			break
		}
	}

	parseContentBody(lines, i, doc)

	return doc, nil
}

func parseContentBody(lines []string, start int, doc *ast.Document) int {
	i := start

	for i < len(lines) {
		line := strings.TrimSpace(lines[i])

		if line == "}" {
			return i + 1
		} else if strings.HasPrefix(line, "case(") {
			caseNode, newI := parseCase(lines, i, doc)
			doc.Cases = append(doc.Cases, *caseNode)
			i = newI
		} else if line == "text {" {
			textBlock, newI := parseTextBlock(lines, i, doc)
			doc.Content = append(doc.Content, textBlock)
			i = newI
		} else if line == "sources {" {
			sourcesBlock, newI := parseSourcesBlock(lines, i)
			doc.Content = append(doc.Content, sourcesBlock)
			i = newI
		} else {
			i++
		}
	}

	return i
}

func parseCase(lines []string, start int, doc *ast.Document) (*ast.CaseNode, int) {
	i := start
	line := strings.TrimSpace(lines[i])

	caseNode := &ast.CaseNode{
		Title:    "",
		Link:     "",
		Body:     []ast.Node{},
		SubCases: []ast.CaseNode{},
	}

	if strings.HasPrefix(line, "case(") {
		titleStart := 5
		titleEnd := strings.Index(line[titleStart:], ")")
		if titleEnd != -1 {
			caseNode.Title = line[titleStart : titleStart+titleEnd]

			remaining := line[titleStart+titleEnd+1:]
			if strings.HasPrefix(remaining, "(") && strings.Contains(remaining, ") {") {
				linkEnd := strings.Index(remaining[1:], ")")
				if linkEnd != -1 {
					caseNode.Link = remaining[1 : linkEnd+1]
				}
			}
		}
	}

	i++

	for i < len(lines) {
		line := strings.TrimSpace(lines[i])

		if line == "}" {
			return caseNode, i + 1
		} else if line == "text {" {
			textBlock, newI := parseTextBlock(lines, i, doc)
			caseNode.Body = append(caseNode.Body, textBlock)
			i = newI
		} else if line == "sources {" {
			sourcesBlock, newI := parseSourcesBlock(lines, i)
			caseNode.Body = append(caseNode.Body, sourcesBlock)
			i = newI
		} else if strings.HasPrefix(line, "case(") {
			subCase, newI := parseCase(lines, i, doc)
			caseNode.SubCases = append(caseNode.SubCases, *subCase)
			i = newI
		} else {
			i++
		}
	}

	return caseNode, i
}

func parseTextBlock(lines []string, start int, doc *ast.Document) (*ast.TextNode, int) {
	i := start + 1

	textBlock := &ast.TextNode{
		Content: "",
		NoNLP:   doc.NoNLP,
	}

	var contentLines []string

	for i < len(lines) {
		line := strings.TrimSpace(lines[i])

		if line == "}" {
			break
		} else if line != "" {
			contentLines = append(contentLines, line)
		}
		i++
	}

	textBlock.Content = strings.Join(contentLines, " ")

	return textBlock, i + 1
}

func parseSourcesBlock(lines []string, start int) (*ast.SourcesNode, int) {
	i := start + 1

	sourcesBlock := &ast.SourcesNode{
		Content: "",
	}

	var contentLines []string

	for i < len(lines) {
		line := strings.TrimSpace(lines[i])

		if line == "}" {
			break
		} else if line != "" {
			contentLines = append(contentLines, line)
		}
		i++
	}

	sourcesBlock.Content = strings.Join(contentLines, " ")

	return sourcesBlock, i + 1
}

func ParseInlineElements(content string) string {
	result := strings.Builder{}
	result.Grow(len(content) * 2)

	i := 0
	for i < len(content) {
		switch {
		case content[i] == '{':
			if newI, replacement, found := tryParseImage(content, i); found {
				result.WriteString(replacement)
				i = newI
			} else if newI, replacement, found := tryParseLink(content, i); found {
				result.WriteString(replacement)
				i = newI
			} else if newI, replacement, found := tryParseFootnotes(content, i); found {
				result.WriteString(replacement)
				i = newI
			} else {
				result.WriteByte(content[i])
				i++
			}
		case content[i] == '$' && i+1 < len(content) && content[i+1] == '{':
			if newI, replacement, found := tryParseReference(content, i); found {
				result.WriteString(replacement)
				i = newI
			} else {
				result.WriteByte(content[i])
				i++
			}
		default:
			result.WriteByte(content[i])
			i++
		}
	}

	return result.String()
}

func tryParseImage(content string, start int) (int, string, bool) {
	if start+5 >= len(content) || content[start:start+5] != "{img/" {
		return start, "", false
	}

	pathEnd := stringUtil.FindClosingBrace(content, start+1)
	if pathEnd == -1 {
		return start, "", false
	}

	if pathEnd+1 >= len(content) || content[pathEnd+1] != '{' {
		return start, "", false
	}

	altEnd := stringUtil.FindClosingBrace(content, pathEnd+2)
	if altEnd == -1 {
		return start, "", false
	}

	imagePath := content[start+5 : pathEnd]
	altText := content[pathEnd+2 : altEnd]

	replacement := fmt.Sprintf(`<img src="%s" alt="%s"/>`,
		imagePath,
		altText)

	return altEnd + 1, replacement, true
}

func tryParseLink(content string, start int) (int, string, bool) {
	if start+6 >= len(content) || !stringUtil.StartsWithHttp(content, start+1) {
		return start, "", false
	}

	urlEnd := stringUtil.FindClosingBrace(content, start+1)
	if urlEnd == -1 {
		return start, "", false
	}

	url := content[start+1 : urlEnd]
	linkText := url // Default link text is the URL
	endPos := urlEnd + 1

	if endPos < len(content) && content[endPos] == '{' {
		textEnd := stringUtil.FindClosingBrace(content, endPos+1)
		if textEnd != -1 {
			linkText = content[endPos+1 : textEnd]
			endPos = textEnd + 1
		}
	}

	replacement := fmt.Sprintf(`<a href="%s">%s</a>`,
		url,
		linkText)

	return endPos, replacement, true
}

func tryParseReference(content string, start int) (int, string, bool) {
	if start+2 >= len(content) || content[start:start+2] != "${" {
		return start, "", false
	}

	keywordEnd := stringUtil.FindClosingBrace(content, start+2)
	if keywordEnd == -1 {
		return start, "", false
	}

	keyword := content[start+2 : keywordEnd]

	var replacement string
	if globalBib != nil {
		replacement = globalBib.GetReference(keyword)
	}

	return keywordEnd + 1, replacement, true
}

func tryParseFootnotes(content string, start int) (int, string, bool) {
	footnotesStr := "{footnotes}"
	if start+len(footnotesStr) > len(content) || content[start:start+len(footnotesStr)] != footnotesStr {
		return start, "", false
	}

	var replacement string
	if globalBib != nil {
		replacement = globalBib.GenerateFootnotes()
	}

	return start + len(footnotesStr), replacement, true
}
