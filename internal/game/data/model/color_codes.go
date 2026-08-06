package models

// ColorCode represents the color codes for item color changes.
type ColorCode string

const (
	ColorCodeNoColorChange ColorCode = ""     // No color change
	ColorCodeWhite         ColorCode = "whit" // White
	ColorCodeLightGrey     ColorCode = "lgry" // Light Grey
	ColorCodeDarkGrey      ColorCode = "dgry" // Dark Grey
	ColorCodeBlack         ColorCode = "blac" // Black
	ColorCodeLightBlue     ColorCode = "lblu" // Light Blue
	ColorCodeDarkBlue      ColorCode = "dblu" // Dark Blue
	ColorCodeCrystalBlue   ColorCode = "cblu" // Crystal Blue
	ColorCodeLightRed      ColorCode = "lred" // Light Red
	ColorCodeDarkRed       ColorCode = "dred" // Dark Red
	ColorCodeCrystalRed    ColorCode = "cred" // Crystal Red
	ColorCodeLightGreen    ColorCode = "lgrn" // Light Green
	ColorCodeDarkGreen     ColorCode = "dgrn" // Dark Green
	ColorCodeCrystalGreen  ColorCode = "cgrn" // Crystal Green
	ColorCodeLightYellow   ColorCode = "lyel" // Light Yellow
	ColorCodeDarkYellow    ColorCode = "dyel" // Dark Yellow
	ColorCodeLightGold     ColorCode = "lgld" // Light Gold
	ColorCodeDarkGold      ColorCode = "dgld" // Dark Gold
	ColorCodeLightPurple   ColorCode = "lpur" // Light Purple
	ColorCodeDarkPurple    ColorCode = "dpur" // Dark Purple
	ColorCodeOrange        ColorCode = "oran" // Orange
	ColorCodeBrightWhite   ColorCode = "bwht" // Bright White
)
