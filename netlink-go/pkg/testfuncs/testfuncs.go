package testfuncs

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

func NsWatcher() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)

	if len(os.Args) >= 2 {
		namespace := os.Args[1]
		fmt.Printf("Starting IntWatcher on namespace: %s\n", namespace)
		go IntWatcher(namespace)
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				fmt.Printf("event: %+v\n", event)
				if event.Op&fsnotify.Create == fsnotify.Create {
					nsName := strings.Split(event.Name, "/")[len(strings.Split(event.Name, "/"))-1]
					fmt.Printf("New network namespace detected: %s\n", nsName)
					fmt.Printf("Starting IntWatcher on namespace: %s\n", nsName)
					// go startIntWatcher(nsName)
					// time.Sleep(30 * time.Second)
					go IntWatcher(nsName)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Printf("error: %+v\n", err)
			}
		}
	}()

	err = watcher.Add("/var/run/netns")
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

func startIntWatcher(namespace string) {
	command := "/home/afredette/go/src/github.com/netlink-test/netlink-go/cmd/int-watcher/int-watcher"
	cmd := exec.Command(command, namespace) // replace "ls -l" with the command you want to run

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
	}
}

func ListLinksInNamespace(nsName string) {
	// Save the current network namespace
	origNS, err := netns.Get()
	if err != nil {
		fmt.Printf("Failed to get current network namespace: %v\n", err)
		return
	}
	defer origNS.Close()

	// Open the desired network namespace
	targetNS, err := netns.GetFromPath("/var/run/netns/" + nsName)
	if err != nil {
		fmt.Printf("Failed to open target network namespace: %v\n", err)
		return
	}
	defer targetNS.Close()

	// Set the network namespace to the target
	if err := netns.Set(targetNS); err != nil {
		fmt.Printf("Failed to set network namespace: %v\n", err)
		return
	}

	// Ensure we switch back to the original namespace
	defer func() {
		if err := netns.Set(origNS); err != nil {
			fmt.Printf("Failed to reset network namespace: %v\n", err)
		}
		origNS.Close()
	}()

	// List the links in the target namespace
	fmt.Printf("Existing links in netns %s\n", nsName)
	links, err := netlink.LinkList()
	if err != nil {
		fmt.Printf("Failed to list links: %v\n", err)
		return
	}

	for _, link := range links {
		fmt.Printf("Link: %s\n", link.Attrs().Name)
	}
}

func IntWatcher(namespace string) {

	var nsHandle netns.NsHandle
	var err error
	// Create a channel to receive updates
	links := make(chan netlink.LinkUpdate)

	// Subscribe to link updates
	done := make(chan struct{})

	if namespace == "" {
		nsHandle = netns.None()
		namespace = "None"
		if err := netlink.LinkSubscribeAt(nsHandle, links, done); err != nil {
			fmt.Printf("LinkSubscribeAt failed for namespace %s, %v.  Waiting 1 second.\n", namespace, err)
			return
		}
	} else {
		subscribed := false
		for i := 0; i < 30; i++ {
			nsHandle, err = netns.GetFromName(namespace)
			if err != nil {
				fmt.Printf("Failed to get nsHandle from namespace %s, %v\n", namespace, err)
				return
			}

			if err := netlink.LinkSubscribeAt(nsHandle, links, done); err != nil {
				fmt.Printf("LinkSubscribeAt failed for namespace %s, %v.  Waiting 50 us...\n", namespace, err)
				time.Sleep(50 * time.Microsecond)
			} else {
				subscribed = true
				fmt.Printf("LinkSubscribeAt succeeded for namespace %s\n", namespace)
				break
			}
		}
		if !subscribed {
			fmt.Printf("Failed to subscribe to link updates for namespace %s. Exiting.\n", namespace)
			return
		}

		// List existing links
		ListLinksInNamespace(namespace)

	}

	// Process updates
	fmt.Printf("Processing link updates for namespace %s\n", namespace)
	for link := range links {
		attrs := link.Attrs()
		if attrs == nil {
			fmt.Printf("received link update without attributes. Ignoring.\nlink: %+v\n", link)
			continue
		}

		if link.Flags&(syscall.IFF_UP|syscall.IFF_RUNNING) != 0 && attrs.OperState == netlink.OperUp {
			fmt.Printf("Interface %s in namespace %s is up and running. OperState: %v, Flags: %v\n",
				attrs.Name, namespace, attrs.OperState, attrs.Flags)
		} else {
			fmt.Printf("Interface %s in Namespace %s is down. OperState: %v, Flags: %v\n",
				attrs.Name, namespace, attrs.OperState, attrs.Flags)
		}
	}
}
