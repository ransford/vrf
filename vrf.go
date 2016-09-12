package main

import (
	"fmt"
	"net/smtp"
	"os"
	"strings"
)

func usage() {
	fmt.Printf("Usage: %s <host> <address>\n", os.Args[0])
}

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
		return false, err
	}
	defer cli.Close()

	err = cli.Mail("foo@mysite.com")
	if err != nil {
		return false, err
	}

	err = cli.Rcpt(address)
	if err != nil {
		return false, err
	}

	err = cli.Reset()
	if err != nil {
		return deliverable, err
	}

	err = cli.Quit()
	if err != nil {
		return deliverable, err
	}

	return deliverable, nil
}

func main() {
	args := os.Args[1:]
	if len(args) != 2 {
		usage()
		os.Exit(1)
	}

	host := fmt.Sprintf("%s:25", args[0])
	address := args[1]

	deliverable, err := is_deliverable(host, address)
	if err != nil {
		fmt.Println("Error:", err)
	}

	if deliverable {
		fmt.Println(address, "is deliverable")
		os.Exit(0)
	} else {
		fmt.Println(address, "is not deliverable")
		os.Exit(1)
	}
}
