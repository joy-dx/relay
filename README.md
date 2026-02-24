# relay

`relay` is a small Go logging facade designed for “broadcast-style” logging with rich, typed event data.

Instead of logging plain strings directly to one logger, you emit **structured events** to a central `RelaySvc`. The relay then **fans out** those events to any number of subscribed **sinks** (stdout loggers, structured loggers, filtered loggers, file loggers, etc).

This pattern makes it easy to:
- attach multiple outputs (terminal + structured logs + CI formatting) without duplicating log calls
- log **rich data** (typed structs) rather than ad-hoc key/value maps
- keep application code independent from specific logging implementations

---

## Concepts

### Relay service
`RelaySvc` is the broadcaster. You call:

- `relay.Info(event)`
- `relay.Warn(event)`
- `relay.Error(event)`
- `relay.Debug(event)`
- `relay.Fatal(event)` (emits to all sinks, then terminates the process)
- `relay.Meta(event)` (special “developer-defined” meta events for CLI UX, sections, status, etc)

### Events
An event is any value implementing:

```go
type RelayEventInterface interface {
	RelayChannel() EventChannel
	RelayType() EventRef
	Message() string
	ToSlog() []slog.Attr
}
```

Key points:
- `RelayType()` (`dto.EventRef`) is the event “key” used for routing/filtering.
- `Message()` is the human-friendly string.
- `ToSlog()` returns `[]slog.Attr` so structured sinks can log rich context.

Two built-in events exist in the repo:
- `events.RlyLog` (basic log message)
- `events.RlyMeta` (special meta events such as `section`, `success`, `failure`)

### Sinks
A sink is any implementation of:

```go
type RelaySinkInterface interface {
	Ref() string
	Debug(data RelayEventInterface)
	Info(data RelayEventInterface)
	Warn(data RelayEventInterface)
	Error(data RelayEventInterface)
	Fatal(data RelayEventInterface)
	Meta(data RelayEventInterface)
}
```

Sinks decide how and where the event is output (stdout, structured slog, filtered stdout, etc).

---

## Installation

```bash
go get github.com/joy-dx/relay
```

---

## Quick start

### 1) Create the relay service

```go
package main

import (
	"github.com/joy-dx/relay"
	"github.com/joy-dx/relay/config"
)

func main() {
	cfg := config.DefaultRelaySvcConfig()
	r := relay.ProvideRelaySvc(&cfg)

	// Register sinks here (next section)
	_ = r
}
```

> `ProvideRelaySvc` returns a singleton instance. The first call wins for configuration.

### 2) Register sinks

```go
package main

import (
	"github.com/joy-dx/relay"
	"github.com/joy-dx/relay/config"
	"github.com/joy-dx/relay/sinks"
)

func main() {
	cfg := config.DefaultRelaySvcConfig()
	r := relay.ProvideRelaySvc(&cfg)

	// Simple terminal logging
	simpleCfg := sinks.DefaultSimpleLoggerConfig().
		WithLevel("info").
		WithKeyPadding(14)
	r.RegisterSink(sinks.NewSimpleLogger(&simpleCfg))

	// Emit events
	// ...
}
```

### 3) Emit events

Using `events.RlyLog`:

```go
package main

import (
	"github.com/joy-dx/relay"
	"github.com/joy-dx/relay/events"
)

func main() {
	r := relay.ProvideRelaySvc(nil) // example only; in real code pass config

	r.Info(events.RlyLog{Msg: "service started"})
	r.Warn(events.RlyLog{Msg: "cache miss"})
}
```

Using `events.RlyMeta` (useful for CLI UX):

```go
r.Meta(events.RlyMeta{
	MetaType: "section",
	Text:     "Build",
})

r.Meta(events.RlyMeta{
	MetaType: "success",
	Text:     "Build completed",
})

r.Meta(events.RlyMeta{
	MetaType: "failure",
	Text:     "Build failed",
})
```

---

## Provided sinks

This repo includes several sink implementations in `github.com/joy-dx/relay/sinks`.

### 1) SimpleLoggerSink (`ref: "simple"`)

**Best for:** human-friendly terminal output with optional padding and basic level gating.

**Behavior:**
- `Debug/Info/Warn` print: `<event-type>: <message>`
- `Error/Fatal` print just the message
- `Meta` prints special formatted output (section headers, success/failure markers)

**Example:**

```go
simpleCfg := sinks.DefaultSimpleLoggerConfig().
	WithLevel(dto.Info).
	WithKeyPadding(16)

r.RegisterSink(sinks.NewSimpleLogger(&simpleCfg))

r.Info(events.RlyLog{Msg: "hello"})
```

### 2) StructuredLogger (`ref: "structured"`)

**Best for:** structured logs via Go’s `log/slog` text handler.

**Behavior:**
- Emits logs using `slog.Logger.LogAttrs(...)`
- Uses `convertLevel` to map `dto.RelayLevel` to `slog.Level`
- `Fatal` logs at ERROR level with the message `"FATAL"` (and includes attributes from `ToSlog()`)

**Example:**

```go
structuredCfg := sinks.DefaultStructuredLoggerConfig().
	WithLevel(dto.Info)

r.RegisterSink(sinks.NewStructuredLogger(&structuredCfg))

r.Info(events.RlyLog{Msg: "structured hello"})
```

To take advantage of structured attributes, use an event that implements `ToSlog()` meaningfully:

```go
type UserLogin struct {
	UserID string
}

func (e UserLogin) RelayChannel() dto.EventChannel { return "auth" }
func (e UserLogin) RelayType() dto.EventRef       { return "auth.login" }
func (e UserLogin) Message() string               { return "user login" }
func (e UserLogin) ToSlog() []slog.Attr {
	return []slog.Attr{
		slog.String("user_id", e.UserID),
	}
}

r.Info(UserLogin{UserID: "123"})
```

### 3) FilteredLoggerSink (`ref: "filtered"`)

**Best for:** keeping terminal noise down by only printing events whose `RelayType()` is explicitly allowed.

**Behavior:**
- `Debug/Info/Warn` print only if BOTH:
    1. level gating allows it, AND
    2. the event’s `RelayType()` exists in `cfg.RelayTypes`
- `Error/Fatal` always print the message (not filtered)
- `Meta` formats similarly to SimpleLoggerSink

**Example:**

```go
cfg := &sinks.FilteredLoggerConfig{
	Level:      dto.Info,
	RelayTypes: []dto.EventRef{"cmd.log"},
}
r.RegisterSink(sinks.NewFilteredLogger(cfg))

// Will print (type allowed)
r.Info(events.RlyMeta{MetaType: "section", Text: "CLI Output"})
r.Info(events.RlyLog{Msg: "this depends on RelayType used by the event"})

// Will be suppressed at Info/Warn/Debug if type not in RelayTypes
```

### 3) FileLoggerSink (`ref: "file"`)

**Best for**: persistent log output to disk (CLI tools, background jobs, desktop apps).

**Behavior:**
- `Debug/Info/Warn` print: <event-type>: <message>
- `Error` prints: ERROR: <message>
- `Fatal` prints: FATAL: <message>
- `Meta` prints: META: <message>

Automatically creates parent directories if they do not exist

Appends to the log file (does not overwrite)

Thread-safe writes

**Example:**

```go
appData := os.Getenv("APPDATA")
logPath := filepath.Join(appData, "Joydx", "joydx.log")

fileCfg := sinks.DefaultFileLoggerConfig().
WithFilePath(logPath).
WithLevel(dto.Info).
WithKeyPadding(16)

fileSink, err := sinks.NewFileLogger(&fileCfg)
if err != nil {
	log.Fatal(err)
}

r.RegisterSink(fileSink)

r.Info(events.RlyLog{Msg: "written to disk"})
```

---

## Logging levels

Relay levels are defined in `dto`:

```go
const (
	Debug dto.RelayLevel = "debug"
	Info  dto.RelayLevel = "info"
	Warn  dto.RelayLevel = "warn"
	Error dto.RelayLevel = "error"
	Fatal dto.RelayLevel = "fatal"
	Meta  dto.RelayLevel = "meta"
)
```

Level ordering is based on:

```go
var Levels = []RelayLevel{Fatal, Error, Warn, Info, Debug}
```

Sinks use `GetLogLevelIndex(cfg.Level, dto.Levels)` and compare indices to decide whether to output.

---

## Fatal behavior

`RelaySvc.Fatal(...)`:
1. dispatches the event to all sinks (`sink.Fatal(event)`)
2. calls `os.Exit(1)`

Because it exits the process, use `Fatal` sparingly (usually only at the application boundary / CLI entrypoint).

If you want `Fatal` to be testable without exiting, inject an exiter into the service (see refactor suggestion in the test discussions).

---

## Creating a custom event

Any struct can be an event by implementing `dto.RelayEventInterface`.

Example:

```go
package myevents

import "log/slog"
import "github.com/joy-dx/relay/dto"

type DBQuery struct {
	Query string
	Rows  int
}

func (e DBQuery) RelayChannel() dto.EventChannel { return "db" }
func (e DBQuery) RelayType() dto.EventRef       { return "db.query" }
func (e DBQuery) Message() string               { return "database query executed" }

func (e DBQuery) ToSlog() []slog.Attr {
	return []slog.Attr{
		slog.String("query", e.Query),
		slog.Int("rows", e.Rows),
	}
}
```

Emit it:

```go
r.Info(myevents.DBQuery{
	Query: "SELECT * FROM users",
	Rows:  42,
})
```

---

## Creating a custom sink

Implement `dto.RelaySinkInterface`. You can route based on:
- `e.RelayChannel()` (broad category)
- `e.RelayType()` (specific event key)
- `e.Message()` (human string)
- `e.ToSlog()` (structured attrs)

A minimal custom sink that counts events by type:

```go
package mysinks

import (
	"sync"

	"github.com/joy-dx/relay/dto"
)

type CountingSink struct {
	mu     sync.Mutex
	counts map[dto.EventRef]int
}

func NewCountingSink() *CountingSink {
	return &CountingSink{counts: map[dto.EventRef]int{}}
}

func (s *CountingSink) Ref() string { return "counting" }

func (s *CountingSink) add(e dto.RelayEventInterface) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counts[e.RelayType()]++
}

func (s *CountingSink) Debug(e dto.RelayEventInterface) { s.add(e) }
func (s *CountingSink) Info(e dto.RelayEventInterface)  { s.add(e) }
func (s *CountingSink) Warn(e dto.RelayEventInterface)  { s.add(e) }
func (s *CountingSink) Error(e dto.RelayEventInterface) { s.add(e) }
func (s *CountingSink) Fatal(e dto.RelayEventInterface) { s.add(e) }
func (s *CountingSink) Meta(e dto.RelayEventInterface)  { s.add(e) }

// Optional: expose counts
func (s *CountingSink) Snapshot() map[dto.EventRef]int {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(map[dto.EventRef]int, len(s.counts))
	for k, v := range s.counts {
		out[k] = v
	}
	return out
}
```

Register it:

```go
counter := mysinks.NewCountingSink()
r.RegisterSink(counter)
```

---

## Running tests

If you have tests that capture `os.Stdout`, ensure they do not run in parallel (or serialize stdout capture with a mutex), since `os.Stdout` is a global process-wide variable.

```bash
go test ./...
```

For race detection:

```bash
go test -race ./...
```
