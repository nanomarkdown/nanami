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
	"testing"
)

func TestParseTextBlock(t *testing.T) {
	text := "I exist!"
	textBlock := fmt.Sprintf(`text {
		%s
	}`, text)
	lines := strings.Split(textBlock, "\n")

	res, count := parseTextBlock(lines, 0, false)

	if res.Content != text {
		t.Errorf("Expected text '%s', got '%s'", text, res.Content)
	}

	if count != len(lines) {
		t.Errorf("Expected count of %d, got %d", len(lines), count)
	}
}
