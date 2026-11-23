package drisl_test

import (
	"testing"

	"github.com/reiver/go-drisl"
)

// The expected strings were computed independently (with Python's base64.b32encode).
func TestCID_String(t *testing.T) {

	tests := []struct{
		CID string
		Expected string
	}{
		{
			// raw, SHA-256
			CID:      "015512205891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03",
			Expected: "bafkreicysg23kiwv34eg2d7qweipxwosdo2py4ldv42nbauguluen5v6am",
		},
		{
			// raw, BLAKE3
			CID:      "01551e208e4c7c1b99dbfd50e7a95185fead5ee1448fa904a2fdd778eaf5f2dbfd629a99",
			Expected: "bafkr4ieojr6bxgo37viopkkrqx7k2xxbish2sbfc7xlxr2xv6ln72yu2te",
		},
	}

	for testNumber, test := range tests {

		var cid drisl.CID = drisl.CID(mustHex(t, test.CID))

		if err := cid.Validate(); nil != err {
			t.Errorf("For test #%d, did not expect an error but actually got one: (%T) %s", testNumber, err, err)
			continue
		}

		actual := cid.String()

		if test.Expected != actual {
			t.Errorf("For test #%d, the actual CID string is not what was expected.", testNumber)
			t.Logf("EXPECTED: %s", test.Expected)
			t.Logf("ACTUAL:   %s", actual)
			continue
		}
	}
}

func TestCID_Validate_errors(t *testing.T) {

	tests := []struct{
		CID drisl.CID
		Expected error
	}{
		{
			CID:      nil,
			Expected: drisl.ErrNilCID,
		},
		{
			CID:      drisl.CID{},
			Expected: drisl.ErrInvalidCID,
		},
		{
			// codec 0x70 (dag-pb) is not a DASL codec
			CID:      drisl.CID(mustHex(t, "01701220e9822efc7c48027a5429fd6b4ea0ce4f0bd76a3a9e1d0b7f0abae39f1d2e1a4f")),
			Expected: drisl.ErrInvalidCID,
		},
		{
			// CIDv0
			CID:      drisl.CID(mustHex(t, "122022ad631c69ee983095b5b8acd029ff94aff1dc6c48837878589a92b90dfea317")),
			Expected: drisl.ErrInvalidCID,
		},
	}

	for testNumber, test := range tests {

		actual := test.CID.Validate()

		if test.Expected != actual {
			t.Errorf("For test #%d, the actual error is not what was expected.", testNumber)
			t.Logf("EXPECTED: %v", test.Expected)
			t.Logf("ACTUAL:   %v", actual)
			continue
		}
	}
}
