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

func TestAddressParse(t *testing.T) {
	/* No errors on good addresses */
	addrs := []struct {
		addr   string
		domain string
	}{
		{"foo@google.com", "google.com"},
		{"p@grrransford.org", "grrransford.org"},
		{"@foot.com", "foot.com"},
		{"@bar", "bar"},
		{"foo@", ""},
		{"bl@h@blah@blah.com", "blah.com"},
	}
	for _, tcase := range addrs {
		_, err := getDomainFromAddress(tcase.addr)
		if err != nil {
			t.Fatal("Error should be nil, but is", err)
		}
	}

	/* Errors on bad addresses */
	badAddrs := []string{
		"foo",
		"foo.com",
		"",
	}
	for _, badaddr := range badAddrs {
		_, err := getDomainFromAddress(badaddr)
		if err == nil {
			t.Fatal("err is nil; shoul be non-nil")
		}
	}
}

func TestDialTimeout(t *testing.T) {
	host := "spacedecode.com"
	addr := "foo@" + host

	passed := false

	// Timeouts: I'm not sue if this is too much time?
	funcTimeout := time.Second

	// If the timeout fails, we impose a slighly higher timeout
	// to this test so it doesn't block forever.
	testTimeout := time.Second * 2

	// Setup the exit channel so we can block
	exit := make(chan struct{})

	// We run this in a goroutine so we can wait on the channel.
	go func() {
		go func() {
			ticker := time.NewTicker(testTimeout)
			select {
			case <-ticker.C:
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
