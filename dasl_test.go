package drisl_test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/reiver/go-drisl"
)

// TestDASLFixtures runs the CBOR fixtures from the dasl-testing suite (see testdata/dasl-testing/README.md).
//
// Only the fixtures tagged "basic", "dag-cbor", or "dasl-cid" apply to DRISL.
// (The others are for other CBOR profiles — some of which disagree with DRISL.)
//
//	• "roundtrip"   — must decode, and re-encode to exactly the same bytes.
//	• "invalid_in"  — must be rejected by the decoder.
//	• "invalid_out" — must never be produced by the encoder. Since a DRISL decoder accepts only what a DRISL encoder may produce, it must also be rejected by the decoder.
func TestDASLFixtures(t *testing.T) {

	type internalFixture struct {
		Type string   `json:"type"`
		Data string   `json:"data"`
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}

	applies := func(tags []string) bool {
		for _, tag := range tags {
			switch tag {
			case "basic", "dag-cbor", "dasl-cid":
				return true
			}
		}
		return false
	}

	filenames, err := filepath.Glob(filepath.Join("testdata", "dasl-testing", "*.json"))
	if nil != err {
		t.Fatalf("Did not expect an error but actually got one: (%T) %s", err, err)
	}
	if len(filenames) <= 0 {
		t.Fatalf("Expected fixture files but did not actually find any.")
	}

	var count int

	for _, filename := range filenames {

		var fixtures []internalFixture
		{
			p, err := os.ReadFile(filename)
			if nil != err {
				t.Errorf("Did not expect an error reading %q but actually got one: (%T) %s", filename, err, err)
				continue
			}

			err = json.Unmarshal(p, &fixtures)
			if nil != err {
				t.Errorf("Did not expect an error parsing %q but actually got one: (%T) %s", filename, err, err)
				continue
			}
		}

		for testNumber, fixture := range fixtures {

			if !applies(fixture.Tags) {
				continue
			}
			count++

			data, err := hex.DecodeString(fixture.Data)
			if nil != err {
				t.Errorf("For %s test #%d (%q), did not expect an error decoding the hex but actually got one: (%T) %s", filename, testNumber, fixture.Name, err, err)
				continue
			}

			value, err := drisl.Unmarshal(data)

			switch fixture.Type {
			case "roundtrip":
				if nil != err {
					t.Errorf("For %s test #%d (%q), did not expect an error but actually got one.", filename, testNumber, fixture.Name)
					t.Logf("ERROR: (%T) %s", err, err)
					t.Logf("DATA: %s", fixture.Data)
					continue
				}

				encoded, err := drisl.Marshal(value)
				if nil != err {
					t.Errorf("For %s test #%d (%q), did not expect an error re-encoding but actually got one.", filename, testNumber, fixture.Name)
					t.Logf("ERROR: (%T) %s", err, err)
					t.Logf("DATA: %s", fixture.Data)
					continue
				}

				if !bytes.Equal(data, encoded) {
					t.Errorf("For %s test #%d (%q), the re-encoding is not what was expected.", filename, testNumber, fixture.Name)
					t.Logf("EXPECTED: %x", data)
					t.Logf("ACTUAL:   %x", encoded)
					continue
				}

			case "invalid_in", "invalid_out":
				if nil == err {
					t.Errorf("For %s test #%d (%q), expected an error but did not actually get one.", filename, testNumber, fixture.Name)
					t.Logf("TYPE: %s", fixture.Type)
					t.Logf("DATA: %s", fixture.Data)
					t.Logf("VALUE: (%T) %#v", value, value)
					continue
				}

			default:
				t.Errorf("For %s test #%d (%q), unknown fixture type %q.", filename, testNumber, fixture.Name, fixture.Type)
				continue
			}
		}
	}

	if count <= 0 {
		t.Errorf("Expected some fixtures to apply but none did.")
	}
	t.Logf("Ran %d applicable fixtures.", count)
}
