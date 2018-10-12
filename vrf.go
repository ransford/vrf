package main

import (
	"errors"
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
	"time"
)

var (
	ErrTimeout = errors.New("Request timeout")
)
var Trace *log.Logger

func getDomainFromAddress(address string) (string, error) {
	at := strings.LastIndex(address, "@")
	if at < 0 {
		return "", fmt.Errorf("Cannot parse address")
	}
	return address[at+1:], nil
}

func isDeliverable(host string, address string, timeout ...time.Duration) (bool, error) {
	deliverable := false
	var conn net.Conn
	var err error

	Trace.Printf("Connecting...")
	if len(timeout) > 0 {
		// We use a different Dialer if timeout is provided
		conn, err = net.DialTimeout("tcp", host, timeout[0])
	} else {
		conn, err = net.Dial("tcp", host)
	}

	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			return false, ErrTimeout
		}
		return false, err
	}

	// We need the address without the port to create
	// the instance of the client.
	hostNoPort, _, _ := net.SplitHostPort(address)
	cli, err := smtp.NewClient(conn, hostNoPort)
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
	timeoutPtr := flag.String("timeout", "", "Timeout after this duration (e.g. 3s)")
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
	var deliverable bool
	if *timeoutPtr != "" {
		timeout, err := time.ParseDuration(*timeoutPtr)
		if err != nil {
			log.Fatal("Invalid duration. Use something like 2s, 1m etc.")
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
