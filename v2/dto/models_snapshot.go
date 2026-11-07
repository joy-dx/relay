package dto

type Snapshot[S any] struct {
	State    S      `json:"state"`
	Revision uint64 `json:"revision,string"`
}

type ReplayResult struct {
	Events  []EmittedEvent `json:"events"`
	Through uint64         `json:"through,string"`
	Gap     bool           `json:"gap"`
}
