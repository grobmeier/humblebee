package ui

import "github.com/fatih/color"

var (
	green  = color.New(color.FgGreen)
	yellow = color.New(color.FgYellow)
	red    = color.New(color.FgRed)
)

func DisableColor() {
	color.NoColor = true
}

func PrintSuccess(msg string) {
	green.Printf("✓ %s\n", msg)
}

func PrintWarning(msg string) {
	yellow.Printf("⚠  %s\n", msg)
}

func PrintError(msg string) {
	red.Printf("✗ %s\n", msg)
}

