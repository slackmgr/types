package common

import "encoding/json"

type Issue interface {
	json.Marshaler
	DedupID() string
	IsOpen() bool
	CurrentPostID() string
}

type MoveMapping interface {
	json.Marshaler
	DedupID() string
}
