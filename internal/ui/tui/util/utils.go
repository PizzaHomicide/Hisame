package util

import "github.com/mattn/go-runewidth"

// TruncateString cuts a string to fit within maxWidth visual width
func TruncateString(s string, maxWidth int) string {
	width := 0
	for i, r := range s {
		charWidth := runewidth.RuneWidth(r)
		// Check if adding this rune would exceed maxWidth
		if width+charWidth > maxWidth-3 { // Reserve space for "..."
			return s[:i] + "..."
		}
		width += charWidth
	}
	return s // Return as is if it fits
}
