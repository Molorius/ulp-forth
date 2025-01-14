/*
Copyright 2024 Blake Felt blake.w.felt@gmail.com

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/
package forth

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

type primitive struct {
	name      string
	goFunc    PrimitiveGo
	ulpAsm    PrimitiveUlp
	ulpAsmSrt PrimitiveUlpSrt
	flag      Flag
}

func primitiveAdd(vm *VirtualMachine, name string, goFunc PrimitiveGo, ulpAsm PrimitiveUlp, ulpAsmSrt PrimitiveUlpSrt, flag Flag) error {
	var entry DictionaryEntry
	entry = DictionaryEntry{
		Name: name,
		Word: &WordPrimitive{
			Go:     goFunc,
			Ulp:    ulpAsm,
			UlpSrt: ulpAsmSrt,
			Entry:  &entry,
		},
		Flag: flag,
	}
	err := vm.Dictionary.AddEntry(&entry)
	if err != nil {
		return JoinEntryError(err, &entry, "error adding primitive entry to dictionary, please file a bug report")
	}
	return nil
}

func PrimitiveSetup(vm *VirtualMachine) error {
	prims := []primitive{
		{
			name: ".S",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				fmt.Fprintf(vm.Out, "%s", vm.Stack)
				return nil
			},
		},
		{
			name: "WORDS",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				fmt.Fprintln(vm.Out)
				for i := len(vm.Dictionary.Entries) - 1; i >= 0; i-- {
					name := vm.Dictionary.Entries[i].Name
					if len(name) > 0 {
						fmt.Fprint(vm.Out, name, " ")
					}
				}
				return nil
			},
		},
		{
			name: "--SEE", // ( )
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				cell, err := vm.Stack.Pop()
				if err != nil {
					return JoinEntryError(err, entry, "could not pop name")
				}
				cellAddr, ok := cell.(CellAddress)
				if !ok {
					return EntryError(entry, "requires an address cell")
				}
				name, err := cellsToString(cellAddr.Entry.Word.(*WordForth).Cells)
				if err != nil {
					return JoinEntryError(err, entry, "could not parse name")
				}
				wordEntry, err := vm.Dictionary.FindName(string(name))
				if err != nil {
					return JoinEntryError(err, entry, "could not find word")
				}
				fmt.Fprint(vm.Out, "\r\n", wordEntry.Details())
				return nil
			},
		},
		{
			name: "WORD",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				n, err := vm.Stack.PopNumber()
				if err != nil {
					return JoinEntryError(err, entry, "could not pop delimiter")
				}
				str, err := vm.ParseArea.Word(byte(n))
				if err != nil {
					return JoinEntryError(err, entry, "could not parse string")
				}
				var de DictionaryEntry
				cells, err := bytesToCells(str, true)
				if err != nil {
					return JoinEntryError(err, entry, "could not convert bytes to cells")
				}
				w := WordForth{cells, &de}
				de = DictionaryEntry{
					Word: &w,
					Flag: Flag{Data: true},
				}
				c := CellAddress{
					Entry:     &de,
					Offset:    0,
					UpperByte: false,
				}
				err = vm.Stack.Push(c)
				if err != nil {
					return JoinEntryError(err, entry, "could not push address")
				}
				return nil
			},
		},
		{
			name: "--CREATE-FORTH",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				cell, err := vm.Stack.Pop()
				if err != nil {
					return JoinEntryError(err, entry, "could not pop name")
				}
				cellAddr, ok := cell.(CellAddress)
				if !ok {
					return EntryError(entry, "requires an address cell")
				}
				name, err := cellsToString(cellAddr.Entry.Word.(*WordForth).Cells)
				if err != nil {
					return JoinEntryError(err, entry, "could not parse name")
				}
				var newEntry DictionaryEntry
				newEntry = DictionaryEntry{
					Name: string(name),
					Word: &WordForth{
						Cells: make([]Cell, 0),
						Entry: &newEntry,
					},
				}
				err = vm.Dictionary.AddEntry(&newEntry)
				if err != nil {
					return JoinEntryError(err, entry, "could not add entry to dictionary")
				}
				return nil
			},
		},
		{
			name: "--CREATE-ASSEMBLY", // ( asm_n ... asm0 count name -- )
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				name, err := parseWord(vm, entry)
				if err != nil {
					return JoinEntryError(err, entry, "could not parse name")
				}
				asm, err := parseAssembly(vm, entry)
				if err != nil {
					return JoinEntryError(err, entry, "could not parse assembly")
				}
				var newEntry DictionaryEntry
				newEntry = DictionaryEntry{
					Name: name,
					Word: &WordPrimitive{
						Go:    notImplemented,
						Ulp:   asm,
						Entry: &newEntry,
					},
				}
				err = vm.Dictionary.AddEntry(&newEntry)
				if err != nil {
					return JoinEntryError(err, entry, "could not add entry to dictionary")
				}
				return nil
			},
		},
		{
			name: "--CREATE-ASSEMBLY-SRT",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				name, err := parseWord(vm, entry)
				if err != nil {
					return JoinEntryError(err, entry, "could not parse name")
				}
				asm, err := parseAssembly(vm, entry)
				if err != nil {
					return JoinEntryError(err, entry, "could not parse assembly")
				}
				var newEntry DictionaryEntry
				newEntry = DictionaryEntry{
					Name: name,
					Word: &WordPrimitive{
						Go: notImplemented,
						UlpSrt: PrimitiveUlpSrt{
							Asm:             asm,
							NonStandardNext: true,
						},
						Entry: &newEntry,
					},
				}
				err = vm.Dictionary.AddEntry(&newEntry)
				if err != nil {
					return JoinEntryError(err, entry, "could not add entry to dictionary")
				}
				return nil
			},
		},
		{
			name: "--CREATE-ASSEMBLY-BOTH",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				name, err := parseWord(vm, entry)
				if err != nil {
					return JoinEntryError(err, entry, "could not parse name")
				}
				asmSrt, err := parseAssembly(vm, entry)
				if err != nil {
					return JoinEntryError(err, entry, "could not parse assembly")
				}
				asm, err := parseAssembly(vm, entry)
				if err != nil {
					return JoinEntryError(err, entry, "could not parse subroutine threaded assembly")
				}
				var newEntry DictionaryEntry
				newEntry = DictionaryEntry{
					Name: name,
					Word: &WordPrimitive{
						Go:  notImplemented,
						Ulp: asm,
						UlpSrt: PrimitiveUlpSrt{
							Asm:             asmSrt,
							NonStandardNext: true,
						},
						Entry: &newEntry,
					},
				}
				err = vm.Dictionary.AddEntry(&newEntry)
				if err != nil {
					return JoinEntryError(err, entry, "could not add entry to dictionary")
				}
				return nil
			},
		},
		{
			name: "[",
			flag: Flag{Immediate: true},
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				err := vm.State.Set(uint16(StateInterpret))
				if err != nil {
					return JoinEntryError(err, entry, "could not set state to interpret mode")
				}
				return nil
			},
		},
		{
			name: "]",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				err := vm.State.Set(uint16(StateCompile))
				if err != nil {
					return JoinEntryError(err, entry, "could not set state to compile mode")
				}
				return nil
			},
		},
		{
			name: "BYE",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				err := vm.State.Set(uint16(StateExit))
				if err != nil {
					return JoinEntryError(err, entry, "could not set state to exit mode")
				}
				return nil
			},
		},
		{
			name: "LAST",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				last := vm.Dictionary.Entries[len(vm.Dictionary.Entries)-1]
				c := CellAddress{last, 0, false}
				err := vm.Stack.Push(c)
				if err != nil {
					return PushError(err, entry)
				}
				return nil
			},
		},
		{
			name: ">C",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				c, err := vm.Stack.Pop()
				if err != nil {
					return PopError(err, entry)
				}
				err = vm.ControlFlowStack.Push(c)
				if err != nil {
					return JoinEntryError(err, entry, "could not push on control flow stack")
				}
				return nil
			},
		},
		{
			name: "C>",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				c, err := vm.ControlFlowStack.Pop()
				if err != nil {
					return JoinEntryError(err, entry, "could not pop from control flow stack")
				}
				err = vm.Stack.Push(c)
				if err != nil {
					return PushError(err, entry)
				}
				return nil
			},
		},
		{
			name: ">DO",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				c, err := vm.Stack.Pop()
				if err != nil {
					return PopError(err, entry)
				}
				err = vm.DoStack.Push(c)
				if err != nil {
					return JoinEntryError(err, entry, "could not push on do stack")
				}
				return nil
			},
		},
		{
			name: "DO>",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				c, err := vm.DoStack.Pop()
				if err != nil {
					return JoinEntryError(err, entry, "could not pop from do stack")
				}
				err = vm.Stack.Push(c)
				if err != nil {
					return PushError(err, entry)
				}
				return nil
			},
		},
		{
			name: ">R",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				c, err := vm.Stack.Pop()
				if err != nil {
					return PopError(err, entry)
				}
				err = vm.ReturnStack.Push(c)
				if err != nil {
					return JoinEntryError(err, entry, "could not push on return stack")
				}
				return nil
			},
			ulpAsm: PrimitiveUlp{
				"ld r0, r3, 0",     // get value from stack
				"ld r1, r2, __rsp", // load rsp
				"add r1, r1, 1",    // increment rsp
				"st r0, r1, 0",     // store value on rsp
				"st r1, r2, __rsp", // save rsp
				"add r3, r3, 1",    // decrement stack
				"jump __next_skip_r2",
			},
			ulpAsmSrt: PrimitiveUlpSrt{
				Asm: []string{
					"move r0, __rsp", // set pointer to rsp location
					"ld r1, r0, 0",   // load rsp
					"add r1, r1, 1",  // increment rsp
					"st r1, r0, 0",   // store rsp
					"ld r0, r3, 0",   // load value from stack
					"st r0, r1, 0",   // store value on rsp
					"add r3, r3, 1",  // decrement stack
				},
			},
		},
		{
			name: "R>",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				c, err := vm.ReturnStack.Pop()
				if err != nil {
					return JoinEntryError(err, entry, "could not pop from return stack")
				}
				err = vm.Stack.Push(c)
				if err != nil {
					return PushError(err, entry)
				}
				return nil
			},
			ulpAsm: PrimitiveUlp{
				"ld r1, r2, __rsp", // load rsp
				"ld r0, r1, 0",     // get value from rsp
				"sub r1, r1, 1",    // decrement rsp
				"st r1, r2, __rsp", // store rsp
				"sub r3, r3, 1",    // increment stack
				"st r0, r3, 0",     // store value on stack
				"jump __next_skip_r2",
			},
			ulpAsmSrt: PrimitiveUlpSrt{
				Asm: []string{
					"move r0, __rsp", // set pointer to rsp location
					"ld r1, r0, 0",   // load rsp
					"sub r1, r1, 1",  // decrement rsp
					"st r1, r0, 0",   // store rsp
					"ld r0, r1, 1",   // load value from rsp
					"sub r3, r3, 1",  // increment stack
					"st r0, r3, 0",   // store value on stack
				},
			},
		},
		{
			name: "BRANCH",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				err := vm.Stack.Push(&CellBranch{})
				if err != nil {
					return PushError(err, entry)
				}
				return nil
			},
		},
		{
			name: "BRANCH0",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				err := vm.Stack.Push(&CellBranch0{})
				if err != nil {
					return PushError(err, entry)
				}
				return nil
			},
		},
		{
			name: "DEST",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				err := vm.Stack.Push(&CellDestination{})
				if err != nil {
					return PushError(err, entry)
				}
				return nil
			},
		},
		{
			name: "RESOLVE-BRANCH",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				destCell, err := vm.Stack.Pop()
				if err != nil {
					return JoinEntryError(err, entry, "could not pop destination")
				}
				branchCell, err := vm.Stack.Pop()
				if err != nil {
					return JoinEntryError(err, entry, "could not pop branch")
				}
				dest, ok := destCell.(*CellDestination)
				if !ok {
					return EntryError(entry, "expected a destination found %s type %T", destCell, destCell)
				}
				switch b := branchCell.(type) {
				case *CellBranch:
					b.dest = dest
				case *CellBranch0:
					b.dest = dest
				default:
					return EntryError(entry, "expected a branch found %s type %T", branchCell, branchCell)
				}
				return nil
			},
		},
		{
			name: "LITERAL",
			flag: Flag{Immediate: true},
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				c, err := vm.Stack.Pop()
				if err != nil {
					return PopError(err, entry)
				}
				newCell := CellLiteral{c}
				last, err := vm.Dictionary.LastForthWord()
				if err != nil {
					return JoinEntryError(err, entry, "could not get last forth word")
				}
				last.Cells = append(last.Cells, newCell)
				return nil
			},
		},
		{
			name: "FIND-WORD",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				cell, err := vm.Stack.Pop()
				if err != nil {
					return PopError(err, entry)
				}
				cellAddr, ok := cell.(CellAddress)
				if !ok {
					return EntryError(entry, "requires an address cell")
				}
				name, err := cellsToString(cellAddr.Entry.Word.(*WordForth).Cells)
				if err != nil {
					return JoinEntryError(err, entry, "could not convert input to string")
				}
				found, err := vm.Dictionary.FindName(string(name))
				if err != nil {
					return JoinEntryError(err, entry, "could not find name: %s", name)
				}
				newCell := CellAddress{found, 0, false}
				err = vm.Stack.Push(newCell)
				if err != nil {
					return PushError(err, entry)
				}
				return nil
			},
		},
		{
			name: "COMPILE,",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				c, err := vm.Stack.Pop()
				if err != nil {
					return PopError(err, entry)
				}
				last, err := vm.Dictionary.LastForthWord()
				if err != nil {
					return JoinEntryError(err, entry, "could not get last forth word")
				}
				last.Cells = append(last.Cells, c)
				dest, ok := c.(*CellDestination)
				if ok {
					dest.Addr = CellAddress{
						Entry:  last.Entry,
						Offset: len(last.Cells) - 1,
					}
				}
				return nil
			},
		},
		{
			name: "EXECUTE",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				c, err := vm.Stack.Pop()
				if err != nil {
					return PopError(err, entry)
				}
				cellAddr, ok := c.(CellAddress)
				if !ok {
					return EntryError(entry, "unable to execute cell %s", c)
				}
				err = cellAddr.Execute(vm)
				if err != nil {
					return JoinEntryError(err, entry, "error while executing")
				}
				return nil
			},
			ulpAsm: PrimitiveUlp{
				"ld r0, r3, 0",   // load the token into r0
				"add r3, r3, 1",  // decrement stack pointer
				"jump __ins_asm", // start execution of the token
			},
			ulpAsmSrt: PrimitiveUlpSrt{
				Asm: []string{
					// Each assembly word is just the address,
					// each forth word is the address plus 0x8000,
					// which denotes that it is a forth word
					// with the 0x8000.
					"ld r0, r3, 0",                         // load the token into r0
					"add r3, r3, 1",                        // decrement stack pointer
					"jumpr __execute.0, __forth_words, ge", // jump if the address is past assembly words
					// it's an assembly word, execute it
					"jump r0",
					"__execute.0:",
					// it's a forth word, put r2 on return stack
					"move r1, __rsp", // put pointer on rsp
					"ld r1, r1, 0",   // load rsp
					"add r1, r1, 1",  // increment rsp
					"st r2, r1, 0",   // store return address
					"move r2, __rsp", // put pointer on rsp
					"st r1, r2, 0",   // store rsp
					"add r2, r0, 1",  // go past the docol
					"jump r2",        // jump to the forth word, past the docol
				},
				NonStandardNext: true,
			},
		},
		{
			name: "ALLOCATE",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				n, err := vm.Stack.PopNumber()
				if err != nil {
					return PopError(err, entry)
				}
				var de DictionaryEntry // note that we don't put this entry into the dictionary
				w := WordForth{
					Cells: make([]Cell, n),
					Entry: &de,
				}
				for i := range w.Cells {
					w.Cells[i] = CellNumber{0}
				}
				de = DictionaryEntry{
					Word: &w,
					Flag: Flag{Data: true},
				}
				cell := CellAddress{
					Offset: 0,
					Entry:  &de,
				}
				err = vm.Stack.Push(cell)
				if err != nil {
					return PushError(err, entry)
				}
				err = vm.Stack.Push(CellNumber{0})
				if err != nil {
					return PushError(err, entry)
				}
				return nil
			},
		},
		{
			name: "@",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				cell, err := vm.Stack.Pop()
				if err != nil {
					return PopError(err, entry)
				}
				switch c := cell.(type) {
				case CellAddress:
					w, ok := c.Entry.Word.(*WordForth)
					if !ok {
						return EntryError(entry, "can only read forth data words found %s type %T", w, w)
					}
					if c.Offset < 0 || c.Offset >= len(w.Cells) {
						return EntryError(entry, "reading outside of data range, offset %d", c.Offset)
					}
					err = vm.Stack.Push(w.Cells[c.Offset])
					if err != nil {
						return PushError(err, entry)
					}
					return nil
				default:
					return EntryError(entry, "can only read address cells found %s type %T", cell, cell)
				}
			},
			ulpAsm: PrimitiveUlp{
				"ld r0, r3, 0", // get the address from stack
				"ld r0, r0, 0", // load the value
				"st r0, r3, 0", // store the value on stack
				"jump __next_skip_load",
			},
			ulpAsmSrt: PrimitiveUlpSrt{
				Asm: []string{
					"ld r0, r3, 0", // get the address from stack
					"ld r0, r0, 0", // load the value
					"st r0, r3, 0", // store the value on stack
				},
			},
		},
		{
			name: "!",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				cell, err := vm.Stack.Pop()
				if err != nil {
					return PopError(err, entry)
				}
				n, err := vm.Stack.Pop()
				if err != nil {
					return PopError(err, entry)
				}
				switch c := cell.(type) {
				case CellAddress:
					w, ok := c.Entry.Word.(*WordForth)
					if !ok {
						return EntryError(entry, "can only write forth data words found %s type %T", w, w)
					}
					if c.Offset < 0 || c.Offset >= len(w.Cells) {
						return EntryError(entry, "writing outside of data range, offset %d", c.Offset)
					}
					w.Cells[c.Offset] = n
					return nil
				default:
					return EntryError(entry, "can only write address cells found %s type %T", cell, cell)
				}
			},
			ulpAsm: PrimitiveUlp{
				"ld r0, r3, 0",  // load the address
				"ld r1, r3, 1",  // load the value
				"st r1, r0, 0",  // store the value in address
				"add r3, r3, 2", // decrement the stack
				"jump __next_skip_r2",
			},
			ulpAsmSrt: PrimitiveUlpSrt{
				Asm: []string{
					"ld r0, r3, 0",  // load the address
					"ld r1, r3, 1",  // load the value
					"st r1, r0, 0",  // store the value in address
					"add r3, r3, 2", // decrement the stack
				},
			},
		},
		{
			name: ">BODY",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				cell, err := vm.Stack.Pop()
				if err != nil {
					return PopError(err, entry)
				}
				c, ok := cell.(CellAddress)
				if !ok {
					return EntryError(entry, "can only find body of an address")
				}
				w, ok := c.Entry.Word.(*WordForth)
				if !ok {
					return EntryError(entry, "requires a forth word")
				}
				if len(w.Cells) < 1 {
					return EntryError(entry, "reading a word without enough memory")
				}
				litCell, ok := w.Cells[0].(CellLiteral)
				if !ok {
					return EntryError(entry, "did not find a cell literal, was the address defined by DEFER ?")
				}
				addrCell, ok := litCell.cell.(CellAddress)
				if !ok {
					return EntryError(entry, "did not find an address, was this address defined by DEFER ?")
				}
				err = vm.Stack.Push(addrCell)
				if err != nil {
					return PushError(err, entry)
				}
				return nil
			},
			ulpAsm: PrimitiveUlp{
				"ld r0, r3, 0", // get the address from stack
				"ld r0, r0, 0", // the body address token is in the front, load it
				"ld r0, r0, 0", // get the address from the token
				"st r0, r3, 0", // store the body address on stack
				"jump __next_skip_load",
			},
			ulpAsmSrt: PrimitiveUlpSrt{
				Asm: []string{
					"ld r0, r3, 0",  // get the address from stack
					"ld r0, r0, 1",  // get the body address in the move instruction
					"rsh r0, r0, 4", // shift it into the correct space
					"st r0, r3, 0",  // store the body address on stack
				},
			},
		},
		{
			name: "C@",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				cell, err := vm.Stack.Pop()
				if err != nil {
					return err
				}
				switch c := cell.(type) {
				case CellAddress:
					w, ok := c.Entry.Word.(*WordForth)
					if !ok {
						return EntryError(entry, "can only read forth data words found %s type %T", w, w)
					}
					if c.Offset < 0 || c.Offset >= len(w.Cells) {
						return EntryError(entry, "reading outside of data range, offset %d", c.Offset)
					}
					readCell := w.Cells[c.Offset]
					numCell, ok := readCell.(CellNumber)
					if !ok {
						return EntryError(entry, "can only read a number found %s type %T", readCell, readCell)
					}
					n := numCell.Number
					if c.UpperByte {
						n = n >> 8
					}
					n = n & 0xFF
					err = vm.Stack.Push(CellNumber{n})
					if err != nil {
						return PushError(err, entry)
					}
					return nil
				default:
					return EntryError(entry, "can only read address cells found %s type %T", cell, cell)
				}
			},
			ulpAsm: PrimitiveUlp{
				"ld r0, r3, 0",                      // load the address
				"ld r1, r0, 0",                      // load the full value
				"jumpr __c_ampersand.0, 0x8000, lt", // jump if the "upper" bit is not set
				"rsh r1, r1, 8",                     // "upper" bit set, shift it down
				"__c_ampersand.0:",
				"and r1, r1, 0xFF", // mask off the upper bits
				"st r1, r3, 0",     // store the masked value
				"jump __next_skip_r2",
			},
			ulpAsmSrt: PrimitiveUlpSrt{
				Asm: []string{
					"ld r0, r3, 0",                      // load the address
					"ld r1, r0, 0",                      // load the full value
					"jumpr __c_ampersand.0, 0x8000, lt", // jump if the "upper" bit is not set
					"rsh r1, r1, 8",                     // "upper" bit set, shift it down
					"__c_ampersand.0:",
					"and r1, r1, 0xFF", // mask off the upper bits
					"st r1, r3, 0",     // store the masked value
				},
			},
		},
		{
			name: "C!",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				cell, err := vm.Stack.Pop()
				if err != nil {
					return PopError(err, entry)
				}
				n, err := vm.Stack.PopNumber()
				if err != nil {
					return PopError(err, entry)
				}
				n = n & 0xFF // mask off the upper bits
				switch c := cell.(type) {
				case CellAddress:
					w, ok := c.Entry.Word.(*WordForth)
					if !ok {
						return EntryError(entry, "can only write forth data words found %s type %T", w, w)
					}
					if c.Offset < 0 || c.Offset >= len(w.Cells) {
						return EntryError(entry, "writing outside of data range, offset %d", c.Offset)
					}
					readCell := w.Cells[c.Offset]
					numCell, ok := readCell.(CellNumber)
					if !ok {
						return EntryError(entry, "can only write a number found %s type %T", readCell, readCell)
					}
					storedN := numCell.Number
					if c.UpperByte {
						storedN = storedN & 0x00FF // mask off the upper bits
						storedN = storedN | (n << 8)
					} else {
						storedN = storedN & 0xFF00 // mask off the lower bits
						storedN = storedN | n
					}
					w.Cells[c.Offset] = CellNumber{storedN}
					return nil
				default:
					return EntryError(entry, "can only write address cells found %s type %T", cell, cell)
				}
			},
			ulpAsm: PrimitiveUlp{
				"ld r0, r3, 0",                        // load the address
				"ld r1, r3, 1",                        // load the value
				"ld r2, r0, 0",                        // load the old value
				"jumpr __c_exclamation.0, 0x8000, lt", // jump if the upper bit is not set
				// upper bit set, store the upper value
				"and r2, r2, 0x00FF", // mask off the upper bits of old value
				"lsh r1, r1, 8",      // shift over the value
				"jump __c_exclamation.1",
				"__c_exclamation.0:",
				// store the lower value
				"and r2, r2, 0xFF00", // mask off the lower bits of old value
				"and r1, r1, 0x00FF", // mask off the value
				"__c_exclamation.1:",
				"or r2, r2, r1", // merge the new value and old value
				"st r2, r0, 0",  // store into the address
				"add r3, r3, 2", // decrement stack
				"jump next",
			},
			ulpAsmSrt: PrimitiveUlpSrt{
				Asm: []string{
					"st r2, r3, -1",                       // store r2 on stack
					"ld r0, r3, 0",                        // load the address
					"ld r1, r3, 1",                        // load the value
					"ld r2, r0, 0",                        // load the old value
					"jumpr __c_exclamation.0, 0x8000, lt", // jump if the upper bit is not set
					// upper bit set, store the upper value
					"and r2, r2, 0x00FF", // mask off the upper bits of old value
					"lsh r1, r1, 8",      // shift over the value
					"jump __c_exclamation.1",
					"__c_exclamation.0:",
					// store the lower value
					"and r2, r2, 0xFF00", // mask off the lower bits of old value
					"and r1, r1, 0x00FF", // mask off the value
					"__c_exclamation.1:",
					"or r2, r2, r1", // merge the value and old value
					"st r2, r0, 0",  // store into the address
					"ld r2, r3, -1", // restore r2
					"add r3, r3, 2", // decrement stack
				},
			},
		},
		{
			name: "CHAR+",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				cell, err := vm.Stack.Pop()
				if err != nil {
					return PopError(err, entry)
				}
				switch c := cell.(type) {
				case CellAddress:
					n := c.Offset
					var upper bool
					if c.UpperByte {
						upper = false
						n += 1
					} else {
						upper = true
					}
					newCell := CellAddress{
						Entry:     c.Entry,
						Offset:    n,
						UpperByte: upper,
					}
					err = vm.Stack.Push(newCell)
					if err != nil {
						return PushError(err, entry)
					}
					return nil
				default:
					return EntryError(entry, "cannot add %s type %T", cell, cell)
				}
			},
			ulpAsm: PrimitiveUlp{
				"ld r0, r3, 0",
				"jumpr __char_plus.0, 0x8000, lt", // jump if the "upper" bit is not set
				// bit is set!
				"and r0, r0, 0x7FFF", // remove that upper bit
				"add r0, r0, 1",      // increment to next position
				"jump __char_plus.1",
				"__char_plus.0:",
				// bit is not set
				"or r0, r0, 0x8000", // set the bit
				"__char_plus.1:",
				"st r0, r3, 0", // store the result
				"jump __next_skip_r2",
			},
			ulpAsmSrt: PrimitiveUlpSrt{
				Asm: []string{
					"ld r0, r3, 0",
					"jumpr __char_plus.0, 0x8000, lt", // jump if the "upper" bit is not set
					// bit is set!
					"and r0, r0, 0x7FFF", // remove that upper bit
					"add r0, r0, 1",      // increment to next position
					"jump __char_plus.1",
					"__char_plus.0:",
					// bit is not set
					"or r0, r0, 0x8000", // set the bit
					"__char_plus.1:",
					"st r0, r3, 0", // store the result
				},
			},
		},
		{
			name: "ALIGNED",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				cell, err := vm.Stack.Pop()
				if err != nil {
					return PopError(err, entry)
				}
				switch c := cell.(type) {
				case CellAddress:
					offset := c.Offset
					if c.UpperByte {
						offset++
					}
					newCell := CellAddress{
						Entry:     c.Entry,
						Offset:    offset,
						UpperByte: false,
					}
					err = vm.Stack.Push(newCell)
					if err != nil {
						return PushError(err, entry)
					}
					return nil
				default:
					return EntryError(entry, "cannot align %s type %T", cell, cell)
				}
			},
			ulpAsm: PrimitiveUlp{
				"ld r0, r3, 0", // load the address
				"jumpr __aligned.0, 0x8000, lt",
				"add r0, r0, 1",      // go to the next major position
				"and r0, r0, 0x7FFF", // mask off the upper bit
				"st r0, r3, 0",       // store the result
				"__aligned.0:",
				"jump __next_skip_load",
			},
			ulpAsmSrt: PrimitiveUlpSrt{
				Asm: []string{
					"ld r0, r3, 0", // load the address
					"jumpr __aligned.0, 0x8000, lt",
					"add r0, r0, 1",      // go to the next major position
					"and r0, r0, 0x7FFF", // mask off the upper bit
					"st r0, r3, 0",       // store the result
					"__aligned.0:",
				},
			},
		},
		{
			name: "--POSTPONE",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				cell, err := vm.Stack.Pop()
				if err != nil {
					return PopError(err, entry)
				}
				cellAddr, ok := cell.(CellAddress)
				if !ok {
					return EntryError(entry, "requires an address cell, found %s type %T", cell, cell)
				}
				last, err := vm.Dictionary.LastForthWord()
				if err != nil {
					return JoinEntryError(err, entry, "could not get last forth word")
				}
				if cellAddr.Entry.Flag.Immediate {
					last.Cells = append(last.Cells, cellAddr)
					return nil
				} else {
					compile, err := vm.Dictionary.FindName("COMPILE,")
					if !ok {
						JoinEntryError(err, entry, "requires the word COMPILE,")
					}
					newCells := []Cell{CellLiteral{cellAddr}, CellAddress{compile, 0, false}}
					last.Cells = append(last.Cells, newCells...)
				}
				return nil
			},
		},
		{
			name: "SET-HIDDEN",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				cell0, err := vm.Stack.Pop()
				if err != nil {
					return PopError(err, entry)
				}
				cellAddr, ok := cell0.(CellAddress)
				if !ok {
					return EntryError(entry, "requires an address cell, found %s type %T", cell0, cell0)
				}
				cellNum, err := vm.Stack.PopNumber()
				if err != nil {
					return JoinEntryError(err, entry, "could not get boolean")
				}
				flag := cellNum != 0
				cellAddr.Entry.Flag.Hidden = flag
				return nil
			},
		},
		{
			name: "SET-IMMEDIATE",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				cell0, err := vm.Stack.Pop()
				if err != nil {
					return PopError(err, entry)
				}
				cellAddr, ok := cell0.(CellAddress)
				if !ok {
					return EntryError(entry, "requires an address cell, found %s type %T", cell0, cell0)
				}
				cellNum, err := vm.Stack.PopNumber()
				if err != nil {
					return JoinEntryError(err, entry, "could not get boolean")
				}
				flag := cellNum != 0
				cellAddr.Entry.Flag.Immediate = flag
				return nil
			},
		},
		{
			name: "EXIT",
			flag: Flag{
				isExit: true,
			},
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				ret, err := vm.ReturnStack.Pop()
				if err != nil {
					return PopError(err, entry)
				}
				addr, ok := ret.(*CellAddress)
				if !ok {
					return EntryError(entry, "requires a return address, found %s type %T", ret, ret)
				}
				vm.IP = addr
				return nil
			},
			ulpAsm: PrimitiveUlp{
				"ld r0, r2, __rsp",      // load the return stack pointer
				"ld r1, r0, 0",          // load the return address into r1
				"sub r0, r0, 1",         // decrement pointer
				"st r0, r2, __rsp",      // store the updated return stack pointer
				"jump __next_skip_load", // skip loading, r1 and r2 are already fine
			},
			ulpAsmSrt: PrimitiveUlpSrt{
				Asm: []string{
					"move r0, __rsp", // point to the return stack pointer
					"ld r1, r0, 0",   // load the return stack pointer
					"ld r2, r1, 0",   // load the return address
					"sub r1, r1, 1",  // decrement rsp
					"st r1, r0, 0",   // store the rsp
					// we will continue executing at the return address
				},
			},
		},
		{
			name: "+",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				right, err := vm.Stack.Pop()
				if err != nil {
					return PopError(err, entry)
				}
				left, err := vm.Stack.Pop()
				if err != nil {
					return PopError(err, entry)
				}
				switch r := right.(type) {
				case CellNumber:
					switch l := left.(type) {
					case CellNumber:
						err = vm.Stack.Push(CellNumber{l.Number + r.Number})
						if err != nil {
							return PushError(err, entry)
						}
						return nil
					case CellAddress:
						err = vm.Stack.Push(CellAddress{l.Entry, l.Offset + int(r.Number), false})
						if err != nil {
							return PushError(err, entry)
						}
						return nil
					default:
						return EntryError(entry, "could not add %s type %T and %s type %T due to types", left, left, right, right)
					}
				case CellAddress:
					switch l := left.(type) {
					case CellNumber:
						err = vm.Stack.Push(CellAddress{r.Entry, int(l.Number) + r.Offset, false})
						if err != nil {
							return PushError(err, entry)
						}
						return nil
					default:
						return EntryError(entry, "could not add %s type %T and %s type %T due to types", left, left, right, right)
					}
				default:
					return EntryError(entry, "could not add %s type %T and %s type %T due to types", left, left, right, right)
				}
			},
			ulpAsm: PrimitiveUlp{
				"ld r0, r3, 0",
				"ld r1, r3, 1",
				"add r0, r0, r1",
				"add r3, r3, 1",
				"st r0, r3, 0",
				"jump __next_skip_r2",
			},
			ulpAsmSrt: PrimitiveUlpSrt{
				Asm: []string{
					"ld r0, r3, 0",
					"ld r1, r3, 1",
					"add r0, r0, r1",
					"add r3, r3, 1",
					"st r0, r3, 0",
				},
			},
		},
		{
			name: "-",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				right, err := vm.Stack.Pop()
				if err != nil {
					return PopError(err, entry)
				}
				left, err := vm.Stack.Pop()
				if err != nil {
					return PopError(err, entry)
				}
				switch r := right.(type) {
				case CellNumber:
					switch l := left.(type) {
					case CellNumber:
						err = vm.Stack.Push(CellNumber{l.Number - r.Number})
						if err != nil {
							return PushError(err, entry)
						}
						return nil
					case CellAddress:
						err = vm.Stack.Push(CellAddress{l.Entry, l.Offset - int(r.Number), false})
						if err != nil {
							return PushError(err, entry)
						}
						return nil
					default:
						return EntryError(entry, "could not subtract %s type %T from %s type %T due to types", right, right, left, left)
					}
				case CellAddress:
					switch l := left.(type) {
					case CellAddress:
						if l.Entry == r.Entry {
							err = vm.Stack.Push(CellNumber{uint16(l.Offset) - uint16(r.Offset)})
							if err != nil {
								return PushError(err, entry)
							}
							return nil
						}
						return EntryError(entry, "could not subtract %s type %T from %s type %T due to types", right, right, left, left)
					default:
						return EntryError(entry, "could not subtract %s type %T from %s type %T due to types", right, right, left, left)
					}
				default:
					return EntryError(entry, "could not subtract %s type %T from %s type %T due to types", right, right, left, left)
				}
			},
			ulpAsm: PrimitiveUlp{
				"ld r0, r3, 1",
				"ld r1, r3, 0",
				"sub r0, r0, r1",
				"add r3, r3, 1",
				"st r0, r3, 0",
				"jump __next_skip_r2",
			},
			ulpAsmSrt: PrimitiveUlpSrt{
				Asm: []string{
					"ld r0, r3, 1",
					"ld r1, r3, 0",
					"sub r0, r0, r1",
					"add r3, r3, 1",
					"st r0, r3, 0",
				},
			},
		},
		{
			name: "AND",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				right, err := vm.Stack.PopNumber()
				if err != nil {
					return PopError(err, entry)
				}
				left, err := vm.Stack.PopNumber()
				if err != nil {
					return PopError(err, entry)
				}
				err = vm.Stack.Push(CellNumber{left & right})
				if err != nil {
					return PushError(err, entry)
				}
				return nil
			},
			ulpAsm: PrimitiveUlp{
				"ld r0, r3, 1",
				"ld r1, r3, 0",
				"and r0, r0, r1",
				"add r3, r3, 1",
				"st r0, r3, 0",
				"jump __next_skip_r2",
			},
			ulpAsmSrt: PrimitiveUlpSrt{
				Asm: []string{
					"ld r0, r3, 1",
					"ld r1, r3, 0",
					"and r0, r0, r1",
					"add r3, r3, 1",
					"st r0, r3, 0",
				},
			},
		},
		{
			name: "OR",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				right, err := vm.Stack.PopNumber()
				if err != nil {
					return PopError(err, entry)
				}
				left, err := vm.Stack.PopNumber()
				if err != nil {
					return PopError(err, entry)
				}
				err = vm.Stack.Push(CellNumber{left | right})
				if err != nil {
					return PushError(err, entry)
				}
				return nil
			},
			ulpAsm: PrimitiveUlp{
				"ld r0, r3, 1",
				"ld r1, r3, 0",
				"or r0, r0, r1",
				"add r3, r3, 1",
				"st r0, r3, 0",
				"jump __next_skip_r2",
			},
			ulpAsmSrt: PrimitiveUlpSrt{
				Asm: []string{
					"ld r0, r3, 1",
					"ld r1, r3, 0",
					"or r0, r0, r1",
					"add r3, r3, 1",
					"st r0, r3, 0",
				},
			},
		},
		{
			name: "*",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				right, err := vm.Stack.PopNumber()
				if err != nil {
					return PopError(err, entry)
				}
				left, err := vm.Stack.PopNumber()
				if err != nil {
					return PopError(err, entry)
				}
				err = vm.Stack.Push(CellNumber{left * right})
				if err != nil {
					return PushError(err, entry)
				}
				return nil
			},
			ulpAsm: PrimitiveUlp{
				// x * y = z
				// x on r1
				// y on r0
				// z on 1, 0 after stack decrement
				"ld r1, r3, 1", // load x
				"ld r0, r3, 0", // load y
				"st r2, r3, 1", // initialize z to 0
				"__mult.0:",
				"and r2, r0, 1",     // get the lowest bit of y
				"jump __mult.1, eq", // check if the bit is set
				// bit is set! z = z + x
				"ld r2, r3, 1",          // load z
				"add r2, r2, r1",        // z = z+x
				"st r2, r3, 1",          // store z
				"__mult.1:",             // then
				"lsh r1, r1, 1",         // x = x<<1
				"rsh r0, r0, 1",         // y = y>>1
				"jumpr __mult.0, 0, gt", // loop if y != 0
				// finalize
				"add r3, r3, 1", // decrement stack, z already in place
				"jump next",
			},
			ulpAsmSrt: PrimitiveUlpSrt{
				Asm: []string{
					// x * y = z
					// x on r1
					// y on r0
					// z on 1, 0 after stack decrement
					"st r2, r3, -1", // store r2
					"ld r1, r3, 1",  // load x
					"ld r0, r3, 0",  // load y
					"move r2, 0",    // r2 = 0
					"st r2, r3, 1",  // initialize z to 0
					"__mult.0:",
					"and r2, r0, 1",     // get the lowest bit of y
					"jump __mult.1, eq", // check if the bit is set
					// bit is set! z = z + x
					"ld r2, r3, 1",          // load z
					"add r2, r2, r1",        // z = z+x
					"st r2, r3, 1",          // store z
					"__mult.1:",             // then
					"lsh r1, r1, 1",         // x = x<<1
					"rsh r0, r0, 1",         // y = y>>1
					"jumpr __mult.0, 0, gt", // loop if y != 0
					// finalize
					"ld r2, r3, -1", // reload r2
					"add r3, r3, 1", // decrement stack, z already in place
				},
			},
		},
		{
			name: "U/MOD",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				right, err := vm.Stack.PopNumber()
				if err != nil {
					return PopError(err, entry)
				}
				left, err := vm.Stack.PopNumber()
				if err != nil {
					return PopError(err, entry)
				}
				quotient := left / right
				remainder := left % right
				err = vm.Stack.Push(CellNumber{remainder})
				if err != nil {
					return PushError(err, entry)
				}
				err = vm.Stack.Push(CellNumber{quotient})
				if err != nil {
					return PushError(err, entry)
				}
				return nil
			},
			ulpAsm: PrimitiveUlp{
				// 'd' on 0
				// 'n' on 1
				// 'q' on r1
				// 'r' on r2 (already set to 0)
				// loop on stage_cnt

				"stage_rst", // stage_cnt = 0

				"__divmod.0:",
				// shift n into r, shift r, shift q
				"lsh r2, r2, 1",                // r = r<<1
				"lsh r1, r1, 1",                // q = q<<1
				"ld r0, r3, 1",                 // load n
				"jumpr __divmod.1, 0x8000, lt", // jump if the top bit is not set
				"or r2, r2, 1",                 // if bit is set, "shift" this bit into r
				"__divmod.1:",                  // then
				"lsh r0, r0, 1",                // n = n<<1
				"st r0, r3, 1",                 // store n
				// attempt subtracting
				"ld r0, r3, 0",        // load d
				"sub r0, r2, r0",      // r0 = r - d
				"jump __divmod.2, ov", // jump ahead if that overflowed
				// no overflow
				"move r2, r0",  // store result into r
				"or r1, r1, 1", // set the lowest bit of q
				"__divmod.2:",
				"stage_inc 1",              // increase the stage counter
				"jumps __divmod.0, 16, lt", // loop over each bit

				// done! store r and q
				"st r2, r3, 1", // r
				"st r1, r3, 0", // q
				"jump next",
			},
			ulpAsmSrt: PrimitiveUlpSrt{
				Asm: []string{
					// 'd' on 0
					// 'n' on 1
					// 'q' on r1
					// 'r' on r2
					// loop on stage_cnt

					"st r2, r3, -1", // store r2
					"move r2, 0",    // initialize r2 to 0
					"stage_rst",     // stage_cnt = 0

					"__divmod.0:",
					// shift n into r, shift r, shift q
					"lsh r2, r2, 1",                // r = r<<1
					"lsh r1, r1, 1",                // q = q<<1
					"ld r0, r3, 1",                 // load n
					"jumpr __divmod.1, 0x8000, lt", // jump if the top bit is not set
					"or r2, r2, 1",                 // if bit is set, "shift" this bit into r
					"__divmod.1:",                  // then
					"lsh r0, r0, 1",                // n = n<<1
					"st r0, r3, 1",                 // store n
					// attempt subtracting
					"ld r0, r3, 0",        // load d
					"sub r0, r2, r0",      // r0 = r - d
					"jump __divmod.2, ov", // jump ahead if that overflowed
					// no overflow
					"move r2, r0",  // store result into r
					"or r1, r1, 1", // set the lowest bit of q
					"__divmod.2:",
					"stage_inc 1",              // increase the stage counter
					"jumps __divmod.0, 16, lt", // loop over each bit

					// done! store r and q
					"st r2, r3, 1",  // r
					"st r1, r3, 0",  // q
					"ld r2, r3, -1", // reload r2
				},
			},
		},
		{
			name: "LSHIFT",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				amount, err := vm.Stack.PopNumber()
				if err != nil {
					return PopError(err, entry)
				}
				num, err := vm.Stack.PopNumber()
				if err != nil {
					return PopError(err, entry)
				}
				err = vm.Stack.Push(CellNumber{num << amount})
				if err != nil {
					return PushError(err, entry)
				}
				return nil
			},
			ulpAsm: PrimitiveUlp{
				"ld r0, r3, 1",
				"ld r1, r3, 0",
				"lsh r0, r0, r1",
				"add r3, r3, 1",
				"st r0, r3, 0",
				"jump __next_skip_r2",
			},
			ulpAsmSrt: PrimitiveUlpSrt{
				Asm: []string{
					"ld r0, r3, 1",
					"ld r1, r3, 0",
					"lsh r0, r0, r1",
					"add r3, r3, 1",
					"st r0, r3, 0",
				},
			},
		},
		{
			name: "RSHIFT",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				amount, err := vm.Stack.PopNumber()
				if err != nil {
					return PopError(err, entry)
				}
				num, err := vm.Stack.PopNumber()
				if err != nil {
					return PopError(err, entry)
				}
				err = vm.Stack.Push(CellNumber{num >> amount})
				if err != nil {
					return PushError(err, entry)
				}
				return nil
			},
			ulpAsm: PrimitiveUlp{
				"ld r0, r3, 1",
				"ld r1, r3, 0",
				"rsh r0, r0, r1",
				"add r3, r3, 1",
				"st r0, r3, 0",
				"jump __next_skip_r2",
			},
			ulpAsmSrt: PrimitiveUlpSrt{
				Asm: []string{
					"ld r0, r3, 1",
					"ld r1, r3, 0",
					"rsh r0, r0, r1",
					"add r3, r3, 1",
					"st r0, r3, 0",
				},
			},
		},
		{
			name: "SWAP",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				right, err := vm.Stack.Pop()
				if err != nil {
					return PopError(err, entry)
				}
				left, err := vm.Stack.Pop()
				if err != nil {
					return PopError(err, entry)
				}
				err = vm.Stack.Push(right)
				if err != nil {
					return PushError(err, entry)
				}
				err = vm.Stack.Push(left)
				if err != nil {
					return PushError(err, entry)
				}
				return nil
			},
			ulpAsm: PrimitiveUlp{
				"ld r1, r3, 0",
				"ld r0, r3, 1",
				"st r1, r3, 1",
				"st r0, r3, 0",
				"jump __next_skip_r2",
			},
			ulpAsmSrt: PrimitiveUlpSrt{
				Asm: []string{
					"ld r1, r3, 0",
					"ld r0, r3, 1",
					"st r1, r3, 1",
					"st r0, r3, 0",
				},
			},
		},
		{
			name: "DUP",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				c, err := vm.Stack.Pop()
				if err != nil {
					return PopError(err, entry)
				}
				err = vm.Stack.Push(c)
				if err != nil {
					return PushError(err, entry)
				}
				err = vm.Stack.Push(c)
				if err != nil {
					return err
				}
				return nil
			},
			ulpAsm: PrimitiveUlp{
				"ld r0, r3, 0",
				"sub r3, r3, 1",
				"st r0, r3, 0",
				"jump __next_skip_load",
			},
			ulpAsmSrt: PrimitiveUlpSrt{
				Asm: []string{
					"ld r0, r3, 0",
					"sub r3, r3, 1",
					"st r0, r3, 0",
				},
			},
		},
		{
			name: "PICK",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				n, err := vm.Stack.PopNumber()
				if err != nil {
					return PopError(err, entry)
				}
				if int(n) >= len(vm.Stack.stack) {
					return EntryError(entry, "number out of range: %d", n)
				}
				index := len(vm.Stack.stack) - int(n) - 1
				cell := vm.Stack.stack[index]
				err = vm.Stack.Push(cell)
				if err != nil {
					return PushError(err, entry)
				}
				return err
			},
			ulpAsm: PrimitiveUlp{
				"ld r0, r3, 0",
				"add r0, r0, r3",
				"ld r0, r0, 1",
				"st r0, r3, 0",
				"jump __next_skip_load",
			},
			ulpAsmSrt: PrimitiveUlpSrt{
				Asm: []string{
					"ld r0, r3, 0",
					"add r0, r0, r3",
					"ld r0, r0, 1",
					"st r0, r3, 0",
				},
			},
		},
		{
			name: "RPICK",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				n, err := vm.Stack.PopNumber()
				if err != nil {
					return PopError(err, entry)
				}
				if int(n) >= len(vm.ReturnStack.stack) {
					return EntryError(entry, "number out of range: %d", n)
				}
				index := len(vm.ReturnStack.stack) - int(n) - 1
				cell := vm.ReturnStack.stack[index]
				err = vm.Stack.Push(cell)
				if err != nil {
					return PushError(err, entry)
				}
				return nil
			},
			ulpAsm: PrimitiveUlp{
				"ld r0, r2, __rsp",
				"ld r1, r3, 0",
				"sub r0, r0, r1",
				"ld r0, r0, 0",
				"st r0, r3, 0",
				"jump __next_skip_r2",
			},
			ulpAsmSrt: PrimitiveUlpSrt{
				Asm: []string{
					"move r1, 0",
					"ld r0, r1, __rsp",
					"ld r1, r3, 0",
					"sub r0, r0, r1",
					"ld r0, r0, 0",
					"st r0, r3, 0",
				},
			},
		},
		{
			name: "ROT",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				c, err := vm.Stack.Pop()
				if err != nil {
					return PopError(err, entry)
				}
				b, err := vm.Stack.Pop()
				if err != nil {
					return PopError(err, entry)
				}
				a, err := vm.Stack.Pop()
				if err != nil {
					return PopError(err, entry)
				}
				err = vm.Stack.Push(b)
				if err != nil {
					return PushError(err, entry)
				}
				err = vm.Stack.Push(c)
				if err != nil {
					return PushError(err, entry)
				}
				err = vm.Stack.Push(a)
				if err != nil {
					return PushError(err, entry)
				}
				return nil
			},
			ulpAsm: PrimitiveUlp{
				"ld r0, r3, 0",
				"ld r1, r3, 1",
				"st r0, r3, 1",
				"ld r0, r3, 2",
				"st r1, r3, 2",
				"st r0, r3, 0",
				"jump __next_skip_r2",
			},
			ulpAsmSrt: PrimitiveUlpSrt{
				Asm: []string{
					"ld r0, r3, 0",
					"ld r1, r3, 1",
					"st r0, r3, 1",
					"ld r0, r3, 2",
					"st r1, r3, 2",
					"st r0, r3, 0",
				},
			},
		},
		{
			name: "DROP",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				_, err := vm.Stack.Pop()
				if err != nil {
					return PopError(err, entry)
				}
				return nil
			},
			ulpAsm: PrimitiveUlp{
				"add r3, r3, 1",
				"jump __next_skip_load",
			},
			ulpAsmSrt: PrimitiveUlpSrt{
				Asm: []string{
					"add r3, r3, 1",
				},
			},
		},
		{
			name: "LOOPCHECK",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				index, err := vm.ReturnStack.PopNumber()
				if err != nil {
					return PopError(err, entry)
				}
				n, err := vm.Stack.PopNumber()
				if err != nil {
					return PopError(err, entry)
				}
				limit, err := vm.ReturnStack.PopNumber()
				if err != nil {
					return JoinEntryError(err, entry, "could not pop from return stack")
				}
				err = vm.ReturnStack.Push(CellNumber{limit}) // we just want to know the limit
				if err != nil {
					return JoinEntryError(err, entry, "could not push on return stack")
				}
				n_next := index + n
				err = vm.ReturnStack.Push(CellNumber{n_next}) // store n+index for next time
				if err != nil {
					return JoinEntryError(err, entry, "could not push on return stack")
				}
				indLim := index - limit
				var val32 uint32
				if int16(n) >= 0 {
					val32 = uint32(indLim) + uint32(n)
				} else {
					val32 = uint32(indLim) - uint32(uint16(-n))
				}
				if val32 > 0xFFFF {
					err = vm.Stack.Push(CellNumber{0xFFFF})
				} else {
					err = vm.Stack.Push(CellNumber{0})
				}
				if err != nil {
					return PushError(err, entry)
				}
				return nil
			},
			ulpAsm: PrimitiveUlp{
				"ld r2, r2, __rsp",                // load rsp
				"ld r0, r2, 0",                    // load loop index
				"ld r1, r3, 0",                    // load n
				"add r1, r1, r0",                  // n+index
				"st r1, r2, 0",                    // store n+index for next time
				"ld r1, r2, -1",                   // r1 has limit
				"sub r1, r0, r1",                  // r1 = index-limit
				"ld r0, r3, 0",                    // r0 has n
				"move r2, 0",                      // put r2 back to 0
				"jumpr __loopcheck.0, 0x7FFF, gt", // check sign of n
				// positive, add
				"add r1, r1, r0",
				"jump __loopcheck.1",
				"__loopcheck.0:",
				// negative, negate and subtract
				"sub r0, r2, r0", // -r0 = 0 - r0
				"sub r1, r1, r0",
				"__loopcheck.1:",
				"move r0, 0xFFFF",        // default to true
				"jump __loopcheck.2, ov", // check if overflow
				// no overflow, set to false
				"move r0, 0",
				"__loopcheck.2:",
				"st r0, r3, 0", // save value
				"jump __next_skip_r2",
			},
			ulpAsmSrt: PrimitiveUlpSrt{
				Asm: []string{
					"st r2, r3, -1",                   // save r2
					"move r2, __rsp",                  // put r2 on rsp address
					"ld r2, r2, 0",                    // load rsp
					"ld r0, r2, 0",                    // load loop index
					"ld r1, r3, 0",                    // load n
					"add r1, r1, r0",                  // n+index
					"st r1, r2, 0",                    // store n+index for next time
					"ld r1, r2, -1",                   // r1 has limit
					"sub r1, r0, r1",                  // r1 = index-limit
					"ld r0, r3, 0",                    // r0 has n
					"jumpr __loopcheck.0, 0x7FFF, gt", // check sign of n
					// positive, add
					"add r1, r1, r0",
					"jump __loopcheck.1",
					"__loopcheck.0:",
					// negative, negate and subtract
					"move r2, 0",     // set r2 to 0
					"sub r0, r2, r0", // -r0 = 0 - r0
					"sub r1, r1, r0",
					"__loopcheck.1:",
					"move r0, 0xFFFF",        // default to true
					"jump __loopcheck.2, ov", // check if overflow
					// no overflow, set to false
					"move r0, 0",
					"__loopcheck.2:",
					"ld r2, r3, -1", // restore r2
					"st r0, r3, 0",  // save value
				},
			},
		},
		{
			name: "U<",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				right, err := vm.Stack.PopNumber()
				if err != nil {
					return PopError(err, entry)
				}
				left, err := vm.Stack.PopNumber()
				if err != nil {
					return PopError(err, entry)
				}
				val := 0
				if left < right {
					val = 0xFFFF
				}
				err = vm.Stack.Push(CellNumber{uint16(val)})
				if err != nil {
					return PushError(err, entry)
				}
				return nil
			},
			ulpAsm: PrimitiveUlp{
				"ld r1, r3, 1",            // left
				"ld r0, r3, 0",            // right
				"sub r1, r1, r0",          // subtract
				"move r0, 0xFFFF",         // default to true
				"jump __u_lessthan.0, ov", // jump if overflow
				"move r0, 0",              // no overflow: false
				"__u_lessthan.0:",         // then
				"add r3, r3, 1",           // decrement stack
				"st r0, r3, 0",            // store the result
				"jump __next_skip_r2",
			},
			ulpAsmSrt: PrimitiveUlpSrt{
				Asm: []string{
					"ld r1, r3, 1",            // left
					"ld r0, r3, 0",            // right
					"sub r1, r1, r0",          // subtract
					"move r0, 0xFFFF",         // default to true
					"jump __u_lessthan.0, ov", // jump if overflow
					"move r0, 0",              // no overflow: false
					"__u_lessthan.0:",         // then
					"add r3, r3, 1",           // decrement stack
					"st r0, r3, 0",            // store the result
				},
			},
		},
		{
			name: "DEPTH",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				depth := len(vm.Stack.stack)
				cell := CellNumber{uint16(depth)}
				err := vm.Stack.Push(cell)
				if err != nil {
					return PushError(err, entry)
				}
				return nil
			},
			ulpAsm: PrimitiveUlp{
				"move r0, __stack_end",
				"sub r0, r0, r3",
				"sub r3, r3, 1",
				"st r0, r3, 0",
				"jump __next_skip_load",
			},
			ulpAsmSrt: PrimitiveUlpSrt{
				Asm: []string{
					"move r0, __stack_end",
					"sub r0, r0, r3",
					"sub r3, r3, 1",
					"st r0, r3, 0",
				},
			},
		},
		{
			name: "VM.STACK.INIT", // initialize the ulp stack
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				vm.Stack.stack = make([]Cell, 0)
				return nil
			},
			ulpAsm: PrimitiveUlp{
				"move r3, __stack_end", // set the stack pointer to the end of the stack
				"jump __next_skip_load",
			},
			ulpAsmSrt: PrimitiveUlpSrt{
				Asm: []string{
					"move r3, __stack_end", // set the stack pointer to the end of the stack
				},
			},
		},
		{
			name:   "HALT",
			goFunc: notImplemented,
			ulpAsm: PrimitiveUlp{
				"halt",
			},
			ulpAsmSrt: PrimitiveUlpSrt{
				Asm: []string{
					"add r2, r2, 1", // go to next instruction
					"halt",          // halt!
				},
				NonStandardNext: true,
			},
		},
		{
			name: "ESP.FUNC.UNSAFE", // use one of the custom esp32/host functions
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				funcType, err := vm.Stack.PopNumber()
				if err != nil {
					return PopError(err, entry)
				}
				cell, err := vm.Stack.Pop()
				if err != nil {
					return PopError(err, entry)
				}
				switch funcType {
				case 0: // nothing
				case 1: // done
				case 2: // print unsigned number (we're not quite accurate here for convenience)
					fmt.Fprint(vm.Out, cell, " ")
				case 3: // print char
					num, ok := cell.(CellNumber)
					if !ok {
						return EntryError(entry, "exected a number got %s type %T", cell, cell)
					}
					c := byte(num.Number)
					fmt.Fprintf(vm.Out, "%c", c)
				default:
					return EntryError(entry, "unknown function %d", funcType)
				}
				return nil
			},
			ulpAsm: PrimitiveUlp{
				"ld r0, r3, 1",           // load the value we want to print
				"ld r1, r3, 0",           // load the method number
				"st r0, r2, HOST_PARAM0", // set the param
				"st r1, r2, HOST_FUNC",   // set the function indicator
				"add r3, r3, 2",          // decrease the stack by 2
				"jump __next_skip_r2",
			},
			ulpAsmSrt: PrimitiveUlpSrt{
				Asm: []string{
					"move r1, 0",
					"ld r0, r3, 1",           // load the value we want to print
					"st r0, r1, HOST_PARAM0", // set the param
					"ld r0, r3, 0",           // load the method number
					"st r0, r1, HOST_FUNC",   // set the function indicator
					"add r3, r3, 2",          // decrease the stack by 2
				},
			},
		},
		{
			name: "ESP.FUNC.READ.UNSAFE",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				err := vm.Stack.Push(CellNumber{0}) // always return 0
				if err != nil {
					return PushError(err, entry)
				}
				return nil
			},
			ulpAsm: PrimitiveUlp{
				"ld r0, r2, HOST_FUNC",
				"sub r3, r3, 1",
				"st r0, r3, 0",
				"jump __next_skip_load",
			},
			ulpAsmSrt: PrimitiveUlpSrt{
				Asm: []string{
					"move r0, HOST_FUNC",
					"ld r0, r0, 0",
					"sub r3, r3, 1",
					"st r0, r3, 0",
				},
			},
		},
		{
			name:   "MUTEX.TAKE",
			goFunc: nop,
			ulpAsm: PrimitiveUlp{
				"move r0, 1",
				"st r0, r2, MUTEX_FLAG0",      // flag0 = 1
				"st r0, r2, MUTEX_TURN",       // turn = 1
				"__mutex.take.0:",             // //while flag1>0 && turn>0
				"ld r0, r2, MUTEX_FLAG1",      // read flag1
				"jumpr __mutex.take.1, 1, lt", // exit if flag1<1
				"ld r0, r2, MUTEX_TURN",       // read turn
				"jumpr __mutex.take.0, 0, gt", // loop if turn>0
				"__mutex.take.1:",
				"jump __next_skip_r2",
			},
			ulpAsmSrt: PrimitiveUlpSrt{
				Asm: []string{
					"move r1, 0", // use r1 as a global pointer
					"move r0, 1",
					"st r0, r1, MUTEX_FLAG0",      // flag0 = 1
					"st r0, r1, MUTEX_TURN",       // turn = 1
					"__mutex.take.0:",             // //while flag1>0 && turn>0
					"ld r0, r1, MUTEX_FLAG1",      // read flag1
					"jumpr __mutex.take.1, 1, lt", // exit if flag1<1
					"ld r0, r1, MUTEX_TURN",       // read turn
					"jumpr __mutex.take.0, 0, gt", // loop if turn>0
					"__mutex.take.1:",
				},
			},
		},
		{
			name:   "MUTEX.GIVE",
			goFunc: nop,
			ulpAsm: PrimitiveUlp{
				"st r2, r2, MUTEX_FLAG0", // flag0 = 0
				"jump __next_skip_load",
			},
			ulpAsmSrt: PrimitiveUlpSrt{
				Asm: []string{
					"move r0, 0",
					"st r0, r0, MUTEX_FLAG0", // flag0 = 0
				},
			},
		},

		{
			name: "D-", // ( xlow xhigh ylow yhigh -- zlow zhigh )
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				yHigh, err := vm.Stack.PopNumber()
				if err != nil {
					return PopError(err, entry)
				}
				yLow, err := vm.Stack.PopNumber()
				if err != nil {
					return PopError(err, entry)
				}
				xHigh, err := vm.Stack.PopNumber()
				if err != nil {
					return PopError(err, entry)
				}
				xLow, err := vm.Stack.PopNumber()
				if err != nil {
					return PopError(err, entry)
				}

				y := (uint32(yHigh) << 16) | uint32(yLow)
				x := (uint32(xHigh) << 16) | uint32(xLow)
				z := x - y
				zLow := uint16(z)
				zHigh := uint16(z >> 16)
				err = vm.Stack.Push(CellNumber{zLow})
				if err != nil {
					return PushError(err, entry)
				}
				err = vm.Stack.Push(CellNumber{zHigh})
				if err != nil {
					return PushError(err, entry)
				}
				return nil
			},
			ulpAsm: PrimitiveUlp{
				"ld r0, r3, 3",   // xlow
				"ld r1, r3, 1",   // ylow
				"sub r0, r0, r1", // subtract low
				"st r0, r3, 3",   // store zlow
				"ld r0, r3, 2",   // load xhigh
				"ld r1, r3, 0",   // load yhigh
				"jump __d_minus.0, ov",
				"jump __d_minus.1",
				"__d_minus.0:",
				"sub r0, r0, 1", // overload, subtract the carry bit
				"__d_minus.1:",
				"sub r0, r0, r1", // subtract high
				"add r3, r3, 2",  // decrement stack
				"st r0, r3, 0",   // store zhigh
				"jump __next_skip_r2",
			},
			ulpAsmSrt: PrimitiveUlpSrt{
				Asm: []string{
					"ld r0, r3, 3",   // xlow
					"ld r1, r3, 1",   // ylow
					"sub r0, r0, r1", // subtract low
					"st r0, r3, 3",   // store zlow
					"ld r0, r3, 2",   // load xhigh
					"ld r1, r3, 0",   // load yhigh
					"jump __d_minus.0, ov",
					"jump __d_minus.1",
					"__d_minus.0:",
					"sub r0, r0, 1", // overload, subtract the carry bit
					"__d_minus.1:",
					"sub r0, r0, r1", // subtract high
					"add r3, r3, 2",  // decrement stack
					"st r0, r3, 0",   // store zhigh
				},
			},
		},
		{
			name: "D+", // ( xlow xhigh ylow yhigh -- zlow zhigh )
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				yHigh, err := vm.Stack.PopNumber()
				if err != nil {
					return PopError(err, entry)
				}
				yLow, err := vm.Stack.PopNumber()
				if err != nil {
					return PopError(err, entry)
				}
				xHigh, err := vm.Stack.PopNumber()
				if err != nil {
					return PopError(err, entry)
				}
				xLow, err := vm.Stack.PopNumber()
				if err != nil {
					return PopError(err, entry)
				}

				y := (uint32(yHigh) << 16) | uint32(yLow)
				x := (uint32(xHigh) << 16) | uint32(xLow)
				z := x + y
				zLow := uint16(z)
				zHigh := uint16(z >> 16)
				err = vm.Stack.Push(CellNumber{zLow})
				if err != nil {
					return PushError(err, entry)
				}
				err = vm.Stack.Push(CellNumber{zHigh})
				if err != nil {
					return PushError(err, entry)
				}
				return nil
			},
			ulpAsm: PrimitiveUlp{
				"ld r0, r3, 3",   // xlow
				"ld r1, r3, 1",   // ylow
				"add r0, r0, r1", // add low
				"st r0, r3, 3",   // store zlow
				"ld r0, r3, 2",   // load xhigh
				"ld r1, r3, 0",   // load yhigh
				"jump __d_minus.0, ov",
				"jump __d_minus.1",
				"__d_minus.0:",
				"add r0, r0, 1", // overload, add the carry bit
				"__d_minus.1:",
				"add r0, r0, r1", // add high
				"add r3, r3, 2",  // decrement stack
				"st r0, r3, 0",   // store zhigh
				"jump __next_skip_r2",
			},
			ulpAsmSrt: PrimitiveUlpSrt{
				Asm: []string{
					"ld r0, r3, 3",   // xlow
					"ld r1, r3, 1",   // ylow
					"add r0, r0, r1", // add low
					"st r0, r3, 3",   // store zlow
					"ld r0, r3, 2",   // load xhigh
					"ld r1, r3, 0",   // load yhigh
					"jump __d_minus.0, ov",
					"jump __d_minus.1",
					"__d_minus.0:",
					"add r0, r0, 1", // overload, add the carry bit
					"__d_minus.1:",
					"add r0, r0, r1", // subtract high
					"add r3, r3, 2",  // decrement stack
					"st r0, r3, 0",   // store zhigh
				},
			},
		},
	}
	for _, p := range prims {
		err := primitiveAdd(vm, p.name, p.goFunc, p.ulpAsm, p.ulpAsmSrt, p.flag)
		if err != nil {
			return err
		}
	}
	return nil
}

func notImplemented(vm *VirtualMachine, entry *DictionaryEntry) error {
	return EntryError(entry, "cannot be executed on the host")
}

func nop(vm *VirtualMachine, entry *DictionaryEntry) error {
	return nil
}

func parseWord(vm *VirtualMachine, entry *DictionaryEntry) (string, error) {
	cellStr, err := vm.Stack.Pop()
	if err != nil {
		return "", JoinEntryError(err, entry, "could not pop name")
	}
	cellAddr, ok := cellStr.(CellAddress)
	if !ok {
		return "", EntryError(entry, "name argument needs to be an address to a string")
	}
	name, err := cellsToString(cellAddr.Entry.Word.(*WordForth).Cells)
	if err != nil {
		return "", JoinEntryError(err, entry, "could not parse name")
	}
	return name, nil
}

func parseAssembly(vm *VirtualMachine, entry *DictionaryEntry) ([]string, error) {
	count, err := vm.Stack.PopNumber()
	if err != nil {
		return nil, JoinEntryError(err, entry, "count argument requires a number")
	}
	asm := make([]string, 0)
	for i := uint16(0); i < count; i++ {
		cell, err := vm.Stack.Pop()
		if err != nil {
			return nil, JoinEntryError(err, entry, "could not pop object")
		}
		switch c := cell.(type) {
		case CellAddress:
			substr, err := cellsToString(c.Entry.Word.(*WordForth).Cells)
			if err != nil {
				return nil, JoinEntryError(err, entry, "could not convert input to string")
			}
			asm = append(asm, substr)
		case CellNumber:
			asm = append(asm, strconv.Itoa(int(c.Number)))
		default:
			return nil, EntryError(entry, "unknown argument type %T argument %s", c, c)
		}
	}
	slices.Reverse(asm)
	asmStr := strings.Join(asm, "")
	asmStr = strings.ReplaceAll(asmStr, "\\r", "\r")
	asmStr = strings.ReplaceAll(asmStr, "\\n", "\n")
	asm = strings.Split(asmStr, "\n")
	return asm, nil
}
