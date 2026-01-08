package ui

import (
	"fmt"
)

// ANSI color and style constants for CLI output
const (
	ColorReset = "\033[0m"
	ColorBold  = "\033[1m"
	ColorDim   = "\033[2m"

	ColorCyan   = "\033[36m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorWhite  = "\033[97m"
	ColorRed    = "\033[31m"
	ColorBlue   = "\033[34m"
)

// Icons - Removed as per request

// Convenience helper to build styled strings.
func Bold(s string) string {
	return ColorBold + s + ColorReset
}

func Dim(s string) string {
	return ColorDim + s + ColorReset
}

func Success(s string) string {
	return ColorGreen + s + ColorReset
}

func Info(s string) string {
	return ColorDim + s + ColorReset
}

func Warn(s string) string {
	return ColorYellow + s + ColorReset
}

func Error(s string) string {
	return ColorRed + s + ColorReset
}

func Hint(s string) string {
	return ColorYellow + s + ColorReset
}

// Styled components
func StyledError(msg string) string {
	return Error(msg)
}

func StyledSuccess(msg string) string {
	return Success(msg)
}

func StyledInfo(msg string) string {
	return Info(msg)
}

func StyledHint(msg string) string {
	return fmt.Sprintf("%s %s", Hint("Hint:"), Dim(msg))
}

func StyledWarn(msg string) string {
	return Warn(msg)
}

// Heading prints a stylized heading
func Heading(s string) string {
	return ColorCyan + s + ColorReset
}
