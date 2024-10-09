package forth

// Flags associated with each word
type Flag struct {
	Hidden    bool // This word is hidden from being found.
	Immediate bool // This word should be executed immediately.
	Data      bool // This Forth word is data and should not be optimized.
}
