package main

import (
	"errors"
	"log"
	"math/rand"
	"net"
	"net/mail"
	"net/smtp"
	"regexp"
	"strings"
	"time"
)

// EmailAddress is an email address to test for deliverability.
type EmailAddress struct {
	Address string
	domain  string
}

// NewEmailAddress returns a validated email address
func NewEmailAddress(address string) (*EmailAddress, error) {
	parser := &mail.AddressParser{}
	addr, err := parser.Parse(address)
	if err != nil {
		return nil, err
	}

	parts := strings.SplitN(addr.Address, "@", 2)
	if len(parts) != 2 {
		return nil, errors.New("Not enough parts in address")
	}
	return &EmailAddress{
		Address: addr.Address,
		domain:  parts[1],
	}, nil
}

// RandomMX returns a random mail exchanger for this address's domain name.
func (e *EmailAddress) RandomMX() (string, error) {
	mxs, err := net.LookupMX(e.domain)
	if err != nil {
		return "", err
	}
	if len(mxs) == 0 {
		return "", errors.New("no MX for domain")
	}

	return mxs[rand.Intn(len(mxs))].Host, nil
}

// FirstMX returns the first (highest-priority) mail exchanger for this
// address's domain name.
func (e *EmailAddress) FirstMX() (string, error) {
	mxs, err := net.LookupMX(e.domain)
	if err != nil {
		return "", err
	}
	if len(mxs) == 0 {
		return "", errors.New("no MX for domain")
	}

	// Return the first MX
	return mxs[0].Host, nil
}

// Domain extracts the domain part of an EmailAddress.
func (e *EmailAddress) Domain() string {
	return e.domain
}

// IsDeliverable returns true if a given address is deliverable at a given MX
// host, with optional timeout; false otherwise.
func (e *EmailAddress) IsDeliverable(host string, timeout ...time.Duration) (bool, error) {
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
	hostNoPort, _, _ := net.SplitHostPort(e.Address)
	cli, err := smtp.NewClient(conn, hostNoPort)
	if err != nil {
		log.Printf("Error on connect: %s\n", err)
		return false, err
	}
	defer cli.Close()

	trace.Printf("Connected.")
	trace.Printf("MAIL FROM:<%s>", e.Address)
	err = cli.Mail(e.Address)
	if err != nil {
		log.Printf("Error on MAIL: %s\n", err)
		return false, err
	}

	trace.Printf("RCPT TO:<%s>", e.Address)
	err = cli.Rcpt(e.Address)
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

func init() {
	rand.Seed(time.Now().Unix())
}
