package main

import (
	"testing"
)

func TestNormalizeAddress(t *testing.T) {
	goodAddrs := []struct {
		rfc5322 string
		name    string
		addr    string
	}{
		{"Foo Bar <foo@bar.com>", "Foo Bar", "foo@bar.com"},
		{"Bar <baz@baz.info>", "Bar", "baz@baz.info"},
		{"user@domain.com", "", "user@domain.com"},
		{"user@localhost", "", "user@localhost"},
		{"<user@domain.com>", "", "user@domain.com"},
	}

	for _, addr := range goodAddrs {
		addrObj, err := normalizeAddress(addr.rfc5322)
		if err != nil {
			t.Fatalf("Failed to parse valid RFC5322 string %s: %v", addr.rfc5322, err)
		}
		if addrObj.Name != addr.name {
			t.Fatal("Name mismatch")
		}
		if addrObj.Address != addr.addr {
			t.Fatal("Address mismatch")
		}
	}
}

func TestAddressParse(t *testing.T) {
	/* No errors on good addresses */
	addrs := []struct {
		addr   string
		domain string
	}{
		{"foo@google.com", "google.com"},
		{"p@grrransford.org", "grrransford.org"},
		{"Foo Bar <foo@bar.info>", "bar.info"},
	}
	for _, tcase := range addrs {
		_, err := getDomainFromAddress(tcase.addr)
		if err != nil {
			t.Fatalf("Failed to parse valid RFC5322 address %s: %v", tcase.addr, err)
		}
	}

	/* Errors on bad addresses */
	badAddrs := []string{
		"foo",
		"<foo>",
		"foo.com",
		"@foot.com",
		"@bar",
		"foo@",
		"bl@h@blah@blah.com",
		"",
	}
	for _, badaddr := range badAddrs {
		_, err := getDomainFromAddress(badaddr)
		if err == nil {
			t.Fatal("err is nil; shoul be non-nil")
		}
	}
}
