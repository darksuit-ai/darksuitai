package utilities

import (
	"fmt"
	"log"
	"strings"
)

// Printer formats and prints a message with specified color and tag.
// It now supports an expanded set of colors based on the ANSI 256-color palette.
// Args:
// - message string: The message to be color formatted
// - tag string: The color code tag
// - color string: The color to be used (now supports extended color codes)
// Returns:
// - Prints the formatted message to log output
func Printer(tag string, message string, color string) {
	colorCodes := map[string]string{
		// Basic Colors (16-color mode)
		"black":       "\033[0;30m",
		"red":         "\033[0;31m",
		"green":       "\033[0;32m",
		"yellow":      "\033[0;33m",
		"blue":        "\033[0;34m",
		"magenta":     "\033[0;35m",
		"cyan":        "\033[0;36m",
		"white":       "\033[0;37m",

		// Bright/Bold variants
		"bold_black":  "\033[1;30m",
		"bold_red":    "\033[1;31m",
		"bold_green":  "\033[1;32m",
		"bold_yellow": "\033[1;33m",
		"bold_blue":   "\033[1;34m",
		"bold_purple": "\033[1;35m",
		"bold_cyan":   "\033[1;36m",
		"bold_white":  "\033[1;37m",

		// Extended colors from the 256-color palette (from the image)
		"light_blue":   "\033[38;5;75m",  // #5fd7ff
		"deep_blue":    "\033[38;5;27m",  // #005fff
		"royal_blue":   "\033[38;5;63m",  // #5f5fff
		"teal":         "\033[38;5;31m",  // #0087af
		"light_green":  "\033[38;5;118m", // #87ff00
		"forest_green": "\033[38;5;28m",  // #008700
		"lime":         "\033[38;5;82m",  // #5fff00
		"orange":       "\033[38;5;214m", // #ffaf00
		"gold":         "\033[38;5;220m", // #ffd700
		"pink":         "\033[38;5;205m", // #ff5faf
		"violet":       "\033[38;5;93m",  // #8700ff
		"purple":       "\033[38;5;129m", // #af00ff
		
		// Additional variants from the image
		"light_cyan":    "\033[38;5;159m", // #afffff
		"deep_magenta":  "\033[38;5;165m", // #d700ff
		"pale_yellow":   "\033[38;5;228m", // #ffff87
		"dark_red":      "\033[38;5;124m", // #af0000
		"bright_red":    "\033[38;5;196m", // #ff0000
		"pastel_green":  "\033[38;5;157m", // #afffaf
		"pastel_blue":   "\033[38;5;153m", // #afd7ff
		"pastel_purple": "\033[38;5;147m", // #afafff

		// Grayscale tones from the image
		"gray_light":  "\033[38;5;252m", // #d0d0d0
		"gray_medium": "\033[38;5;244m", // #808080
		"gray_dark":   "\033[38;5;238m", // #444444

		// Special formatting
		"reset":       "\033[0m",
		"bold":        "\033[1m",
		"dim":         "\033[2m",
		"italic":      "\033[3m",
		"underline":   "\033[4m",
	}

	// // Create a gradient function for smooth color transitions
	// gradientColors := func(startColor, endColor, steps int) []string {
	// 	var colors []string
	// 	for i := 0; i < steps; i++ {
	// 		colorCode := startColor + (endColor-startColor)*i/(steps-1)
	// 		colors = append(colors, fmt.Sprintf("\033[38;5;%dm", colorCode))
	// 	}
	// 	return colors
	// }

	colorCode := colorCodes[strings.ToLower(color)]
	if colorCode == "" {
		colorCode = colorCodes["white"] // Default to white if color not found
	}

	// Format and print the message
	message = fmt.Sprintf("%s%s", tag, message)
	coloredMessage := fmt.Sprintf("%s%s%s\n", colorCode, message, colorCodes["reset"])
	log.Println(coloredMessage)
}

// Helper function to print rainbow text
func PrintRainbow(message string) {
	rainbowColors := []string{
		"\033[38;5;196m", // red
		"\033[38;5;214m", // orange
		"\033[38;5;226m", // yellow
		"\033[38;5;46m",  // green
		"\033[38;5;21m",  // blue
		"\033[38;5;93m",  // violet
	}

	var coloredMessage string
	for i, char := range message {
		colorIndex := i % len(rainbowColors)
		coloredMessage += fmt.Sprintf("%s%c", rainbowColors[colorIndex], char)
	}
	log.Println(coloredMessage + "\033[0m")
}