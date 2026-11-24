package drisl

import (
	"codeberg.org/reiver/go-erorr"

	"github.com/reiver/go-cid"
)

// CID is a DASL CID (content identifier), in binary form — as found in DRISL tag 42 (after the leading 0x00 byte).
//
// It is the same type as [cid.CID] — see that package for parsing, making, and verifying CIDs.
//
// See: https://dasl.ing/cid.html
type CID = cid.CID

// invalidCID returns an error that is both [ErrInvalidCID] and 'err' (the reason, from [cid.CID.Validate]).
func invalidCID(err error) error {
	return erorr.Errorf("%w: %w", ErrInvalidCID, err)
}
