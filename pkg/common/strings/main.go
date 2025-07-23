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

package strings

func StartsWithHttp(content string, start int) bool {
	if start+7 <= len(content) && content[start:start+7] == "http://" {
		return true
	}
	if start+8 <= len(content) && content[start:start+8] == "https://" {
		return true
	}
	return false
}

func FindClosingBrace(content string, start int) int {
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
