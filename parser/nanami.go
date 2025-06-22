package parser

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

type Node interface {
	Render(w io.Writer, indent int)
}

func writeIndent(w io.Writer, level int, s string) {
	indent := strings.Repeat("  ", level)
	fmt.Fprint(w, indent, s, "\n")
}

type WBibEntry struct {
	Keyword string
	URL     string
	Name    string
	Date    string
}

type Webography struct {
	entries map[string]*WBibEntry
	ordered []*WBibEntry
	counter int
}

var globalBib *Webography

func NewWebography() *Webography {
	return &Webography{
		entries: make(map[string]*WBibEntry),
		ordered: []*WBibEntry{},
		counter: 0,
	}
}

func (b *Webography) LoadFromFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err // File doesn't exist, continue without webography
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentEntry *WBibEntry

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			if currentEntry != nil && currentEntry.Keyword != "" {
				b.entries[currentEntry.Keyword] = currentEntry
				currentEntry = nil
			}
			continue
		}

		if strings.HasPrefix(line, "T: ") {
			// In case there was no linebreak between entries
			if currentEntry != nil && currentEntry.Keyword != "" {
				b.entries[currentEntry.Keyword] = currentEntry
			}
			currentEntry = &WBibEntry{Keyword: strings.TrimSpace(line[3:])}
		} else if strings.HasPrefix(line, "L: ") && currentEntry != nil {
			currentEntry.URL = strings.TrimSpace(line[3:])
		} else if strings.HasPrefix(line, "N: ") && currentEntry != nil {
			currentEntry.Name = strings.TrimSpace(line[3:])
		} else if strings.HasPrefix(line, "D: ") && currentEntry != nil {
			currentEntry.Date = strings.TrimSpace(line[3:])
		}
	}

	// Process remaining entry
	if currentEntry != nil && currentEntry.Keyword != "" {
		b.entries[currentEntry.Keyword] = currentEntry
	}

	return scanner.Err()
}

func (b *Webography) GetReference(keyword string) string {
	if entry, exists := b.entries[keyword]; exists {
		for i, orderedEntry := range b.ordered {
			if orderedEntry.Keyword == keyword {
				return fmt.Sprintf(`<sup><a href="#s%[1]d">[%[1]d]</a></sup>`, i+1)
			}
		}

		b.ordered = append(b.ordered, entry)
		b.counter++
		return fmt.Sprintf(`<sup><a href="#s%[1]d">[%[1]d]</a></sup>`, b.counter)
	}
	return ""
}

func (b *Webography) GenerateFootnotes() string {
	if len(b.ordered) == 0 {
		return ""
	}

	var result strings.Builder
	result.WriteString("<ol>\n")

	for i, entry := range b.ordered {
		result.WriteString(fmt.Sprintf(`            <li id="s%[1]d">%[1]d. %s, %s`,
			i+1, entry.Name, entry.Date))
		if entry.URL != "" {
			result.WriteString(fmt.Sprintf(` <a href="%[1]s">%[1]s</a>`,
				entry.URL))
		}
		result.WriteString("</li>\n")
	}

	result.WriteString("        </ol>")
	return result.String()
}

type Document struct {
	Title   string
	Content []Node
	Cases   []CaseNode
	NoNLP   bool
}

func (d *Document) Render(w io.Writer, indent int) {
	writeIndent(w, indent, "<html>")
	writeIndent(w, indent+1, "<head>")
	writeIndent(w, indent+2, "<title>"+d.Title+"</title>")
	writeIndent(w, indent+1, "</head>")
	writeIndent(w, indent+1, "<body>")
	for _, n := range d.Content {
		n.Render(w, indent+2)
	}
	for _, c := range d.Cases {
		c.Render(w, indent+2)
	}
	writeIndent(w, indent+1, "</body>")
	writeIndent(w, indent, "</html>")
}

type InlineProcessor struct{}

func (ip *InlineProcessor) ProcessInlineElements(content string) string {
	result := strings.Builder{}
	result.Grow(len(content) * 2)

	i := 0
	for i < len(content) {
		switch {
		case content[i] == '{':
			if newI, replacement, found := ip.tryParseImage(content, i); found {
				result.WriteString(replacement)
				i = newI
			} else if newI, replacement, found := ip.tryParseLink(content, i); found {
				result.WriteString(replacement)
				i = newI
			} else if newI, replacement, found := ip.tryParseFootnotes(content, i); found {
				result.WriteString(replacement)
				i = newI
			} else {
				result.WriteByte(content[i])
				i++
			}
		case content[i] == '$' && i+1 < len(content) && content[i+1] == '{':
			if newI, replacement, found := ip.tryParseReference(content, i); found {
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

func (ip *InlineProcessor) tryParseImage(content string, start int) (int, string, bool) {
	if start+5 >= len(content) || content[start:start+5] != "{img/" {
		return start, "", false
	}

	pathEnd := ip.findClosingBrace(content, start+1)
	if pathEnd == -1 {
		return start, "", false
	}

	if pathEnd+1 >= len(content) || content[pathEnd+1] != '{' {
		return start, "", false
	}

	altEnd := ip.findClosingBrace(content, pathEnd+2)
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

func (ip *InlineProcessor) tryParseLink(content string, start int) (int, string, bool) {
	if start+6 >= len(content) || !ip.startsWithHttp(content, start+1) {
		return start, "", false
	}

	urlEnd := ip.findClosingBrace(content, start+1)
	if urlEnd == -1 {
		return start, "", false
	}

	url := content[start+1 : urlEnd]
	linkText := url // Default link text is the URL
	endPos := urlEnd + 1

	if endPos < len(content) && content[endPos] == '{' {
		textEnd := ip.findClosingBrace(content, endPos+1)
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

func (ip *InlineProcessor) tryParseReference(content string, start int) (int, string, bool) {
	if start+2 >= len(content) || content[start:start+2] != "${" {
		return start, "", false
	}

	keywordEnd := ip.findClosingBrace(content, start+2)
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

func (ip *InlineProcessor) tryParseFootnotes(content string, start int) (int, string, bool) {
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

func (ip *InlineProcessor) findClosingBrace(content string, start int) int {
	braceCount := 1
	for i := start; i < len(content); i++ {
		switch content[i] {
		case '{':
			braceCount++
		case '}':
			braceCount--
			if braceCount == 0 {
				return i
			}
		}
	}
	return -1
}

func (ip *InlineProcessor) startsWithHttp(content string, start int) bool {
	if start+7 <= len(content) && content[start:start+7] == "http://" {
		return true
	}
	if start+8 <= len(content) && content[start:start+8] == "https://" {
		return true
	}
	return false
}

type TextBlockNode struct {
	Content string
	NoNLP   bool
}

func (tb *TextBlockNode) Render(w io.Writer, indent int) {
	writeIndent(w, indent, `<div class="text-block">`)

	processor := &InlineProcessor{}
	content := processor.ProcessInlineElements(tb.Content)
	content = strings.TrimSpace(content)

	if tb.NoNLP {
		if content != "" {
			writeIndent(w, indent+1, content)
		}
	} else {
		if content != "" {
			writeIndent(w, indent+1, "<p>"+content+"</p>")
		}
	}

	writeIndent(w, indent, "</div>")
}

type SourcesNode struct {
	Content string
}

func (s *SourcesNode) Render(w io.Writer, indent int) {
	writeIndent(w, indent, `<div class="sources">`)

	processor := &InlineProcessor{}
	content := processor.ProcessInlineElements(s.Content)
	content = strings.TrimSpace(content)

	if content != "" {
		writeIndent(w, indent+1, content)
	}
	writeIndent(w, indent, "</div>")
}

type CaseNode struct {
	Title    string
	Link     string
	Body     []Node
	SubCases []CaseNode
}

func (c *CaseNode) Render(w io.Writer, indent int) {
	writeIndent(w, indent, `<div class="case">`)

	id := strings.ReplaceAll(strings.TrimSpace(c.Title), " ", "_")

	if c.Link != "" {
		titleTag := fmt.Sprintf(`<h4 id="%s"><a href="%s">%s</a></h4>`,
			id,
			c.Link,
			c.Title)
		writeIndent(w, indent+1, titleTag)
	} else {
		titleTag := fmt.Sprintf(`<h4 id="%s">%s</h4>`,
			id,
			c.Title)
		writeIndent(w, indent+1, titleTag)
	}

	for _, n := range c.Body {
		n.Render(w, indent+1)
	}

	for _, subCase := range c.SubCases {
		subCase.Render(w, indent+1)
	}

	writeIndent(w, indent, "</div>")
}

func ParseFile(lines []string) (*Document, error) {
	globalBib = NewWebography()
	globalBib.LoadFromFile("webography")

	doc := &Document{
		Title:   "",
		Content: []Node{},
		Cases:   []CaseNode{},
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

func parseContentBody(lines []string, start int, doc *Document) int {
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

func parseCase(lines []string, start int, doc *Document) (*CaseNode, int) {
	i := start
	line := strings.TrimSpace(lines[i])

	caseNode := &CaseNode{
		Title:    "",
		Link:     "",
		Body:     []Node{},
		SubCases: []CaseNode{},
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

func parseTextBlock(lines []string, start int, doc *Document) (*TextBlockNode, int) {
	i := start + 1

	textBlock := &TextBlockNode{
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

func parseSourcesBlock(lines []string, start int) (*SourcesNode, int) {
	i := start + 1

	sourcesBlock := &SourcesNode{
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
