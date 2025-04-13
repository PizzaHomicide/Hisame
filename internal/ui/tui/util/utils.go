package util

import (
	"fmt"
	"github.com/mattn/go-runewidth"
	"time"
)

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

// FormatTimeUntilAiring formats a duration into a human-readable string
// showing two levels of time (days/hours or hours/minutes) at most
func FormatTimeUntilAiring(seconds int64) string {
	timeUntil := time.Duration(seconds) * time.Second

	// Calculate days, hours, minutes
	days := int(timeUntil.Hours() / 24)
	hours := int(timeUntil.Hours()) % 24
	minutes := int(timeUntil.Minutes()) % 60

	// Format with consistent spacing:
	return fmt.Sprintf("%3dd %02dh %02dm", days, hours, minutes)
}
