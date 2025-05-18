package main

import (
	"runtime"
	"runtime/debug"
)

func CollectAndFree() {
	// Free the memory
	runtime.GC()
	// Free the memory|release the memory
	debug.FreeOSMemory()
}
