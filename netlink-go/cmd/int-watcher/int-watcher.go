package main

import (
	"fmt"
	"os"

	"github.com/anfredette/netlinktest/netlink-go/pkg/testfuncs"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide a namespace as an argument")
		os.Exit(1)
	}
	namespace := os.Args[1]
	testfuncs.IntWatcher(namespace)
}
