package main

import (
	"fmt"
	"log"
	"net"
	"net/smtp"
	"os"
	"strings"
)

func getDomainFromAddress(address string) (domain string, err error) {
	at := strings.LastIndex(address, "@")
	if at < 0 {
		return "", fmt.Errorf("Invalid domain")
	}
	return address[at+1:], nil
}

func isDeliverable(host string, address string) (bool, error) {
	deliverable := false

	cli, err := smtp.Dial(host)
	if err != nil {
		log.Printf("Error on connect: %s\n", err)
		return false, err
	}
	defer cli.Close()

	err = cli.Mail(address)
	if err != nil {
		log.Printf("Error on MAIL: %s\n", err)
		return false, err
	}

	err = cli.Rcpt(address)
	if err != nil {
		log.Printf("Error on RCPT: %s\n", err)
		return false, err
	}
	deliverable = true

	err = cli.Reset()
	if err != nil {
		log.Printf("Error on RSET: %s\n", err)
		return deliverable, err
	}

	err = cli.Quit()
	if err != nil {
		log.Printf("Error on QUIT: %s\n", err)
		return deliverable, err
	}

	return deliverable, nil
}

func getMxFromDomain(domain string) (mxhost string, err error) {
	mxs, err := net.LookupMX(domain)
	if len(mxs) == 0 {
		return "", fmt.Errorf("No MX for domain")
	}
	return mxs[0].Host, nil
}

func main() {
	log.SetOutput(os.Stderr)

	args := os.Args[1:]
	if len(args) != 1 {
		log.Fatalf("Usage: %s <address>\n", os.Args[0])
	}

	address := args[0]
	log.Printf("Address: %s\n", address)

	domain, err := getDomainFromAddress(address)
	if err != nil {
		log.Fatal("Error: cannot get domain from address.")
	}
	log.Printf("Domain: %s\n", domain)

	mxHost, err := getMxFromDomain(domain)
	if err != nil {
		log.Fatal("Error: cannot get domain from address.")
	}
	log.Printf("MX host: %s\n", mxHost)

	host := fmt.Sprintf("%s:25", mxHost)

	deliverable, err := isDeliverable(host, address)
	if err != nil {
		log.Fatalf("Error checking deliverability: %s\n", err)
	}

	if deliverable {
		fmt.Println(address, "is deliverable")
		os.Exit(0)
	} else {
		fmt.Println(address, "is not deliverable")
		os.Exit(1)
	}
}
