/*
Copyright 2024-2025 Blake Felt blake.w.felt@gmail.com

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/
package forth

import (
	"errors"
	"fmt"
)

type DictionaryEntryError struct {
	Entry  *DictionaryEntry
	Format string
	Args   []interface{}
}

func (e DictionaryEntryError) Error() string {
	end := fmt.Sprintf(e.Format, e.Args...)
	errString := fmt.Sprintf("%s %s", e.Entry, end)
	return errString
}

func EntryError(entry *DictionaryEntry, format string, a ...interface{}) error {
	return DictionaryEntryError{entry, format, a}
}

func JoinEntryError(err error, entry *DictionaryEntry, format string, a ...interface{}) error {
	entryError := DictionaryEntryError{entry, format, a}
	return errors.Join(entryError, err)
}

func PushError(err error, entry *DictionaryEntry) error {
	return JoinEntryError(err, entry, "could not push on stack")
}

func PopError(err error, entry *DictionaryEntry) error {
	return JoinEntryError(err, entry, "could not pop from stack")
}

// type JoinEntryError struct {
// 	Err    error
// 	Entry  *DictionaryEntry
// 	Format string
// 	Args   []interface{}
// }

// func (e JoinEntryError) Error() string {
// 	ee := EntryError{e.Entry, e.Format, e.Args}
// 	return errors.Join(ee, e.Err).Error()
// }
