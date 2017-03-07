package main

import "testing"

func TestAddressParse(t *testing.T) {
	/* No errors on good addresses */
	addrs := []struct {
		addr   string
		domain string
	}{
		{"foo@google.com", "google.com"},
		{"p@grrransford.org", "grrransford.org"},
		{"@foot.com", "foot.com"},
		{"@bar", "bar"},
		{"foo@", ""},
		{"bl@h@blah@blah.com", "blah.com"},
	}
	for _, tcase := range addrs {
		_, err := getDomainFromAddress(tcase.addr)
		if err != nil {
			t.Fatal("Error should be nil, but is", err)
		}
	}

	/* Errors on bad addresses */
	badAddrs := []string{
		"foo",
		"foo.com",
		"",
	}
	for _, badaddr := range badAddrs {
		_, err := getDomainFromAddress(badaddr)
		if err == nil {
			t.Fatal("err is nil; shoul be non-nil")
		}
	}
}
