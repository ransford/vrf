package main

import (
	"fmt"
	"log"
	"net"
	"net/smtp"
	"os"
	"strings"
)

func get_domain_from_address(address string) (domain string, err error) {
	at := strings.LastIndex(address, "@")
	if at < 0 {
		return "", fmt.Errorf("Invalid domain")
	} else {
		return address[at+1:], nil
	}
}

func is_deliverable(host string, address string) (bool, error) {
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

func get_mx_from_domain(domain string) (mxhost string, err error) {
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

	domain, err := get_domain_from_address(address)
	if err != nil {
		log.Fatal("Error: cannot get domain from address.")
	}
	log.Printf("Domain: %s\n", domain)

	mx_host, err := get_mx_from_domain(domain)
	if err != nil {
		log.Fatal("Error: cannot get domain from address.")
	}
	log.Printf("MX host: %s\n", mx_host)

	host := fmt.Sprintf("%s:25", mx_host)

	deliverable, err := is_deliverable(host, address)
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
