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
	"io"
	"strings"
)

func writeIndent(w io.Writer, level int, s string) {
	indent := strings.Repeat("  ", level)
	fmt.Fprint(w, indent, s, "\n")
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
