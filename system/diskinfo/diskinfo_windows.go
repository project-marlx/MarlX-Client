// +build windows

package diskinfo

import (
	"syscall"
	"log"
	"unsafe"
)

// GetDiskInfo returns a struct of type DiskInfo
// that contains information about the storage
// space of the given partition.
// E.g.: GetDiskInfo("C:")
func GetDiskInfo(partition string) DiskInfo {
	var di DiskInfo

	kernel32, err := syscall.LoadLibrary("Kernel32.dll")
	if err != nil {
		log.Panic(err)
	}

	defer syscall.FreeLibrary(kernel32)
	GetDiskFreeSpaceEx, err := syscall.GetProcAddress(syscall.Handle(kernel32), "GetDiskFreeSpaceExW")

	if err != nil {
		log.Panic(err)
	}

	var lpFreeBytesAvailable, lpTotalNumberOfBytes, lpTotalNumberOfFreeBytes int64
	_, _, _ = syscall.Syscall6(uintptr(GetDiskFreeSpaceEx), 4, 
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(partition))),
		uintptr(unsafe.Pointer(&lpFreeBytesAvailable)),
		uintptr(unsafe.Pointer(&lpTotalNumberOfBytes)),
		uintptr(unsafe.Pointer(&lpTotalNumberOfFreeBytes)), 0, 0)

	di.Total = uint64(lpTotalNumberOfBytes)
	di.Free = uint64(lpTotalNumberOfFreeBytes)
	di.Used = di.Total - di.Free

	return di
}

// GetTotalBytes returns the total amount of 
// bytes the given partition can store.
// E.g.: GetTotalBytes("C:")
func GetTotalBytes(partition string) uint64 {
	di := GetDiskInfo(partition)
	return di.Total
}

// GetUsedBytes returns the amount of already
// used bytes of the given partition.
// E.g.: GetUsedBytes("C:")
func GetUsedBytes(partition string) uint64 {
	di := GetDiskInfo(partition)
	return di.Used
}

// GetFreeBytes returns the amount of bytes
// bytes that are still unused on the given
// partition.
// E.g.: GetFreeBytes("C:")
func GetFreeBytes(partition string) uint64 {
	di := GetDiskInfo(partition)
	return di.Free
}