package testfuncs

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"

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
					go startIntWatcher(nsName)
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

func IntWatcher(namespace string) {

	var nsHandle netns.NsHandle
	var err error
	if namespace == "" {
		nsHandle = netns.None()
		namespace = "None"
	} else {
		nsHandle, err = netns.GetFromName(namespace)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Create a channel to receive updates
	links := make(chan netlink.LinkUpdate)

	// Subscribe to link updates
	done := make(chan struct{})
	if err := netlink.LinkSubscribeAt(nsHandle, links, done); err != nil {
		log.Fatal(err)
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
