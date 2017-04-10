[![Build Status](https://travis-ci.org/ransford/vrf.svg?branch=master)](https://travis-ci.org/ransford/vrf) [![Go Report Card](https://goreportcard.com/badge/github.com/ransford/vrf)](https://goreportcard.com/report/github.com/ransford/vrf)

# vrf: An SMTP Address Verifier

Checks whether a single email address is deliverable.  Exits successfully if so.

Is `foo@bar.quux` a deliverable email address?

    $ vrf foo@bar.quux
    foo@bar.quux is deliverable
    $ vrf -quiet foo@bar.quux && echo "yes"
    yes

What about `oiwperwer@google.com`?

    $ vrf oiwperwer@google.com
    oiwperwer@google.com is not deliverable
    $ vrf -quiet oiwperwer@google.com || echo "no"
    no

# How it Works

`vrf` looks up the mail exchanger (MX) records for a given address, then
connects to the highest priority server in the list and goes partway through
the process of delivering an email messsage.  Essentially:

> Client: Hello `domain.com` mail server!
> Server: Hello!
> Client: I have email for user `bob@domain.com`.
> Server: Sure, I'll accept it.  ***OR***  Sorry, no such address.
> Client: Never mind.
> Server: Fine, whatever.
> Client: Bye!
> Server: Bye.

Technically speaking, `vrf` disconnects (politely, with `RSET`) after `MAIL`.
