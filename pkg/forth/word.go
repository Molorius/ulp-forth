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
	"strings"
)

// An executable Forth Word.
type Word interface {
	Execute(*VirtualMachine) error // Execute the word.
	AddToList(*Ulp) error
	BuildAssembly(*Ulp) (string, error)
	IsRecursive(*WordForth) bool
}

// A Word built using other Forth words/numbers.
type WordForth struct {
	Cells []Cell           // The cells that constitute this word.
	Entry *DictionaryEntry // The associated dictionary entry.
}

func (w *WordForth) ExecuteOffset(vm *VirtualMachine, offset int) error {
	// push the instruction pointer onto the return stack
	startDepth := vm.ReturnStack.Depth()
	previous := vm.IP // keep the previous address for error popping
	err := vm.ReturnStack.Push(previous)
	if err != nil {
		return errors.Join(fmt.Errorf("%s could not push instruction pointer to return stack.", w.Entry), err)
	}
	if len(w.Cells) == 0 {
		return EntryError(w.Entry, "This forth word doesn't have a definition")
	}
	vm.IP = &CellAddress{Entry: w.Entry, Offset: offset} // create a new instruction pointer

	// execute each lower word
	for vm.ReturnStack.Depth() > startDepth { // while the original instruction pointer hasn't been popped
		if vm.IP.Entry != w.Entry { // we somehow left this word
			vm.IP = previous                    // reset the instruction pointer
			vm.ReturnStack.SetDepth(startDepth) // attempt to reset the stack depth to "fix" part of the problem
			return fmt.Errorf("%s instruction pointer somehow left the calling word.", w.Entry)
		}
		if vm.IP.Offset < 0 || vm.IP.Offset >= len(w.Cells) { // the instruction pointer is out of bounds
			vm.IP = previous                    // reset the instruction pointer
			vm.ReturnStack.SetDepth(startDepth) // attempt to reset the stack depth to "fix" part of the problem
			return fmt.Errorf("%s instruction pointer went outside of definition.", w.Entry)
		}
		currentOffset := vm.IP.Offset
		nxt := w.Cells[currentOffset]
		vm.IP.Offset += 1
		err := nxt.Execute(vm)
		if err != nil { // error when executing lower word
			vm.IP = previous                    // reset the instruction pointer
			vm.ReturnStack.SetDepth(startDepth) // attempt to reset the stack depth to "fix" part of the problem
			return errors.Join(fmt.Errorf("%s error while executing %s in position %d", w.Entry, nxt, currentOffset), err)
		}
	}

	// check that return stack pointer and instruction pointer are the same as when we started
	if vm.IP != previous {
		vm.IP = previous                    // reset the instruction pointer
		vm.ReturnStack.SetDepth(startDepth) // attempt to reset the stack depth to "fix" part of the problem
		return fmt.Errorf("%s instruction pointer not correct on exit", w.Entry)
	}
	if vm.ReturnStack.Depth() != startDepth {
		vm.ReturnStack.SetDepth(startDepth) // attempt to reset the stack depth to "fix" part of the problem
		return fmt.Errorf("%s return stack wrong size on exit", w.Entry)
	}
	return nil
}

func (w *WordForth) Execute(vm *VirtualMachine) error {
	return w.ExecuteOffset(vm, 0)
}

func (w *WordForth) AddToList(u *Ulp) error {
	if w.Entry.Flag.addedToList {
		return nil
	}
	// create the compiled name
	if w.Entry.Flag.Data { // if this is a data entry
		if w.Entry.Name == "" { // if it doesn't have a name assigned
			w.Entry.Name = u.name("data", "unnamed", true) // create one
		}
		w.Entry.ulpName = w.Entry.Name
	} else { // if this is a word entry
		w.Entry.ulpName = u.name("forth", w.Entry.Name, true)
	}
	// add this word to the list
	w.Entry.Flag.addedToList = true
	if w.Entry.Flag.Data {
		u.dataWords = append(u.dataWords, w)
	} else {
		u.forthWords = append(u.forthWords, w)
	}

	// add every cell
	for _, c := range w.Cells {
		err := c.AddToList(u)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *WordForth) BuildAssembly(u *Ulp) (string, error) {
	output := make([]string, 1)
	label := w.Entry.ulpName + ":"
	bodyLabel := w.Entry.BodyLabel() + ":"
	output[0] = label
	if w.Entry.Flag.Data { // data word
		output = append(output, bodyLabel)
		for _, cell := range w.Cells {
			ref, err := cell.OutputReference(u)
			if err != nil {
				return "", err
			}
			_, ok := cell.(CellAddress)
			if ok {
				ref = "__body" + ref
			}
			val := ".int " + ref
			output = append(output, val)
		}
	} else { // executable forth word
		if u.compileTarget == UlpCompileTargetSubroutine {
			if w.Entry.Flag.calls != 0 { // if this word is directly called
				output = append(output, "jump __docol")
			}
		}
		output = append(output, bodyLabel)
		for _, cell := range w.Cells {
			asm, err := cell.BuildExecution(u)
			if err != nil {
				return "", err
			}
			output = append(output, asm)
		}
	}
	return strings.Join(output, "\r\n"), nil
}

func (w *WordForth) IsRecursive(check *WordForth) bool {
	if w == check {
		return true
	}
	if w.Entry.Flag.visited {
		return false
	}
	w.Entry.Flag.visited = true
	for _, c := range w.Cells {
		if c.IsRecursive(check) {
			return true
		}
	}
	return false
}

// The Go code for a primitive Word.
type PrimitiveGo func(*VirtualMachine, *DictionaryEntry) error

type TokenNextType int

const (
	// update the constants in 11_misc.f if this list changes
	TokenNextNonstandard TokenNextType = iota
	TokenNextNormal
	TokenNextSkipR2
	TokenNextSkipLoad
)

// The ULP assembly for a primitive Word that uses token threading.
// type PrimitiveUlp []string
type PrimitiveUlp struct {
	Asm  []string
	Next TokenNextType
}

// The ULP assembly for a primitive Word that uses subroutine threading
type PrimitiveUlpSrt struct {
	Asm             []string // the assembly, not including NEXT if standard
	NonStandardNext bool     // this word uses a nonstandard NEXT ending
}

// A Word defined using Go and ULP assembly.
type WordPrimitive struct {
	Go     PrimitiveGo      // The Go function to be executed.
	Ulp    PrimitiveUlp     // The ULP assembly to be compiled.
	UlpSrt PrimitiveUlpSrt  // The ULP assembly using subroutine threading to be compiled.
	Entry  *DictionaryEntry // The associated dictionary entry.
}

func (w *WordPrimitive) Execute(vm *VirtualMachine) error {
	return w.Go(vm, w.Entry)
}

func (w *WordPrimitive) AddToList(u *Ulp) error {
	if w.Entry.Flag.addedToList {
		return nil
	}
	// create the compiled name
	w.Entry.ulpName = u.name("asm", w.Entry.Name, true)
	// add it to the list
	u.assemblyWords = append(u.assemblyWords, w)
	w.Entry.Flag.addedToList = true
	return nil
}

func (w *WordPrimitive) BuildAssembly(u *Ulp) (string, error) {
	label := w.Entry.ulpName + ":\r\n"
	bodyLabel := w.Entry.BodyLabel() + ":\r\n"
	asm := make([]string, 0)
	switch u.compileTarget {
	case UlpCompileTargetToken:
		asm = append(asm, w.Ulp.Asm...)
		switch w.Ulp.Next {
		case TokenNextNonstandard:
		case TokenNextNormal:
			asm = append(asm, "jump next")
		case TokenNextSkipR2:
			asm = append(asm, "jump __next_skip_r2")
		case TokenNextSkipLoad:
			asm = append(asm, "jump __next_skip_load")
		default:
			return "", fmt.Errorf("Unknown compile target %d, please file a bug report", w.Ulp.Next)
		}
	case UlpCompileTargetSubroutine:
		asm = append(asm, w.UlpSrt.Asm...)
		if !w.UlpSrt.NonStandardNext {
			standardNext := []string{
				"add r2, r2, 1",
				"jump r2",
			}
			asm = append(asm, standardNext...)
		}
	default:
		return "", fmt.Errorf("Unknown compile target %d, please file a bug report", u.compileTarget)
	}
	asmStr := strings.Join(asm, "\r\n")
	out := label + bodyLabel + asmStr
	return out, nil
}

func (w *WordPrimitive) IsRecursive(check *WordForth) bool {
	return false
}
