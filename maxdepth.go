package drisl

// MaxDepth is the deepest that arrays and maps may be nested, when unmarshaling and marshaling.
//
// It guards against data crafted to exhaust the stack.
const MaxDepth = 128
