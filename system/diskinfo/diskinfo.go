// Package diskinfo provides info about
// some disk's storage space.
package diskinfo

import "fmt"

// DiskInfo holds the amount of total,
// used, and free Bytes of some disk/partition.
type DiskInfo struct {
	Total	uint64	`json:"total"`
	Used	uint64	`json:"used"`
	Free 	uint64	`json:"free"`
}

// Returns the string representation of a 
// DiskInfo-Struct. 
// DiskInfo{Total: [int64], Used: [int64], Free: [int64]}
func (di *DiskInfo) String() string {
	return fmt.Sprintf("DiskInfo{Total: %dB, Used: %dB, Free: %dB}", di.Total, di.Used, di.Free)
}