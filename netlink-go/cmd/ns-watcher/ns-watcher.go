package main

import (
	"runtime"

	"github.com/anfredette/netlinktest/netlink-go/pkg/testfuncs"
)

func main() {
	// Ensure the main goroutine runs on its own thread
	runtime.LockOSThread()

	// Watch for new interfaces
	go testfuncs.IntWatcher("") // Use the fully qualified name of the function

	// Watch for new network namespaces
	go testfuncs.NsWatcher()

	// Block forever
	select {}
}
