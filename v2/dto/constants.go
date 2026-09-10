package dto

type RelayLevel string

const (
	Debug RelayLevel = "debug"
	Info  RelayLevel = "info"
	Warn  RelayLevel = "warn"
	Error RelayLevel = "error"
	Fatal RelayLevel = "fatal"
	Meta  RelayLevel = "meta"
)

var Levels = []RelayLevel{Fatal, Error, Warn, Info, Debug}
