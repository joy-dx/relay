package sinks

import "github.com/fatih/color"

var (
	successColor = color.New(color.FgBlack).Add(color.BgHiGreen)
	errorColor   = color.New(color.FgBlack).Add(color.BgRed)
)
