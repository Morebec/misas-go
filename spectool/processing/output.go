package processing

// OutputFile represents a file the needs to be outputted.
type OutputFile struct {
	// Absolute path to the file
	Path string

	// Content that should be written.
	Data []byte
}
