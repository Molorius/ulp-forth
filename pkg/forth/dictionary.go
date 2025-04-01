/*
Copyright 2024-2025 Blake Felt blake.w.felt@gmail.com

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/
package forth

import (
	"fmt"
	"strings"
)

// A Dictionary entry. Contains the name, the word itself, and the flags.
type DictionaryEntry struct {
	Name    string
	ulpName string // the name we're going to compile this to
	Word    Word
	Flag    Flag
}

func (d DictionaryEntry) String() string {
	return d.Name
}

func (d DictionaryEntry) Details() string {
	return fmt.Sprintf("%s", d.Word)
}

func (d *DictionaryEntry) AddToList(u *Ulp) error {
	return d.Word.AddToList(u)
}

func (d *DictionaryEntry) ClearVisited() {
	d.Flag.visited = false
}

func (d *DictionaryEntry) BodyLabel() string {
	return "__body" + d.ulpName
}

// The Forth Dictionary. This architecture uses individual entries
// representing words rather than a flat cell structure.
type Dictionary struct {
	Entries  []*DictionaryEntry
	vm       *VirtualMachine
	entryMap map[string][]*DictionaryEntry
}

// Set up the empty dictionary.
func (d *Dictionary) Setup(vm *VirtualMachine) error {
	d.Entries = make([]*DictionaryEntry, 0)
	d.entryMap = make(map[string][]*DictionaryEntry)
	d.vm = vm
	return nil
}

func (d *Dictionary) AddEntry(entry *DictionaryEntry) error {
	if d.Entries == nil {
		return fmt.Errorf("dictionary not set up when adding entry, please file a bug report")
	}
	name := entry.Name
	if name != "" {
		previous, _ := d.FindName(name)
		if previous != nil {
			fmt.Fprintf(d.vm.Out, "Redefining %s ", name)
		}
	}
	lower := d.standardizeName(name)
	d.Entries = append(d.Entries, entry)
	lst, ok := d.entryMap[lower]
	if !ok {
		lst = make([]*DictionaryEntry, 0)
	}
	lst = append(lst, entry)
	d.entryMap[lower] = lst
	return nil
}

// standardizeName takes in the name of a word and
// returns the standardized version. Currently this is
// the lower case version for case-insensitivity but
// it could be changes to a case-sensitive meaning.
func (d *Dictionary) standardizeName(name string) string {
	return strings.ToLower(name)
}

func (d *Dictionary) FindName(name string) (*DictionaryEntry, error) {
	if d.Entries == nil {
		return nil, fmt.Errorf("dictionary not set up when finding name, please file a bug report")
	}
	nameLower := d.standardizeName(name)
	same, ok := d.entryMap[nameLower]
	if ok {
		for i := len(same) - 1; i >= 0; i-- {
			entry := same[i]
			if !entry.Flag.Hidden {
				return entry, nil
			}
		}
	}
	return nil, fmt.Errorf("%s not found in dictionary", name)
}

func (d *Dictionary) LastForthWord() (*WordForth, error) {
	lastEntry := d.Entries[len(d.Entries)-1]
	last, ok := lastEntry.Word.(*WordForth)
	if !ok {
		return nil, fmt.Errorf("the last word in dictionary is not a Forth word")
	}
	return last, nil
}
