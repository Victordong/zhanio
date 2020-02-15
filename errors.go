package zhanio

import "errors"

var (
	errProtocolNotSupported = errors.New("not supported protocol on this platform")
	errServerShutdown       = errors.New("server is going to be shutdown")
	ErrInvalidFixedLength   = errors.New("invalid fixed length of bytes")
	ErrUnexpectedEOF        = errors.New("there is no enough data")
	ErrDelimiterNotFound    = errors.New("there is no such a delimiter")
	ErrCRLFNotFound         = errors.New("there is no CRLF")
	ErrUnsupportedLength    = errors.New("unsupported lengthFieldLength. (expected: 1, 2, 3, 4, or 8)")
	ErrTooLessLength        = errors.New("adjusted frame length is less than zero")
)
