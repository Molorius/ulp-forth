/*
Copyright 2024 Blake Felt blake.w.felt@gmail.com

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/
package forth

// Flags associated with each word
type Flag struct {
	Hidden    bool // This word is hidden from being found.
	Immediate bool // This word should be executed immediately.
	Data      bool // This Forth word is data and should not be optimized.

	addedToList bool // This word is already added to the output list.
	recursive   bool // This Forth word is recursive.
	visited     bool // This Forth word has already been visited in this optimization pass.
}
