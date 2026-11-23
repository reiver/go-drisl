package drisl_test

import (
	"math/big"
	"testing"

	"github.com/reiver/go-drisl"
)

func TestToJSON(t *testing.T) {

	tests := []struct{
		Value any
		Expected string
	}{
		{
			Value:    nil,
			Expected: `null`,
		},
		{
			Value:    int64(-7),
			Expected: `-7`,
		},
		{
			Value:    new(big.Int).SetUint64(18446744073709551615),
			Expected: `18446744073709551615`,
		},
		{
			Value:    0.25,
			Expected: `0.25`,
		},
		{
			// HTML characters are not escaped.
			Value:    "Hello <world> & friends",
			Expected: `"Hello <world> & friends"`,
		},
		{
			Value:    []byte{0xde, 0xad, 0xbe, 0xef},
			Expected: `{"$bytes":"3q2+7w"}`,
		},
		{
			Value:    drisl.CID(mustHex(t, "015512205891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03")),
			Expected: `{"$link":"bafkreicysg23kiwv34eg2d7qweipxwosdo2py4ldv42nbauguluen5v6am"}`,
		},
		{
			Value: map[string]any{
				"$type": "blob",
				"ref":   drisl.CID(mustHex(t, "015512205891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03")),
				"mimeType": "image/png",
				"size":  int64(12345),
			},
			Expected: `{"$type":"blob","mimeType":"image/png","ref":{"$link":"bafkreicysg23kiwv34eg2d7qweipxwosdo2py4ldv42nbauguluen5v6am"},"size":12345}`,
		},
		{
			Value:    []any{int64(1), "two", []any{}},
			Expected: `[1,"two",[]]`,
		},
	}

	for testNumber, test := range tests {

		actual, err := drisl.ToJSON(test.Value)
		if nil != err {
			t.Errorf("For test #%d, did not expect an error but actually got one.", testNumber)
			t.Logf("ERROR: (%T) %s", err, err)
			continue
		}

		{
			expected := test.Expected

			if expected != string(actual) {
				t.Errorf("For test #%d, the actual JSON is not what was expected.", testNumber)
				t.Logf("EXPECTED: %s", expected)
				t.Logf("ACTUAL:   %s", actual)
				continue
			}
		}
	}
}

// TestToJSON_fromDRISL decodes DRISL bytes and renders them as JSON, end to end.
func TestToJSON_fromDRISL(t *testing.T) {

	// {"a": h'01', "b": 42(cid)}
	p := mustHex(t, "a2616141016162d82a582500015512205891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03")

	value, err := drisl.Unmarshal(p)
	if nil != err {
		t.Fatalf("Did not expect an error but actually got one: (%T) %s", err, err)
	}

	actual, err := drisl.ToJSON(value)
	if nil != err {
		t.Fatalf("Did not expect an error but actually got one: (%T) %s", err, err)
	}

	const expected string = `{"a":{"$bytes":"AQ"},"b":{"$link":"bafkreicysg23kiwv34eg2d7qweipxwosdo2py4ldv42nbauguluen5v6am"}}`

	if expected != string(actual) {
		t.Errorf("The actual JSON is not what was expected.")
		t.Logf("EXPECTED: %s", expected)
		t.Logf("ACTUAL:   %s", actual)
	}
}
