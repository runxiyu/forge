package irc

import "errors"

var (
	ErrInvalidIRCv3Tag = errors.New("invalid ircv3 tag")
	ErrMalformedMsg    = errors.New("malformed irc message")
)
