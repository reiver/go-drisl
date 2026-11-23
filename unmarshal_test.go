package drisl_test

import (
	"encoding/hex"
	"math/big"
	"reflect"
	"testing"

	"github.com/reiver/go-drisl"
)

func mustHex(t *testing.T, s string) []byte {
	t.Helper()

	p, err := hex.DecodeString(s)
	if nil != err {
		t.Fatalf("Did not expect an error decoding hex %q but actually got one: (%T) %s", s, err, err)
	}
	return p
}

func TestUnmarshal(t *testing.T) {

	bigMax := new(big.Int).SetUint64(18446744073709551615)
	bigMin := new(big.Int).Neg(new(big.Int).Add(bigMax, big.NewInt(1)))

	tests := []struct{
		Data string
		Expected any
	}{
		{
			Data:     "f6",
			Expected: nil,
		},
		{
			Data:     "f4",
			Expected: false,
		},
		{
			Data:     "f5",
			Expected: true,
		},



		{
			Data:     "00",
			Expected: int64(0),
		},
		{
			Data:     "1903e8",
			Expected: int64(1000),
		},
		{
			Data:     "3903e7",
			Expected: int64(-1000),
		},
		{
			Data:     "1b7fffffffffffffff",
			Expected: int64(9223372036854775807),
		},
		{
			Data:     "3b7fffffffffffffff",
			Expected: int64(-9223372036854775808),
		},
		{
			Data:     "1bffffffffffffffff",
			Expected: bigMax,
		},
		{
			Data:     "3bffffffffffffffff",
			Expected: bigMin,
		},



		{
			Data:     "fb3ff199999999999a",
			Expected: 1.1,
		},



		{
			Data:     "6449455446",
			Expected: "IETF",
		},
		{
			Data:     "4401020304",
			Expected: []byte{0x01, 0x02, 0x03, 0x04},
		},



		{
			Data:     "80",
			Expected: []any{},
		},
		{
			Data:     "83010203",
			Expected: []any{int64(1), int64(2), int64(3)},
		},
		{
			Data:     "a0",
			Expected: map[string]any{},
		},
		{
			// {"a": 1, "b": [2, 3]}
			Data:     "a26161016162820203",
			Expected: map[string]any{"a": int64(1), "b": []any{int64(2), int64(3)}},
		},
		{
			// {"b": 2, "aa": 1} — a shorter key comes first, even though "aa" < "b" byte-wise.
			Data:     "a261620262616101",
			Expected: map[string]any{"b": int64(2), "aa": int64(1)},
		},



		{
			Data:     "d82a582500015512205891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03",
			Expected: drisl.CID(mustHex(t, "015512205891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03")),
		},
	}

	for testNumber, test := range tests {

		actual, err := drisl.Unmarshal(mustHex(t, test.Data))
		if nil != err {
			t.Errorf("For test #%d, did not expect an error but actually got one.", testNumber)
			t.Logf("ERROR: (%T) %s", err, err)
			t.Logf("DATA: %s", test.Data)
			continue
		}

		{
			expected := test.Expected

			if !reflect.DeepEqual(expected, actual) {
				t.Errorf("For test #%d, the actual value is not what was expected.", testNumber)
				t.Logf("EXPECTED: (%T) %#v", expected, expected)
				t.Logf("ACTUAL:   (%T) %#v", actual, actual)
				t.Logf("DATA: %s", test.Data)
				continue
			}
		}
	}
}

func TestUnmarshal_errors(t *testing.T) {

	tests := []struct{
		Data string
		Expected error
	}{
		{
			Data:     "",
			Expected: drisl.ErrUnexpectedEnd,
		},



		{
			Data:     "1801",
			Expected: drisl.ErrNotMinimal,
		},
		{
			Data:     "b801616100",
			Expected: drisl.ErrNotMinimal,
		},
		{
			// {"a": 1}, but the key's length is not encoded minimally
			Data:     "a178016101",
			Expected: drisl.ErrNotMinimal,
		},



		{
			Data:     "a10000",
			Expected: drisl.ErrNonStringKey,
		},
		{
			Data:     "a2616201616100",
			Expected: drisl.ErrKeysNotInOrder,
		},
		{
			// {"aa": 1, "b": 2} — "b" is shorter, so must come first
			Data:     "a262616101616202",
			Expected: drisl.ErrKeysNotInOrder,
		},
		{
			Data:     "a2616100616101",
			Expected: drisl.ErrDuplicateKey,
		},



		{
			Data:     "d8207368747470733a2f2f6578616d706c652e636f6d",
			Expected: drisl.ErrUnsupportedTag,
		},
		{
			Data:     "d82a4100",
			Expected: drisl.ErrInvalidCID,
		},
		{
			// no 0x00 prefix
			Data:     "d82a5824015512205891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03",
			Expected: drisl.ErrInvalidCID,
		},



		{
			Data:     "bfff",
			Expected: drisl.ErrIndefiniteLength,
		},
		{
			Data:     "7fff",
			Expected: drisl.ErrIndefiniteLength,
		},



		{
			Data:     "f93e00",
			Expected: drisl.ErrFloatNot64Bit,
		},
		{
			Data:     "fa47c35000",
			Expected: drisl.ErrFloatNot64Bit,
		},
		{
			Data:     "fb7ff8000000000000",
			Expected: drisl.ErrFloatNotAllowed,
		},
		{
			Data:     "fb8000000000000000",
			Expected: drisl.ErrFloatNotAllowed,
		},



		{
			Data:     "f7",
			Expected: drisl.ErrSimpleValue,
		},
		{
			Data:     "e0",
			Expected: drisl.ErrSimpleValue,
		},



		{
			Data:     "0000",
			Expected: drisl.ErrTrailingBytes,
		},
		{
			// array says 3 items, has 1
			Data:     "8301",
			Expected: drisl.ErrUnexpectedEnd,
		},
		{
			// array says 2^32 items
			Data:     "9b0000000100000000",
			Expected: drisl.ErrUnexpectedEnd,
		},
	}

	for testNumber, test := range tests {

		_, err := drisl.Unmarshal(mustHex(t, test.Data))

		{
			expected := test.Expected
			actual   := err

			if expected != actual {
				t.Errorf("For test #%d, the actual error is not what was expected.", testNumber)
				t.Logf("EXPECTED: %v", expected)
				t.Logf("ACTUAL:   %v", actual)
				t.Logf("DATA: %s", test.Data)
				continue
			}
		}
	}
}

func TestUnmarshal_tooDeep(t *testing.T) {
	var p []byte
	for i:=0; i<=drisl.MaxDepth+1; i++ {
		p = append(p, 0x81) // an array of 1 item
	}
	p = append(p, 0xf6) // null

	_, err := drisl.Unmarshal(p)
	if drisl.ErrTooDeep != err {
		t.Errorf("The actual error is not what was expected.")
		t.Logf("EXPECTED: %v", drisl.ErrTooDeep)
		t.Logf("ACTUAL:   %v", err)
	}
}
