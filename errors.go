package drisl

import (
	"codeberg.org/reiver/go-erorr"
)

const (
	ErrDuplicateKey         = erorr.Error("drisl: duplicate map key")
	ErrFloatNot64Bit        = erorr.Error("drisl: floating-point number not encoded in 64 bits")
	ErrFloatNotAllowed      = erorr.Error("drisl: floating-point NaN, Infinity, -Infinity, and -0 are not allowed")
	ErrIndefiniteLength     = erorr.Error("drisl: indefinite-length data item not allowed")
	ErrIntegerOutOfRange    = erorr.Error("drisl: integer out of range — must be in -2⁶⁴..2⁶⁴-1")
	ErrInvalidCID           = erorr.Error("drisl: invalid CID")
	ErrKeysNotInOrder       = erorr.Error("drisl: map keys not in order")
	ErrNilCID               = erorr.Error("drisl: nil CID")
	ErrNonStringKey         = erorr.Error("drisl: map key is not a text string")
	ErrNotMinimal           = erorr.Error("drisl: not encoded in the fewest bytes possible")
	ErrSimpleValue          = erorr.Error("drisl: simple value not allowed — only false, true, and null are")
	ErrTooDeep              = erorr.Error("drisl: nested too deeply")
	ErrTrailingBytes        = erorr.Error("drisl: bytes after the data item")
	ErrUnexpectedEnd        = erorr.Error("drisl: unexpected end of data")
	ErrUnsupportedTag       = erorr.Error("drisl: tag not allowed — only tag 42 (CID) is")
	ErrUnsupportedType      = erorr.Error("drisl: unsupported Go type")
)
