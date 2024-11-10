package utilities

import (
	"fmt"
	"strings"
	"time"
)

// This is a fast approach for concatenating strings in Go
// strings.Join is optimized for this purpose and generally faster than manual concatenation
func ConcatWords(joiner []byte, words ...[]byte) string {
    // Calculate the total length of the resulting byte slice
    totalLen := len(joiner) * (len(words) - 1) // Length contributed by joiners
    for _, word := range words {
        totalLen += len(word)
    }
    
    if totalLen <= 0 {
        return ""
    }

    // Create a byte slice with the required length
    b := make([]byte, totalLen)

    // Populate the byte slice
    currentIndex := 0
    for i, word := range words {
        copy(b[currentIndex:], word)
        currentIndex += len(word)
        if i < len(words) - 1 {
            copy(b[currentIndex:], joiner)
            currentIndex += len(joiner)
        }
    }

    return string(b)
}

// GetCurrentDateTime returns the current date and time details in a different format.
// It returns six values: the current date in "YYYY-MM-DD" format, the current year, the current month,
// the current day, the current day of the week, and the current time in "03:04 PM MST" format.
// Returns:
// - string: Current date in "YYYY-MM-DD" format.
// - int: Current year.
// - int: Current month as an integer.
// - int: Current day of the month.
// - string: Current day of the week.
// - string: Current time in "03:04 PM MST" format.
func GetCurrentDateTimeWithTimeZoneShift(userTimeZone string) (string, int, int, int, string,  string) {
	now := time.Now().UTC()
	location, err := time.LoadLocation(strings.Split(userTimeZone, ",")[0])
	if err != nil {
		fmt.Println("Error loading location, defaulting to UTC:", err)
		return now.Format("2006-01-02"), now.Year(), int(now.Month()), now.Day(), now.Weekday().String(), now.Format("03:04 PM MST")
	}

	// Convert the UTC time to the specified userTimeZone
	localTime := now.In(location)

	// Format the time components according to the local time
	date := localTime.Format("2006-01-02")
	year := localTime.Year()
	month := int(localTime.Month())
	day := localTime.Day()
	weekday := localTime.Weekday().String()
	formattedTime := localTime.Format("03:04 PM MST")

	return date, year, month, day, weekday, formattedTime
}