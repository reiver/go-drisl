/*
Package drisl implements DRISL — "Deterministic Representation for Interoperable Structures & Links" — for the Go programming language.

DRISL is a deterministic profile of CBOR (IETF RFC-8949), with native support for CIDs (content identifiers) as links.
It is the encoding used for records in the AT-Protocol (ATProto).
It is defined here:

https://dasl.ing/drisl.html

Because DRISL is deterministic, the same data always has the same bytes — and so the same CID.

# Strictness

[Unmarshal] rejects anything that is not valid DRISL, including:

	• integers, lengths, counts, and tag numbers not encoded in the fewest bytes possible,
	• map keys that are not text strings, are not in order, or repeat,
	• any tag other than 42 (CID), and a tag 42 whose content is not a valid DASL CID,
	• indefinite-length byte strings, text strings, arrays, and maps,
	• floating-point numbers not encoded in 64 bits, and NaN, Infinity, -Infinity, and -0,
	• simple values other than false, true, and null,
	• text strings that are not valid UTF-8,
	• anything after the data item.

# Go Types

	┌─────────────────────────────┬──────────────────────────────────────────┐
	│ DRISL                       │ Go                                       │
	├─────────────────────────────┼──────────────────────────────────────────┤
	│ null                        │ nil                                      │
	│ false, true                 │ bool                                     │
	│ integer (fits in an int64)  │ int64                                    │
	│ integer (otherwise)         │ *big.Int                                 │
	│ 64-bit floating-point       │ float64                                  │
	│ text string                 │ string                                   │
	│ byte string                 │ []byte                                   │
	│ tag 42 (CID)                │ CID                                      │
	│ array                       │ []any                                    │
	│ map                         │ map[string]any                           │
	└─────────────────────────────┴──────────────────────────────────────────┘

[Marshal] accepts those types (and also the other Go integer types, and float32).

# JSON

[ToJSON] renders a value as JSON, using the AT-Protocol conventions:
a byte string becomes {"$bytes": "<base64>"}, and a CID becomes {"$link": "<CID string>"}.
*/
package drisl
