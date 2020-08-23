package main

import (
	"log"
	"os"
	"testing"
	"time"
)

func init() {
	trace = log.New(os.Stdout, "INFO: ", log.LstdFlags)
}

func TestDialTimeout(t *testing.T) {
	host := "spacedecode.com"
	addr := "foo@" + host

	passed := false

	// Timeouts: I'm not sue if this is too much time?
	funcTimeout := time.Second

	// If the timeout fails, we impose a slightly higher timeout
	// to this test so it doesn't block forever.
	testTimeout := time.Second * 2

	// Setup the exit channel so we can block
	exit := make(chan struct{})

	// We run this in a goroutine so we can wait on the channel.
	go func() {
		go func() {
			ticker := time.NewTicker(testTimeout)
			for range ticker.C {
				exit <- struct{}{}
			}
		}()

		_, err := isDeliverable(host+":25", addr, funcTimeout)
		if err == errTimeout {
			exit <- struct{}{}
			passed = true
			return
		}
	}()
	<-exit

	if !passed {
		t.Fail()
	}
}
