/*
Copyright 2024 Blake Felt blake.w.felt@gmail.com

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/
package forth

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type ulpAsm struct {
	name string
	asm  []string
}

func (uAsm ulpAsm) build() string {
	var sb strings.Builder
	sb.WriteString(uAsm.name)
	sb.WriteString(":\r\n")
	for _, asm := range uAsm.asm {
		if !strings.Contains(asm, ":") {
			sb.WriteString("    ")
		}
		sb.WriteString(asm)
		sb.WriteString("\r\n")
	}
	return sb.String()
}

type ulpForthCell struct {
	cell Cell
	name string
}

type ulpForth struct {
	name  string
	cells []ulpForthCell
}

func (uForth ulpForth) build() string {
	var sb strings.Builder
	sb.WriteString(uForth.name)
	sb.WriteString(":\r\n")
	for _, cell := range uForth.cells {
		if strings.Contains(cell.name, ":") {
			sb.WriteString("  ")
		} else {
			sb.WriteString("    .int ")
		}
		sb.WriteString(cell.name)
		sb.WriteString("\r\n")
	}
	return sb.String()
}

type Ulp struct {
	// output strings
	assembly []ulpAsm
	forth    []ulpForth
	data     map[string]string
	outCount int

	// output definitions
	forthWords    []*WordForth
	assemblyWords []*WordPrimitive
	dataWords     []*WordForth
}

func (u *Ulp) build() string {
	var sb strings.Builder
	// header
	sb.WriteString(u.buildInterpreter())
	// assembly
	sb.WriteString("\r\n.text\r\n")
	sb.WriteString("__asmwords_start:\r\n")
	for _, asm := range u.assembly {
		sb.WriteString(asm.build())
	}
	sb.WriteString("__asmwords_end:\r\n\r\n")
	// forth
	sb.WriteString(".data\r\n")
	sb.WriteString("__forthwords_start:\r\n")
	for _, forth := range u.forth {
		sb.WriteString(forth.build())
	}
	sb.WriteString("__forthwords_end:\r\n\r\n")
	// data
	sb.WriteString("__forthdata_start:\r\n")
	for _, d := range u.data {
		sb.WriteString(d)
		sb.WriteString("\r\n")
	}
	sb.WriteString("__forthdata_end:\r\n")
	return sb.String()
}

// Build the assembly using the word passed in as the main function.
// Note that the virtual machine will be unusable after this.
func (u *Ulp) BuildAssembly(vm *VirtualMachine, word string) (string, error) {
	// put back into interpret state and compile the main ulp program
	// u.Entries = make([]*DictionaryEntry, 0)
	u.assembly = make([]ulpAsm, 0)
	u.forth = make([]ulpForth, 0)
	u.data = make(map[string]string)

	vm.State.Set(uint16(StateInterpret))
	err := vm.Execute([]byte(" : VM.INIT VM.STACK.INIT " + word + " BEGIN HALT AGAIN ; "))
	if err != nil {
		return "", errors.Join(fmt.Errorf("could not compile the supporting words for ulp cross-compiling."), err)
	}
	vmInitEntry := vm.Dictionary.Entries[len(vm.Dictionary.Entries)-1]
	_, err = u.findUsedEntry(vmInitEntry)
	str := u.build()
	return str, err
}

func (u *Ulp) BuildAssemblySrt(vm *VirtualMachine, word string) (string, error) {
	vm.State.Set(uint16(StateInterpret))
	err := vm.Execute([]byte(" : VM.INIT VM.STACK.INIT " + word + " BEGIN HALT AGAIN ; "))
	if err != nil {
		return "", errors.Join(fmt.Errorf("could not compile the supporting words for ulp cross-compiling."), err)
	}
	vmInitEntry := vm.Dictionary.Entries[len(vm.Dictionary.Entries)-1]
	err = u.buildLists(vmInitEntry)
	if err != nil {
		return "", err
	}

	return "", fmt.Errorf("subroutine threading not implemented")
}

// recursively find all dictionary entries that the entry uses
func (u *Ulp) findUsedEntry(entry *DictionaryEntry) (string, error) {
	if entry.ulpName != "" { // name is already set, we've already found this
		return entry.ulpName, nil
	}
	if entry.Flag.Data { // if this is a data entry
		if entry.Name == "" {
			entry.Name = u.name("data", "unnamed", true)
		}
		entry.ulpName = entry.Name
		_, ok := u.data[entry.ulpName]
		if !ok {
			// build the definition
			var sb strings.Builder
			sb.WriteString(".global ")
			sb.WriteString(entry.Name)
			sb.WriteString("\r\n")
			sb.WriteString(entry.Name)
			sb.WriteString(":")
			w, ok := entry.Word.(*WordForth)
			if !ok {
				return "", fmt.Errorf("%s cannot build a data entry that doesn't use a forth word", entry.Name)
			}
			vals := make([]string, 0)
			for _, c := range w.Cells {
				str, err := u.findUsedCell(c)
				if err != nil {
					return "", errors.Join(fmt.Errorf("%s error while compiling", entry.Name), err)
				}
				if strings.Contains(str, ":") {
					return "", fmt.Errorf("%s cannot compile an address inside data", entry.Name)
				}
				vals = append(vals, str)
			}
			sb.WriteString("  .int ")
			sb.WriteString(strings.Join(vals, ", "))
			u.data[entry.Name] = sb.String()
		}
		return entry.ulpName, nil
	}
	switch w := entry.Word.(type) {
	case *WordForth:
		name := u.name("forth", entry.Name, true)
		entry.ulpName = name
		forth := ulpForth{
			name:  name,
			cells: make([]ulpForthCell, 0),
		}

		for _, c := range w.Cells {
			str, err := u.findUsedCell(c)
			if err != nil {
				return "", errors.Join(fmt.Errorf("%s error while compiling", entry.Name), err)
			}
			if str == "DOCOL" {
				continue
			}

			cell := ulpForthCell{
				cell: c,
				name: str,
			}
			forth.cells = append(forth.cells, cell)
		}
		u.forth = append(u.forth, forth)
		return name, nil
	case *WordPrimitive:
		if entry.Name == "DOCOL" {
			return "DOCOL", nil
		}
		name := u.name("asm", entry.Name, true)
		entry.ulpName = name
		if w.Ulp == nil {
			return "", fmt.Errorf("Cannot compile primitive without ulp assembly: %v", w)
		}
		asm := ulpAsm{
			name: name,
			asm:  w.Ulp,
		}
		u.assembly = append(u.assembly, asm)
		return name, nil
	default:
		return "", fmt.Errorf("Type %T not supported for cross compile, word: %s", w, w)
	}
}

func (u *Ulp) buildLists(entry *DictionaryEntry) error {
	u.forthWords = make([]*WordForth, 0)
	u.assemblyWords = make([]*WordPrimitive, 0)
	u.dataWords = make([]*WordForth, 0)

	return entry.AddToList(u)
}

func (u *Ulp) findUsedCell(cell Cell) (string, error) {
	switch c := cell.(type) {
	case CellNumber:
		return strconv.Itoa(int(c.Number)), nil
	case CellLiteral:
		pointedName, err := u.findUsedCell(c.cell)
		if err != nil {
			return "", err
		}
		var name string
		_, isNum := c.cell.(CellNumber)
		if isNum {
			name = u.name("number", pointedName, false)
		} else {
			name = u.name("literal", pointedName, false)
		}
		_, ok := u.data[name]
		if !ok {
			u.data[name] = fmt.Sprintf("%s: .int %s", name, pointedName)
		}
		return name, nil
	case CellAddress:
		name, err := u.findUsedEntry(c.Entry)
		if c.Offset != 0 {
			name = fmt.Sprintf("%s+%d", name, c.Offset)
		}
		if c.UpperByte {
			name = name + "+0x8000" // set the highest bit
		}
		return name, err
	case *CellBranch0:
		return fmt.Sprintf("%s + 0x4000", c.dest.name(u)), nil
	case *CellBranch:
		return fmt.Sprintf("%s + 0x8000", c.dest.name(u)), nil
	case *CellDestination:
		return fmt.Sprintf("%s:", c.name(u)), nil
	default:
		return "", fmt.Errorf("Type %T not supported for cross compile, cell: %s", c, c)
	}
}

func (u *Ulp) name(middle string, word string, addSuffix bool) string {
	// Forth words can have any character, replace them all
	fixed := u.replaceOtherChars(word)
	if word == "VM.INIT" { // keep this static so we can keep the vm code static
		return "__forth_VM.INIT"
	}
	if addSuffix {
		n := fmt.Sprintf("__%s_%s_%d", middle, fixed, u.outCount)
		u.outCount++
		return n
	} else {
		return fmt.Sprintf("__%s_%s", middle, fixed)
	}
}

// replace all non-alphanumeric chars.
func (u *Ulp) replaceOtherChars(str string) string {
	var sb strings.Builder
	bytes := []byte(str)

	for _, b := range bytes {
		switch true {
		case b >= '0' && b <= '9',
			b >= 'a' && b <= 'z',
			b >= 'A' && b <= 'Z',
			b == '.',
			b == '_':
			sb.WriteByte(b)
		default:
			sb.WriteString(fmt.Sprintf("_ascii%d_", int(b)))
		}
	}
	return sb.String()
}

func (u *Ulp) buildInterpreter() string {
	i := []string{
		// required data, will be placed at the start of .data
		".boot.data",
		".global MUTEX_FLAG0",
		".global MUTEX_FLAG1",
		".global MUTEX_TURN",
		".global HOST_FUNC",
		".global HOST_PARAM0",
		"MUTEX_FLAG0: .int 0", // DO NOT reorder these, the same address relative to .data
		"MUTEX_FLAG1: .int 0", // is used to easily find for the esp32 and the emulator.
		"MUTEX_TURN:  .int 0",
		"HOST_FUNC:   .int 0",
		"HOST_PARAM0: .int 0",
		".data",
		"__ip:  .int __forth_VM.INIT", // instruction pointer starts at word VM.INIT
		"__rsp: .int __stack_start",   // return stack pointer starts at the beginning of the stack section

		// boot labels
		".boot",
		".global entry",
		"entry:",
		"next:",

		// load the instruction
		"move r2, 0",        // r2 is 0 at the start of every loop as a global pointer
		"__next_skip_r2:",   // address to skip loading r2
		"ld r1, r2, __ip",   // load the instruction pointer
		"__next_skip_load:", // address to skip loading IP
		"add r1, r1, 1",     // increment the pointer to the next instruction
		"ld r0, r1, -1",     // load the current instruction

		// determine instruction type
		"__ins_asm:",
		"jumpr __ins_forth, __forthwords_start, ge",
		// it's assembly
		"st r1, r2, __ip", // store the instruction pointer so asm can use r1
		"jump r0",         // jump to the assembly

		"__ins_forth:",
		"jumpr __ins_num, __forthdata_start, ge",
		// it's forth
		"st r0, r2, __ip",     // put the address into the instruction pointer
		"ld r0, r2, __rsp",    // load the return stack pointer
		"add r0, r0, 1",       // increment the rsp
		"st r1, r0, 0",        // store the instruction we were about to execute onto the return stack
		"st r0, r2, __rsp",    // store the rsp
		"jump __next_skip_r2", // then start the vm again at the defined instruction

		"__ins_num:",
		"jumpr __ins_branch0, __forthdata_end, gt",
		// it's a number or variable
		"ld r0, r0, 0",          // load the number
		"sub r3, r3, 1",         // increase the stack by 1
		"st r0, r3, 0",          // store the number
		"jump __next_skip_load", // next!

		"__ins_branch0:",
		"jumpr __ins_branch, 0x8000, ge",
		// it's a conditional branch, check stack
		"ld r0, r3, 0",                  // get value from stack
		"add r3, r3, 1",                 // decrement stack
		"jumpr __next_skip_load, 1, ge", // continue forth execution if not 0
		"ld r0, r1, -1",                 // otherwise reload the address and branch!

		"__ins_branch:",
		// it's a definite branch
		"and r1, r0, 0x3FFF",    // get the lowest 14 bits
		"jump __next_skip_load", // then continue vm at this newer address
	}
	return strings.Join(i, "\r\n") + "\r\n"
}

func (u *Ulp) buildInterpreterSrt() string {
	i := []string{
		// required data, will be placed at the start of .data
		".boot.data",
		".global MUTEX_FLAG0",
		".global MUTEX_FLAG1",
		".global MUTEX_TURN",
		".global HOST_FUNC",
		".global HOST_PARAM0",
		"MUTEX_FLAG0: .int 0", // DO NOT reorder these, the same address relative to .data
		"MUTEX_FLAG1: .int 0", // is used to easily find for the esp32 and the emulator.
		"MUTEX_TURN:  .int 0",
		"HOST_FUNC:   .int 0",
		"HOST_PARAM0: .int 0",
		".data",
		"__rsp: .int __stack_start", // return stack pointer starts at the beginning of the stack section

		// boot labels
		".boot",
		".global entry",
		"entry:",

		// TODO start executing code in some better way.
		// the +1 is to jump over the DOCOL
		"jump __forth_VM.INIT + 1",
		".text",
	}
	return strings.Join(i, "\r\n") + "\r\n"
}
