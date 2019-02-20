package main

import (
	"fmt"
	"log"
	"net"
	"net/mail"
	"net/smtp"
	"regexp"
	"strings"
	"time"
)

func firstMxFromDomain(domain string) (string, error) {
	mxs, err := net.LookupMX(domain)
	if err != nil {
		return "", err
	}

	// Return the first MX
	return mxs[0].Host, nil
}

func normalizeAddress(address string) (*mail.Address, error) {
	parser := new(mail.AddressParser)
	addr, err := parser.Parse(address)
	if err != nil {
		return nil, err
	}

	return addr, nil
}

func getDomainFromAddress(address string) (string, error) {
	addr, err := normalizeAddress(address)
	if err != nil {
		return "", err
	}
	at := strings.LastIndex(addr.Address, "@")
	if at < 0 {
		return "", fmt.Errorf("No @domain in address")
	}
	return address[at+1:], nil
}

// Decide whether a given address is deliverable at a given MX host, with optional timeout.
func isDeliverable(host string, address string, timeout ...time.Duration) (bool, error) {
	deliverable := false
	var conn net.Conn
	var err error

	trace.Printf("Connecting...")
	if len(timeout) > 0 {
		// We use a different Dialer if timeout is provided
		conn, err = net.DialTimeout("tcp", host, timeout[0])
	} else {
		conn, err = net.Dial("tcp", host)
	}

	if err != nil {
		if err, ok := err.(net.Error); ok && err.Timeout() {
			return false, errTimeout
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

	trace.Printf("Connected.")
	trace.Printf("MAIL FROM:<%s>", address)
	err = cli.Mail(address)
	if err != nil {
		log.Printf("Error on MAIL: %s\n", err)
		return false, err
	}

	trace.Printf("RCPT TO:<%s>", address)
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

	trace.Printf("RSET")
	err = cli.Reset()
	if err != nil {
		log.Printf("Error on RSET: %s\n", err)
		return deliverable, err
	}

	trace.Printf("QUIT")
	err = cli.Quit()
	if err != nil {
		log.Printf("Error on QUIT: %s\n", err)
		return deliverable, err
	}

	return deliverable, nil
}
