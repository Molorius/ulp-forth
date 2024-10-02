package forth

// Flags associated with each word
type Flag struct {
	Hidden    bool // This word is hidden from being found.
	Immediate bool // This word should be executed immediately.
}
