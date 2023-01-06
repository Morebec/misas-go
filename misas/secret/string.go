package secret

type String string

const RedactedValue = "REDACTED"

// Implements the Stringer interface to return a redacted value.
func (s String) String() string {
	return RedactedValue
}

// Value Returns the actual value of the string
func (s String) Value() string {
	return string(s)
}
