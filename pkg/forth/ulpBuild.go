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
	"strings"
)

// The type of information that we are currently cross-compiling.
type UlpCompileType int

const (
	UlpCompileProgram UlpCompileType = iota
	UlpCompileData
)

// The target that we are cross-compiling.
type UlpCompileTarget int

const (
	UlpCompileTargetToken = iota
	UlpCompileTargetSubroutine
)

type Ulp struct {
	outCount int

	// output definitions
	forthWords    []*WordForth
	assemblyWords []*WordPrimitive
	dataWords     []*WordForth
	literals      map[string]string

	// current state of compilation
	compileTarget UlpCompileTarget
}

// Build the assembly using the word passed in as the main function.
// Note that the virtual machine will be unusable after this.
func (u *Ulp) BuildAssembly(vm *VirtualMachine, word string) (string, error) {
	// put back into interpret state and compile the main ulp program

	vm.State.Set(uint16(StateInterpret))
	// create the VM.INIT word without an EXIT
	err := vm.Execute([]byte(" BL WORD VM.INIT --CREATE-FORTH ] VM.STACK.INIT " + word + " BEGIN HALT AGAIN [ LAST HIDE "))
	if err != nil {
		return "", errors.Join(fmt.Errorf("could not compile the supporting words for ulp cross-compiling"), err)
	}
	u.compileTarget = UlpCompileTargetToken
	return u.buildAssemblyHelper(vm, u.buildInterpreter())
}

func (u *Ulp) BuildAssemblySrt(vm *VirtualMachine, word string) (string, error) {
	vm.State.Set(uint16(StateInterpret))
	// create the VM.INIT word without an EXIT
	err := vm.Execute([]byte(" BL WORD VM.INIT --CREATE-FORTH ] " + word + " BEGIN HALT AGAIN [ LAST HIDE "))
	if err != nil {
		return "", errors.Join(fmt.Errorf("could not compile the supporting words for ulp cross-compiling"), err)
	}
	u.compileTarget = UlpCompileTargetSubroutine
	return u.buildAssemblyHelper(vm, u.buildInterpreterSrt())
}

func (u *Ulp) buildAssemblyHelper(vm *VirtualMachine, header string) (string, error) {
	vmInitEntry := vm.Dictionary.Entries[len(vm.Dictionary.Entries)-1]
	// generate the various lists
	err := u.buildLists(vmInitEntry)
	if err != nil {
		return "", err
	}
	// optimize!
	optimizer := Optimizer{u}
	err = optimizer.Optimize()
	if err != nil {
		return "", err
	}
	// rebuild the various lists because optimizations may have changed things
	err = u.clearLists()
	if err != nil {
		return "", err
	}
	err = u.buildLists(vmInitEntry)
	if err != nil {
		return "", err
	}
	// count the number of calls
	u.countCalls()
	// create the different assemblies
	asm, err := u.buildAssemblyWords()
	if err != nil {
		return "", err
	}
	forth, err := u.buildForthWords()
	if err != nil {
		return "", err
	}
	data, err := u.buildDataWords()
	if err != nil {
		return "", err
	}
	literals, err := u.buildLiterals()
	if err != nil {
		return "", err
	}
	forthSection := ".data"
	if u.compileTarget == UlpCompileTargetSubroutine {
		forthSection = ".text"
	}

	// put assemblies together
	i := []string{
		header,
		"__assembly_words:",
		".text",
		asm,
		forthSection,
		"__forth_words:",
		forth,
		"__data_words:",
		".data",
		literals,
		data,
		"__data_end:",
	}
	return strings.Join(i, "\r\n"), nil
}

// Convert list of used subroutine-threaded assembly
// words into a string.
func (u *Ulp) buildAssemblyWords() (string, error) {
	asmList := make([]string, len(u.assemblyWords))
	for i, word := range u.assemblyWords {
		asm, err := word.BuildAssembly(u)
		if err != nil {
			return "", err
		}
		asmList[i] = asm
	}
	return strings.Join(asmList, "\r\n\r\n"), nil
}

// Convert list of subroutine-threaded forth
// words into a string.
func (u *Ulp) buildForthWords() (string, error) {
	output := make([]string, len(u.forthWords))
	for i, word := range u.forthWords {
		asm, err := word.BuildAssembly(u)
		if err != nil {
			return "", err
		}
		output[i] = asm
	}
	return strings.Join(output, "\r\n\r\n"), nil
}

// Convert list of data
// words into a string.
func (u *Ulp) buildDataWords() (string, error) {
	output := make([]string, len(u.dataWords))
	for i, word := range u.dataWords {
		asm, err := word.BuildAssembly(u)
		if err != nil {
			return "", err
		}
		if word.Entry.Flag.GlobalData {
			label := word.Entry.ulpName
			global := ".global " + label + "\r\n"
			asm = global + asm
		}
		output[i] = asm
	}
	return strings.Join(output, "\r\n"), nil
}

func (u *Ulp) buildLiterals() (string, error) {
	switch u.compileTarget {
	case UlpCompileTargetToken:
		output := make([]string, len(u.literals))
		i := 0
		for name, val := range u.literals {
			output[i] = fmt.Sprintf("%s: .int %s", name, val)
			i += 1
		}
		return strings.Join(output, "\r\n"), nil
	case UlpCompileTargetSubroutine:
		return "", nil // literals are compiled along the way, not as a final list
	default:
		return "", fmt.Errorf("unknown compile target %d, please file a bug report", u.compileTarget)
	}
}

func (u *Ulp) buildLists(entry *DictionaryEntry) error {
	u.forthWords = make([]*WordForth, 0)
	u.assemblyWords = make([]*WordPrimitive, 0)
	u.dataWords = make([]*WordForth, 0)
	u.literals = make(map[string]string)

	return entry.AddToList(u)
}

func (u *Ulp) clearLists() error {
	for _, w := range u.forthWords {
		w.Entry.Flag.addedToList = false
	}
	for _, w := range u.assemblyWords {
		w.Entry.Flag.addedToList = false
	}
	for _, w := range u.dataWords {
		w.Entry.Flag.addedToList = false
	}
	return nil
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

// Count the number of times that each word is called,
// not including tail calls.
func (u *Ulp) countCalls() {
	u.resetCalls()
	for _, w := range u.forthWords {
		for _, c := range w.Cells {
			switch cell := c.(type) {
			case CellAddress:
				cell.Entry.Flag.calls += 1
			}
		}
	}
}

// Reset the number of times that each word is called.
func (u *Ulp) resetCalls() {
	for _, w := range u.forthWords {
		for _, c := range w.Cells {
			switch cell := c.(type) {
			case CellAddress:
				cell.Entry.Flag.calls = 0
			}
		}
	}
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
		"__ip:  .int __body__forth_VM.INIT", // instruction pointer starts at word VM.INIT
		"__rsp: .int __stack_start",         // return stack pointer starts at the beginning of the stack section

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
		"jumpr __ins_forth, __forth_words, ge",
		// it's assembly
		"st r1, r2, __ip", // store the instruction pointer so asm can use r1
		"jump r0",         // jump to the assembly

		"__ins_forth:",
		"jumpr __ins_num, __data_words, ge",
		// it's forth
		"st r0, r2, __ip",     // put the address into the instruction pointer
		"ld r0, r2, __rsp",    // load the return stack pointer
		"add r0, r0, 1",       // increment the rsp
		"st r1, r0, 0",        // store the instruction we were about to execute onto the return stack
		"st r0, r2, __rsp",    // store the rsp
		"jump __next_skip_r2", // then start the vm again at the defined instruction

		"__ins_num:",
		"jumpr __ins_branch0, __data_end, gt",
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

		// registers are set to 0 when the esp32
		// initializes the ulp program,
		// so we can see if r2 has been set or not
		"add r0, r2, 0xFFFF", // will overflow unless r2 is 0
		"jump r2, ov",        // jump to next instruction if overflowed
		// setup the pointers
		"move r2, __body__forth_VM.INIT", // instruction pointer goes to init word
		"move r3, __stack_end",           // set up stack pointer
		"jump r2",                        // begin execution

		".text",
		// subroutine to set up the forth word return
		"__docol:",
		"move r0, 0",
		"ld r1, r0, __rsp", // load the return stack pointer
		"add r1, r1, 1",    // increase rsp
		"st r2, r1, 0",     // store the current address on return stack
		"st r1, r0, __rsp", // store rsp
		"ld r2, r2, 0",     // load the lower half of the "jump" instruction
		"rsh r2, r2, 2",    // isolate the jump address
		"add r2, r2, 1",    // increment the instruction pointer past the "jump __docol"
		"jump r2",          // jump to first instruction!

		// subroutine to add the value on r0 to the stack
		"__add_to_stack:",
		"sub r3, r3, 1", // increment the stack
		"st r0, r3, 0",  // store the value
		"add r2, r2, 2", // increase the instruction pointer by 2
		"jump r2",       // and continue executing!

		// subroutine to conditional jump
		"__branch_if:",
		"ld r0, r3, 0",  // load the top of stack
		"add r3, r3, 1", // decrement stack
		"jumpr __branch_if.0, 1, lt",
		// don't jump
		"add r2, r2, 2",
		"jump r2",
		// jump
		"__branch_if.0:",
		"move r2, r1", // copy the new address
		"jump r2",     // and jump to it!
	}
	return strings.Join(i, "\r\n") + "\r\n"
}
