package drisl

import (
	"encoding/base32"
	"strings"
)

// CID is a DASL CID (content identifier), in binary form — as found in DRISL tag 42 (after the leading 0x00 byte).
//
// A DASL CID is:
//
//	• version 1 (0x01),
//	• a codec of raw (0x55) or DRISL / dag-cbor (0x71),
//	• a hash of SHA-256 (0x12) or BLAKE3 (0x1e),
//	• a digest length of 32 bytes (0x20),
//	• and the 32-byte digest.
//
// See: https://dasl.ing/cid.html
type CID []byte

var base32LowerNoPadding = base32.StdEncoding.WithPadding(base32.NoPadding)

// String returns the CID's string form: "b" followed by its lower-case base32 encoding (no padding).
//
// Example usage:
//
//	var cid drisl.CID = // ...
//	
//	s := cid.String()
//	
//	// s == "bafyrei..."
func (receiver CID) String() string {
	return "b" + strings.ToLower(base32LowerNoPadding.EncodeToString(receiver))
}

// Validate returns an error if the CID is not a valid DASL CID.
func (receiver CID) Validate() error {
	const length = 4 + 32

	if nil == receiver {
		return ErrNilCID
	}
	if length != len(receiver) {
		return ErrInvalidCID
	}

	// version
	if 0x01 != receiver[0] {
		return ErrInvalidCID
	}

	// codec
	switch receiver[1] {
	case 0x55, 0x71: // raw, DRISL (dag-cbor)
		// Nothing here.
	default:
		return ErrInvalidCID
	}

	// hash
	switch receiver[2] {
	case 0x12, 0x1e: // SHA-256, BLAKE3
		// Nothing here.
	default:
		return ErrInvalidCID
	}

	// digest length
	if 0x20 != receiver[3] {
		return ErrInvalidCID
	}

	return nil
}
