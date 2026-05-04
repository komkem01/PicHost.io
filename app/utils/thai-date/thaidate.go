// Package thaidate provides utilities for working with Thai calendar dates,
// including conversion from Unix timestamps and time.Time to Thai Buddhist era dates.
package thaidate

import (
	"fmt"
	"time"
)

// GetThaiDateString converts a Unix timestamp to a Thai date format string.
// Returns an empty string if timestamp is 0.
func GetThaiDateString(timestamp int64) string {
	if timestamp == 0 {
		return ""
	}
	// format 2006-01-02 to 02 กุมพาพันธ์ 2549
	months := []string{
		"มกราคม", "กุมภาพันธ์", "มีนาคม", "เมษายน", "พฤษภาคม", "มิถุนายน",
		"กรกฎาคม", "สิงหาคม", "กันยายน", "ตุลาคม", "พฤศจิกายน", "ธันวาคม",
	}
	t := time.Unix(timestamp, 0)
	year := t.Year() + 543 // Convert to Thai Buddhist year
	return fmt.Sprintf("%02d %s %d", t.Day(), months[t.Month()-1], year)
}

// GetThaiDateFromTime converts a time.Time value to a Thai date format string.
func GetThaiDateFromTime(t time.Time) string {
	return GetThaiDateString(t.Unix())
}

// GetCurrentThaiDateString returns the current date in Thai date format.
func GetCurrentThaiDateString() string {
	return GetThaiDateFromTime(time.Now())
}
