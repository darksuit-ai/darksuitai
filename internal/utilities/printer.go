package utilities

import (
	"fmt"
	"log"
	"strings"
)

// This function maps ANSI codes to color names to change text color.
// It returns the passed message in the colour name and returns a white color as a default color if the color name passed wasn't found in color dictionary.
// Args:
// - message string: The message to be be color formatted
// - tag string: The color code tag
// - color string: The color to be used
// Returns:
// - string responses of top border, rows, and headers
func Printer(tag string, message string, color string) {
	colorCodes := map[string]string{
		"orange":      "33;1", // orange
		"sky_blue":    "\033[0;36m", // Cyan is often substituted for sky blue
		"green":       "\x1b[32m",   // Correct
		"magenta":     "\x1b[35m",   // Corrected from \x1b[36m (cyan) to \x1b[35m (magenta)
		"red":         "\033[0;31m", // Correct
		"cyan":        "\033[0;36m", // 
		"violet":      "\033[38;5;93m",   // Violet
		"pink":        "\033[38;5;205m",
		"yellow":      "\033[0;33m", // Correct
		"blue":        "\033[0;34m", // Correct
		"purple":      "\033[0;35m", // Correct
		"white":       "\033[0;37m", // Correct
		"gold":        "\033[1;33m", // Bright yellow can be used as a substitute for gold
		"bold_black":  "\033[1;30m", // Correct
		"bold_red":    "\033[1;31m", // Correct
		"bold_green":  "\033[1;32m", // Correct
		"bold_yellow": "\033[1;33m", // Correct
		"bold_blue":   "\033[1;34m", // Correct
		"bold_purple": "\033[1;35m", // Correct
		"bold_cyan":   "\033[1;36m", // Correct
		"bold_white":  "\033[1;37m", // Correct
		"reset":       "\033[0m",    // Correct
	}

	colorCode := colorCodes[strings.ToLower(color)]
	if colorCode == "" {
		colorCode = colorCodes["white"]
	}
	message= fmt.Sprintf("%s%s", tag, message)
	coloredMessage := fmt.Sprintf("%s%s%s\n", colorCode, message, colorCodes["reset"])
	log.Println(coloredMessage)
}
