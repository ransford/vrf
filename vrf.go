package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/smtp"
	"os"
	"regexp"
	"strings"
)

var Trace *log.Logger

func getDomainFromAddress(address string) (string, error) {
	at := strings.LastIndex(address, "@")
	if at < 0 {
		return "", fmt.Errorf("Cannot parse address")
	}
	return address[at+1:], nil
}

func isDeliverable(host string, address string) (bool, error) {
	deliverable := false

	Trace.Printf("Connecting...")
	cli, err := smtp.Dial(host)
	if err != nil {
		log.Printf("Error on connect: %s\n", err)
		return false, err
	}
	defer cli.Close()
	Trace.Printf("Connected.")

	Trace.Printf("MAIL FROM:<%s>", address)
	err = cli.Mail(address)
	if err != nil {
		log.Printf("Error on MAIL: %s\n", err)
		return false, err
	}

	Trace.Printf("RCPT TO:<%s>", address)
	err = cli.Rcpt(address)
	if err != nil {
		rx := regexp.MustCompile("^(451|550) [0-9]\\.1\\.1")

		// SMTP 550 X.1.1 means invalid address, but other errors mean other things
		if !rx.MatchString(err.Error()) {
			log.Printf("Error on RCPT: %s\n", err)
			return false, err
		}
		return false, nil
	}

	// If RCPT succeeded, the server thinks the address is deliverable
	deliverable = true

	Trace.Printf("RSET")
	err = cli.Reset()
	if err != nil {
		log.Printf("Error on RSET: %s\n", err)
		return deliverable, err
	}

	Trace.Printf("QUIT")
	err = cli.Quit()
	if err != nil {
		log.Printf("Error on QUIT: %s\n", err)
		return deliverable, err
	}

	return deliverable, nil
}

func firstMxFromDomain(domain string) (string, error) {
	mxs, err := net.LookupMX(domain)
	if err != nil {
		return "", err
	}

	// Return the first MX
	return mxs[0].Host, nil
}

func main() {
	// Parse command-line flags
	verbosePtr := flag.Bool("verbose", false, "Show verbose messages")
	quietPtr := flag.Bool("quiet", false, "Quiet (no output)")
	flag.Parse()

	if *verbosePtr && *quietPtr {
		log.Fatalf("Cannot be both quiet and verbose.")
	}

	// Set up verbose logging if required
	var traceDest io.Writer
	traceDest = ioutil.Discard
	if *verbosePtr {
		traceDest = os.Stderr
	}
	Trace = log.New(traceDest, "INFO: ", log.LstdFlags)

	log.SetOutput(os.Stderr)

	args := flag.Args()
	if len(args) != 1 {
		log.Fatalf("Usage: %s <address>\n", os.Args[0])
	}

	address := args[0]
	Trace.Printf("Address: %s\n", address)

	domain, err := getDomainFromAddress(address)
	if err != nil {
		log.Fatal(err)
	}
	Trace.Printf("Domain: %s\n", domain)

	mxHost, err := firstMxFromDomain(domain)
	if err != nil {
		log.Fatal(err)
	}
	Trace.Printf("MX host: %s\n", mxHost)

	host := fmt.Sprintf("%s:25", mxHost)

	deliverable, err := isDeliverable(host, address)
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
