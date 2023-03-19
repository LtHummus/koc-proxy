package vbackend

import "fmt"

const (
	GameVersion          = 264847
	Platform             = "win64"
	GameBuild            = "final"
	BootSessionID        = "8c5d1500-42d0458f-60029d8c-01d9548d" // TODO: this is static for now, change it? (Also this is a weird way of writing a guid?)
	PID           uint64 = 2875485078631744371                   // no idea what this is
)

var (
	GameVersionString = fmt.Sprintf("%d", GameVersion)
)
