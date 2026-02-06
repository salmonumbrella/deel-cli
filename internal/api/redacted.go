package api

// RedactedString holds a sensitive string value that is hidden from fmt, JSON, and debug output.
type RedactedString struct {
	val string
}

// NewRedactedString wraps a sensitive string.
func NewRedactedString(s string) RedactedString {
	return RedactedString{val: s}
}

// Value returns the underlying string for authorized use (e.g. setting HTTP headers).
func (r RedactedString) Value() string {
	return r.val
}

// String implements fmt.Stringer — always returns a redacted placeholder.
func (r RedactedString) String() string {
	return "[REDACTED]"
}

// GoString implements fmt.GoStringer — prevents leaking via %#v.
func (r RedactedString) GoString() string {
	return "RedactedString{[REDACTED]}"
}

// MarshalJSON prevents the value from leaking into JSON output.
func (r RedactedString) MarshalJSON() ([]byte, error) {
	return []byte(`"[REDACTED]"`), nil
}
