package ofx

import "errors"

var (
	ErrFileNotSent      = errors.New("invalid file sent in request")
	ErrFileNotSupported = errors.New("this file extension is not supported by the server")
	ErrFailedOpenFile   = errors.New("file could not be read by the server")
	ErrFailedParseFile  = errors.New("failed to parse file during server")
)
