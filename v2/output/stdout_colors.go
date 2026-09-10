package output

import "github.com/fatih/color"

var (
	SuccessColor = color.New(color.FgBlack).Add(color.BgHiGreen)
	ErrorColor   = color.New(color.FgBlack).Add(color.BgRed)
)
