package prlp

import "errors"

var (
	ErrValueNotSupport          = errors.New("value not supported")
	ErrUnexpectedLength         = errors.New("unexpected length")
	ErrNotAList                 = errors.New("not a list")
	ErrNotAString               = errors.New("not a string")
	ErrCannotValueBeASingleByte = errors.New("cannot be a single byte")
	ErrTxTypeNotSupported       = errors.New("tx type not supported")
)
