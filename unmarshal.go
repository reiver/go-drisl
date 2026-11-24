package drisl

import (
	"bytes"
	"math"
	"math/big"

	"codeberg.org/reiver/go-erorr"

	"github.com/reiver/go-rfc8949/head"
	"github.com/reiver/go-rfc8949/initialbyte/additionalinfo"
	"github.com/reiver/go-rfc8949/majortype"
	"github.com/reiver/go-rfc8949/types/bytestrings"
	"github.com/reiver/go-rfc8949/types/float64s"
	"github.com/reiver/go-rfc8949/types/textstrings"
)

const tagNumberCID uint64 = 42

// Unmarshal decodes 'p', which must be exactly one DRISL data item.
//
// It returns one of: nil, bool, int64, *big.Int, float64, string, []byte, [CID], []any, map[string]any.
// (Arrays and maps contain those same types.)
//
// Unmarshal rejects anything that is not valid DRISL. (See the package documentation for the list.)
//
// Example usage:
//
//	value, err := drisl.Unmarshal(p)
//	if nil != err {
//		return err
//	}
//
//	record, casted := value.(map[string]any)
func Unmarshal(p []byte) (any, error) {
	value, rest, err := unmarshalItem(p, 0)
	if nil != err {
		return nil, err
	}
	if 0 < len(rest) {
		return nil, ErrTrailingBytes
	}

	return value, nil
}

func unmarshalItem(p []byte, depth int) (any, []byte, error) {
	if MaxDepth < depth {
		return nil, p, ErrTooDeep
	}

	h, err := head.Decode(p)
	if nil != err {
		return nil, p, translateError(err)
	}
	if h.IsIndefinite() {
		return nil, p, ErrIndefiniteLength
	}
	if !h.IsMinimal() {
		return nil, p, ErrNotMinimal
	}

	switch h.MajorType {
	case majortype.UnsignedInteger:
		if h.Argument <= math.MaxInt64 {
			return int64(h.Argument), p[h.Length:], nil
		}
		return new(big.Int).SetUint64(h.Argument), p[h.Length:], nil

	case majortype.NegativeInteger:
		// The value is -1 minus the argument.
		if h.Argument <= math.MaxInt64 {
			return -1 - int64(h.Argument), p[h.Length:], nil
		}
		var value *big.Int = new(big.Int).SetUint64(h.Argument)
		value.Neg(value)
		value.Sub(value, big.NewInt(1))
		return value, p[h.Length:], nil

	case majortype.ByteString:
		value, rest, err := bytestrings.Unmarshal(p)
		if nil != err {
			return nil, p, translateError(err)
		}
		return value, rest, nil

	case majortype.TextString:
		value, rest, err := textstrings.Unmarshal(p)
		if nil != err {
			return nil, p, translateError(err)
		}
		return value, rest, nil

	case majortype.Array:
		return unmarshalArray(h, p, depth)

	case majortype.Map:
		return unmarshalMap(h, p, depth)

	case majortype.Tagged:
		return unmarshalTagged(h, p)

	case majortype.MajorType7:
		return unmarshalMajorType7(h, p)

	default:
		// This should never be reached; the major type is 3 bits, and all 8 values are covered above.
		return nil, p, erorr.Errorf("drisl: unknown major type %d", h.MajorType>>5)
	}
}

func unmarshalArray(h head.Head, p []byte, depth int) (any, []byte, error) {
	var rest []byte = p[h.Length:]

	// Every data item takes at least 1 byte.
	// (This stops a crafted count from causing a huge allocation.)
	if uint64(len(rest)) < h.Argument {
		return nil, p, ErrUnexpectedEnd
	}

	var array []any = make([]any, 0, int(h.Argument))
	for i:=uint64(0); i<h.Argument; i++ {
		var item any
		var err error

		item, rest, err = unmarshalItem(rest, depth+1)
		if nil != err {
			return nil, p, err
		}

		array = append(array, item)
	}

	return array, rest, nil
}

func unmarshalMap(h head.Head, p []byte, depth int) (any, []byte, error) {
	var rest []byte = p[h.Length:]

	// Every key and every value takes at least 1 byte.
	// (This stops a crafted count from causing a huge allocation.)
	if uint64(len(rest))/2 < h.Argument {
		return nil, p, ErrUnexpectedEnd
	}

	var m map[string]any = make(map[string]any, int(h.Argument))

	var previousEncodedKey []byte
	for i:=uint64(0); i<h.Argument; i++ {

		var key string
		{
			keyHead, err := head.Decode(rest)
			if nil != err {
				return nil, p, translateError(err)
			}
			if majortype.TextString != keyHead.MajorType {
				return nil, p, ErrNonStringKey
			}
			if keyHead.IsIndefinite() {
				return nil, p, ErrIndefiniteLength
			}
			if !keyHead.IsMinimal() {
				return nil, p, ErrNotMinimal
			}

			var afterKey []byte
			key, afterKey, err = textstrings.Unmarshal(rest)
			if nil != err {
				return nil, p, translateError(err)
			}

			// Keys must be in order — by their encoded bytes — with no duplicates.
			// (For text strings, this is the same as "shorter first; then byte-wise".)
			var encodedKey []byte = rest[:len(rest)-len(afterKey)]
			if nil != previousEncodedKey {
				switch bytes.Compare(previousEncodedKey, encodedKey) {
				case 0:
					return nil, p, ErrDuplicateKey
				case 1:
					return nil, p, ErrKeysNotInOrder
				}
			}
			previousEncodedKey = encodedKey

			rest = afterKey
		}

		{
			var value any
			var err error

			value, rest, err = unmarshalItem(rest, depth+1)
			if nil != err {
				return nil, p, err
			}

			m[key] = value
		}
	}

	return m, rest, nil
}

func unmarshalTagged(h head.Head, p []byte) (any, []byte, error) {
	if tagNumberCID != h.Argument {
		return nil, p, ErrUnsupportedTag
	}

	var rest []byte = p[h.Length:]

	contentHead, err := head.Decode(rest)
	if nil != err {
		return nil, p, translateError(err)
	}
	if majortype.ByteString != contentHead.MajorType {
		return nil, p, ErrInvalidCID
	}
	if contentHead.IsIndefinite() {
		return nil, p, ErrIndefiniteLength
	}
	if !contentHead.IsMinimal() {
		return nil, p, ErrNotMinimal
	}

	content, rest, err := bytestrings.Unmarshal(rest)
	if nil != err {
		return nil, p, translateError(err)
	}

	// For historical reasons, the binary CID is prefixed with a 0x00 byte.
	if len(content) < 1 || 0x00 != content[0] {
		return nil, p, ErrInvalidCID
	}

	var cid CID = CID(content[1:])
	if err := cid.Validate(); nil != err {
		return nil, p, invalidCID(err)
	}

	return cid, rest, nil
}

func unmarshalMajorType7(h head.Head, p []byte) (any, []byte, error) {
	switch h.AdditionalInformation {
	case 20:
		return false, p[h.Length:], nil
	case 21:
		return true, p[h.Length:], nil
	case 22:
		return nil, p[h.Length:], nil
	case additionalinfo.Uint16, additionalinfo.Uint32:
		return nil, p, ErrFloatNot64Bit
	case additionalinfo.Uint64:
		value, _, rest, err := float64s.Unmarshal(p)
		if nil != err {
			return nil, p, translateError(err)
		}
		if math.IsNaN(value) || math.IsInf(value, 0) || (0 == value && math.Signbit(value)) {
			return nil, p, ErrFloatNotAllowed
		}
		return value, rest, nil
	default:
		return nil, p, ErrSimpleValue
	}
}

// translateError turns the "ran out of data" errors from package rfc8949 into this package's error,
// and passes the others through.
func translateError(err error) error {
	switch err {
	case head.ErrEmpty, head.ErrUnexpectedEnd:
		return ErrUnexpectedEnd
	case bytestrings.ErrIndefiniteLength, textstrings.ErrIndefiniteLength:
		return ErrIndefiniteLength
	default:
		return err
	}
}
