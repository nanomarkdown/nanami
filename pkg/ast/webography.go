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

package ast

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Webography struct {
	entries map[string]*WBibEntry
	ordered []*WBibEntry
	counter int
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

// TODO: All rendering must be decoupled from ASTs
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
