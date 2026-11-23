package drisl

import (
	"bytes"
	"math"
	"math/big"
	"sort"

	"codeberg.org/reiver/go-erorr"

	"github.com/reiver/go-rfc8949/majortype"
	"github.com/reiver/go-rfc8949/types/bools"
	"github.com/reiver/go-rfc8949/types/bytestrings"
	"github.com/reiver/go-rfc8949/types/float64s"
	"github.com/reiver/go-rfc8949/types/int64s"
	"github.com/reiver/go-rfc8949/types/nils"
	"github.com/reiver/go-rfc8949/types/textstrings"
	"github.com/reiver/go-rfc8949/types/uint64s"
)

// Marshal encodes 'value' as DRISL.
//
// 'value' may be: nil, bool, any Go integer type, *big.Int, float32, float64, string, []byte, [CID], []any, map[string]any.
// (Arrays and maps may contain those same types.)
//
// The result is deterministic: integers, lengths, and counts use the fewest bytes possible;
// floating-point numbers use 64 bits; and map keys are in order.
//
// Marshal returns an error for: an integer outside -2⁶⁴..2⁶⁴-1; NaN, Infinity, -Infinity, or -0;
// a string that is not valid UTF-8; an invalid CID; and any other Go type.
//
// Example usage:
//
//	p, err := drisl.Marshal(map[string]any{
//		"$type": "com.example.post",
//		"text":  "Hello world!",
//	})
func Marshal(value any) ([]byte, error) {
	return marshalItem(value, 0)
}

func marshalItem(value any, depth int) ([]byte, error) {
	if MaxDepth < depth {
		return nil, ErrTooDeep
	}

	switch casted := value.(type) {
	case nil:
		return nils.Marshal()
	case bool:
		return bools.Marshal(casted)

	case int:
		return int64s.Marshal(int64(casted))
	case int8:
		return int64s.Marshal(int64(casted))
	case int16:
		return int64s.Marshal(int64(casted))
	case int32:
		return int64s.Marshal(int64(casted))
	case int64:
		return int64s.Marshal(casted)
	case uint:
		return uint64s.Marshal(uint64(casted))
	case uint8:
		return uint64s.Marshal(uint64(casted))
	case uint16:
		return uint64s.Marshal(uint64(casted))
	case uint32:
		return uint64s.Marshal(uint64(casted))
	case uint64:
		return uint64s.Marshal(casted)
	case *big.Int:
		return marshalBigInt(casted)

	case float32:
		return marshalFloat64(float64(casted))
	case float64:
		return marshalFloat64(casted)

	case string:
		return textstrings.Marshal(casted)
	case []byte:
		return bytestrings.Marshal(casted)
	case CID:
		return marshalCID(casted)

	case []any:
		return marshalArray(casted, depth)
	case map[string]any:
		return marshalMap(casted, depth)

	default:
		return nil, erorr.Errorf("drisl: cannot marshal value of type %T: %w", value, ErrUnsupportedType)
	}
}

// marshalHead encodes the head of a data item with major type 'majorType' and argument 'argument',
// using the fewest bytes possible.
//
// (The head of an unsigned integer has exactly this shape; only the major type differs.)
func marshalHead(majorType byte, argument uint64) []byte {
	p, _ := uint64s.Marshal(argument)
	p[0] = (p[0] & 0b000_11111) | majorType
	return p
}

var (
	bigMaxUint64    = new(big.Int).SetUint64(math.MaxUint64)
	bigMinNegative  = new(big.Int).Neg(new(big.Int).Add(bigMaxUint64, big.NewInt(1))) // -2⁶⁴
)

func marshalBigInt(value *big.Int) ([]byte, error) {
	if nil == value {
		return nil, erorr.Errorf("drisl: cannot marshal a nil *big.Int: %w", ErrUnsupportedType)
	}

	switch {
	case value.IsInt64():
		return int64s.Marshal(value.Int64())
	case 0 < value.Sign() && value.Cmp(bigMaxUint64) <= 0:
		return uint64s.Marshal(value.Uint64())
	case value.Sign() < 0 && bigMinNegative.Cmp(value) <= 0:
		// The argument of a negative integer is -1 minus the value.
		var argument *big.Int = new(big.Int).Neg(value)
		argument.Sub(argument, big.NewInt(1))
		return marshalHead(majortype.NegativeInteger, argument.Uint64()), nil
	default:
		return nil, ErrIntegerOutOfRange
	}
}

func marshalFloat64(value float64) ([]byte, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) || (0 == value && math.Signbit(value)) {
		return nil, ErrFloatNotAllowed
	}

	return float64s.Marshal(value)
}

func marshalCID(cid CID) ([]byte, error) {
	if err := cid.Validate(); nil != err {
		return nil, err
	}

	// For historical reasons, the binary CID is prefixed with a 0x00 byte.
	var content []byte = make([]byte, 0, 1+len(cid))
	content = append(content, 0x00)
	content = append(content, cid...)

	encodedContent, err := bytestrings.Marshal(content)
	if nil != err {
		return nil, err
	}

	var result []byte = marshalHead(majortype.Tagged, tagNumberCID)
	result = append(result, encodedContent...)
	return result, nil
}

func marshalArray(array []any, depth int) ([]byte, error) {
	var result []byte = marshalHead(majortype.Array, uint64(len(array)))

	for index, item := range array {
		encoded, err := marshalItem(item, depth+1)
		if nil != err {
			return nil, erorr.Errorf("drisl: array item [%d]: %w", index, err)
		}
		result = append(result, encoded...)
	}

	return result, nil
}

func marshalMap(m map[string]any, depth int) ([]byte, error) {
	type internalPair struct {
		encodedKey   []byte
		encodedValue []byte
	}

	var pairs []internalPair = make([]internalPair, 0, len(m))

	for key, value := range m {
		encodedKey, err := textstrings.Marshal(key)
		if nil != err {
			return nil, erorr.Errorf("drisl: map key %q: %w", key, err)
		}

		encodedValue, err := marshalItem(value, depth+1)
		if nil != err {
			return nil, erorr.Errorf("drisl: map value for key %q: %w", key, err)
		}

		pairs = append(pairs, internalPair{
			encodedKey:   encodedKey,
			encodedValue: encodedValue,
		})
	}

	// Keys are put in order by their encoded bytes.
	// (For text strings, this is the same as "shorter first; then byte-wise".)
	sort.Slice(pairs, func(i, j int) bool {
		return bytes.Compare(pairs[i].encodedKey, pairs[j].encodedKey) < 0
	})

	var result []byte = marshalHead(majortype.Map, uint64(len(pairs)))
	for _, pair := range pairs {
		result = append(result, pair.encodedKey...)
		result = append(result, pair.encodedValue...)
	}

	return result, nil
}
