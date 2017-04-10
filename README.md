[![Build Status](https://travis-ci.org/ransford/vrf.svg?branch=master)](https://travis-ci.org/ransford/vrf) [![Go Report Card](https://goreportcard.com/badge/github.com/ransford/vrf)](https://goreportcard.com/report/github.com/ransford/vrf)

# vrf: An SMTP Address Verifier

Checks whether a single email address is deliverable.

Is `foo@bar.quux` a deliverable email address?

    $ vrf foo@bar.quux
    2017/04/10 14:38:21 Address: foo@bar.quux
    2017/04/10 14:38:21 Domain: bar.quux
    2017/04/10 14:38:21 MX host: mail-in-baz.bar.quux
    foo@bar.quux is deliverable

What about `oiwperwer@google.com`?

    $ vrf oiwperwer@google.com
    2017/04/10 14:39:51 Address: oiwperwer@google.com
    2017/04/10 14:39:51 Domain: google.com
    2017/04/10 14:39:51 MX host: aspmx.l.google.com.
    2017/04/10 14:39:51 Error on RCPT: 550 5.1.1 The email account that you tried to reach does not exist. Please try
    5.1.1 double-checking the recipient's email address for typos or
    5.1.1 unnecessary spaces. Learn more at
    5.1.1  https://support.google.com/mail/?p=NoSuchUser 64si14696711pfl.160 - gsmtp
    2017/04/10 14:39:51 Error checking deliverability: 550 5.1.1 The email account that you tried to reach does not exist. Please try
    5.1.1 double-checking the recipient's email address for typos or
    5.1.1 unnecessary spaces. Learn more at
    5.1.1  https://support.google.com/mail/?p=NoSuchUser 64si14696711pfl.160 - gsmtp

# How it Works

`vrf` looks up the mail exchanger (MX) records for a given address, then
connects to the highest priority server in the list and goes partway through
the process of delivering an email messsage.  Essentially:

> Hello `domain.com` mail server, I have email for user `bob@domain.com`.  Get
> ready to accept it!  Oh, never mind.  Bye!
