package forth

import (
	"fmt"
	"strings"
)

// A Dictionary entry. Contains the name, the word itself, and the flags.
type DictionaryEntry struct {
	Name string
	Word Word
	Flag Flag
}

func (d DictionaryEntry) String() string {
	return d.Name
}

// The Forth Dictionary. This architecture uses individual entries
// representing words rather than a flat cell structure.
type Dictionary struct {
	Entries []*DictionaryEntry
}

// Set up the empty dictionary.
func (d *Dictionary) Setup() error {
	d.Entries = make([]*DictionaryEntry, 0)
	return nil
}

func (d *Dictionary) AddEntry(entry *DictionaryEntry) error {
	if d.Entries == nil {
		return fmt.Errorf("Dictionary not set up when adding entry, please file a bug report.")
	}
	d.Entries = append(d.Entries, entry)
	return nil
}

func (d *Dictionary) FindName(name string) (*DictionaryEntry, error) {
	if d.Entries == nil {
		return nil, fmt.Errorf("Dictionary not set up when finding name, please file a bug report.")
	}
	nameLower := strings.ToLower(name) // case insensitive finding
	for i := len(d.Entries) - 1; i >= 0; i-- {
		entry := d.Entries[i]
		if !entry.Flag.Hidden && strings.ToLower(entry.Name) == nameLower {
			return entry, nil
		}
	}
	return nil, fmt.Errorf("%s not found in dictionary.", name)
}

func (d *Dictionary) LastForthWord() (*WordForth, error) {
	lastEntry := d.Entries[len(d.Entries)-1]
	last, ok := lastEntry.Word.(WordForth)
	if !ok {
		return nil, fmt.Errorf("The last word in dictionary is not a Forth word.")
	}
	return &last, nil
}
