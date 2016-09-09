package main

import (
	"fmt"
	"net/smtp"
	"os"
)

func usage() {
	fmt.Printf("Usage: %s <host> <address>\n", os.Args[0])
}

func main() {
	args := os.Args[1:]
	if len(args) != 2 {
		usage()
		os.Exit(1)
	}

	host := fmt.Sprintf("%s:25", args[0])
	address := args[1]

	cli, err := smtp.Dial(host)

	if err != nil {
		fmt.Println("Error", err)
		os.Exit(2)
	}

	merr := cli.Mail("foo@mysite.com")
	if merr != nil {
		fmt.Println("Error", merr)
		os.Exit(2)
	}

	rcpterr := cli.Rcpt(address)
	if rcpterr != nil {
		fmt.Printf("Address %s is probably invalid\n", address)
		fmt.Printf("(Server said: %s)\n", rcpterr)
		os.Exit(2)
	} else {
		fmt.Printf("Address %s is valid\n", address)
	}

	reseterr := cli.Reset()
	if reseterr != nil {
		fmt.Println("Error", reseterr)
		os.Exit(2)
	}

	qerr := cli.Quit()
	if qerr != nil {
		fmt.Println("Error", qerr)
		os.Exit(2)
	}

	cli.Close()
}
