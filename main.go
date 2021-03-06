/*
vrf tests whether a given email address is likely to be deliverable.

To test whether an address is deliverable, i.e., whether it's a "valid" email
address that can receive email, vrf finds an email server responsible for the
domain, then conncts to that server and follows *most* of the protocol to
deliver an email message, up to the point at which a message is actually
delivered.

 $ vrf good.guess@nowhere.biz
 good.guess@nowhere.biz is deliverable

 $ vrf bad.guess@nowhere.biz
 bad.guess@nowhere.biz is not deliverable

 # exits with 0 if deliverable, 1 otherwise.
 $ vrf -quiet good.guess@nowhere.biz && echo "good guess!"
 good guess!
*/
package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"
)

var errTimeout = errors.New("request timeout")
var trace *log.Logger

func setupLogging(verbose bool) {
	// Set up verbose logging if required
	var traceDest = ioutil.Discard

	if verbose {
		traceDest = os.Stderr
	}
	trace = log.New(traceDest, "INFO: ", log.LstdFlags)

	log.SetOutput(os.Stderr)
}

func main() {
	// Parse command-line flags
	verbosePtr := flag.Bool("verbose", false, "Show verbose messages")
	quietPtr := flag.Bool("quiet", false, "Quiet (no output; exit value reflects success)")
	timeoutPtr := flag.String("timeout", "", "Timeout after this duration (e.g., 3s)")
	flag.Parse()

	if *verbosePtr && *quietPtr {
		log.Fatalf("Cannot be both quiet and verbose.")
	}

	setupLogging(*verbosePtr)

	args := flag.Args()
	if len(args) != 1 {
		log.Fatalf("Usage: %s <address>\n", os.Args[0])
	}

	address := args[0]
	trace.Printf("Address: %s\n", address)

	domain, err := getDomainFromAddress(address)
	if err != nil {
		log.Fatalf("Invalid email address.")
	}
	trace.Printf("Domain: %s\n", domain)

	mxHost, err := firstMxFromDomain(domain)
	if err != nil {
		log.Fatal(err)
	}
	trace.Printf("MX host: %s\n", mxHost)

	host := fmt.Sprintf("%s:25", mxHost)
	var deliverable bool
	if *timeoutPtr != "" {
		timeout, err := time.ParseDuration(*timeoutPtr)
		if err != nil {
			log.Fatal("Invalid duration. Use something like 2s, 1m, etc.")
		}
		deliverable, err = isDeliverable(host, address, timeout)
	} else {
		deliverable, err = isDeliverable(host, address)
	}
	if err != nil {
		log.Fatal(err)
	}

	if deliverable {
		if !*quietPtr {
			fmt.Println(address, "is deliverable")
		}
		os.Exit(0)
	} else {
		if !*quietPtr {
			fmt.Println(address, "is not deliverable")
		}
		os.Exit(1)
	}
}
