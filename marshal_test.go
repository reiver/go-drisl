package drisl_test

import (
	"bytes"
	"encoding/hex"
	"errors"
	"math"
	"math/big"
	"reflect"
	"testing"

	"github.com/reiver/go-drisl"
)

func TestMarshal(t *testing.T) {

	tests := []struct{
		Value any
		Expected string
	}{
		{
			Value:    nil,
			Expected: "f6",
		},
		{
			Value:    true,
			Expected: "f5",
		},



		{
			Value:    0,
			Expected: "00",
		},
		{
			Value:    int8(-1),
			Expected: "20",
		},
		{
			Value:    uint16(1000),
			Expected: "1903e8",
		},
		{
			Value:    int64(-1000),
			Expected: "3903e7",
		},
		{
			Value:    new(big.Int).SetUint64(18446744073709551615),
			Expected: "1bffffffffffffffff",
		},
		{
			Value:    new(big.Int).Neg(new(big.Int).Add(new(big.Int).SetUint64(18446744073709551615), big.NewInt(1))),
			Expected: "3bffffffffffffffff",
		},
		{
			// A *big.Int that fits in an int64 is encoded like an int64.
			Value:    big.NewInt(-1000),
			Expected: "3903e7",
		},



		{
			// Always 64 bits.
			Value:    1.5,
			Expected: "fb3ff8000000000000",
		},
		{
			Value:    float32(1.5),
			Expected: "fb3ff8000000000000",
		},
		{
			Value:    0.0,
			Expected: "fb0000000000000000",
		},



		{
			Value:    "IETF",
			Expected: "6449455446",
		},
		{
			Value:    []byte{0x01, 0x02, 0x03, 0x04},
			Expected: "4401020304",
		},
		{
			Value:    drisl.CID(mustHex(t, "015512205891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03")),
			Expected: "d82a582500015512205891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03",
		},



		{
			Value:    []any{1, 2, 3},
			Expected: "83010203",
		},
		{
			Value:    map[string]any{},
			Expected: "a0",
		},
		{
			// Keys come out shorter-first, then byte-wise.
			Value:    map[string]any{"aa": 1, "b": 2, "a": 0},
			Expected: "a3616100616202626161" + "01",
		},
	}

	for testNumber, test := range tests {

		actual, err := drisl.Marshal(test.Value)
		if nil != err {
			t.Errorf("For test #%d, did not expect an error but actually got one.", testNumber)
			t.Logf("ERROR: (%T) %s", err, err)
			t.Logf("VALUE: (%T) %#v", test.Value, test.Value)
			continue
		}

		{
			expected := mustHex(t, test.Expected)

			if !bytes.Equal(expected, actual) {
				t.Errorf("For test #%d, the actual encoding is not what was expected.", testNumber)
				t.Logf("EXPECTED: %x", expected)
				t.Logf("ACTUAL:   %x", actual)
				t.Logf("VALUE: (%T) %#v", test.Value, test.Value)
				continue
			}
		}
	}
}

func TestMarshal_errors(t *testing.T) {

	tooBig := new(big.Int).Add(new(big.Int).SetUint64(18446744073709551615), big.NewInt(1))
	tooSmall := new(big.Int).Neg(new(big.Int).Add(tooBig, big.NewInt(1)))

	tests := []struct{
		Value any
		Expected error
	}{
		{
			Value:    math.NaN(),
			Expected: drisl.ErrFloatNotAllowed,
		},
		{
			Value:    math.Inf(1),
			Expected: drisl.ErrFloatNotAllowed,
		},
		{
			Value:    math.Inf(-1),
			Expected: drisl.ErrFloatNotAllowed,
		},
		{
			Value:    math.Copysign(0, -1),
			Expected: drisl.ErrFloatNotAllowed,
		},
		{
			Value:    tooBig,
			Expected: drisl.ErrIntegerOutOfRange,
		},
		{
			Value:    tooSmall,
			Expected: drisl.ErrIntegerOutOfRange,
		},
		{
			Value:    drisl.CID{0x01, 0x70, 0x12, 0x20},
			Expected: drisl.ErrInvalidCID,
		},
	}

	for testNumber, test := range tests {

		_, err := drisl.Marshal(test.Value)

		if !errors.Is(err, test.Expected) {
			t.Errorf("For test #%d, the actual error is not what was expected.", testNumber)
			t.Logf("EXPECTED: %v", test.Expected)
			t.Logf("ACTUAL:   %v", err)
			t.Logf("VALUE: (%T) %#v", test.Value, test.Value)
			continue
		}
	}
}

func TestMarshal_errorsWrapped(t *testing.T) {

	tests := []any{
		struct{}{},
		map[int]any{1: 2},
		[]any{struct{}{}},
		map[string]any{"a": math.NaN()},
		"\xc3\x28", // not valid UTF-8
	}

	for testNumber, test := range tests {

		_, err := drisl.Marshal(test)
		if nil == err {
			t.Errorf("For test #%d, expected an error but did not actually get one.", testNumber)
			t.Logf("VALUE: (%T) %#v", test, test)
			continue
		}
	}
}

// TestMarshal_roundTrip makes sure Marshal and Unmarshal are inverses for a record-like value.
func TestMarshal_roundTrip(t *testing.T) {

	var value any = map[string]any{
		"$type": "com.atmoboards.thread",
		"title": "Hello <world> & friends",
		"createdAt": "2026-11-23T12:00:00.000Z",
		"tags": []any{"go", "atproto"},
		"count": int64(42),
		"negative": int64(-7),
		"ratio": 0.25,
		"draft": false,
		"note": nil,
		"blob": map[string]any{
			"$type": "blob",
			"ref": drisl.CID(mustHex(t, "015512205891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03")),
			"mimeType": "image/png",
			"size": int64(12345),
		},
		"raw": []byte{0xde, 0xad, 0xbe, 0xef},
	}

	p, err := drisl.Marshal(value)
	if nil != err {
		t.Fatalf("Did not expect an error from Marshal but actually got one: (%T) %s", err, err)
	}

	actual, err := drisl.Unmarshal(p)
	if nil != err {
		t.Fatalf("Did not expect an error from Unmarshal but actually got one: (%T) %s", err, err)
	}

	if !reflect.DeepEqual(value, actual) {
		t.Errorf("The round-trip did not work.")
		t.Logf("EXPECTED: %#v", value)
		t.Logf("ACTUAL:   %#v", actual)
	}

	again, err := drisl.Marshal(actual)
	if nil != err {
		t.Fatalf("Did not expect an error from the second Marshal but actually got one: (%T) %s", err, err)
	}
	if !bytes.Equal(p, again) {
		t.Errorf("Marshal is not deterministic.")
		t.Logf("FIRST:  %s", hex.EncodeToString(p))
		t.Logf("SECOND: %s", hex.EncodeToString(again))
	}
}
