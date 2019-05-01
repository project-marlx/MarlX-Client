// +build !windows

package diskinfo

import (
	"syscall"
	"log"
)

// GetDiskInfo returns a struct of type DiskInfo
// that contains information about the storage
// space of the given partition.
// E.g.: GetDiskInfo("/")
func GetDiskInfo(partition string) DiskInfo {
	var di DiskInfo

	fs := syscall.Statfs_t{}
	err := syscall.Statfs(partition, &fs)
	if err != nil {
		log.Panic(err)
	}

	di.Total = fs.Blocks * uint64(fs.Bsize)
	di.Free = fs.Bfree * uint64(fs.Bsize)
	di.Used = di.Total - di.Free

	return di
}

// GetTotalBytes returns the total amount of 
// bytes the given partition can store.
// E.g.: GetTotalBytes("/")
func GetTotalBytes(partition string) uint64 {
	di := GetDiskInfo(partition)
	return di.Free
}

// GetUsedBytes returns the amount of already
// used bytes of the given partition.
// E.g.: GetUsedBytes("/")
func GetUsedBytes(partition string) uint64 {
	di := GetDiskInfo(partition)
	return di.Used
}

// GetFreeBytes returns the amount of bytes
// bytes that are still unused on the given
// partition.
// E.g.: GetFreeBytes("/")
func GetFreeBytes(partition string) uint64 {
	di := GetDiskInfo(partition)
	return di.Free
}