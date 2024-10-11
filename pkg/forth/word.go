package forth

import (
	"errors"
	"fmt"
)

// An executable Forth Word.
type Word interface {
	Execute(*VirtualMachine) error // Execute the word.
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

// The Go code for a primitive Word.
type PrimitiveGo func(*VirtualMachine, *DictionaryEntry) error

// The ULP assembly for a primitive Word.
type PrimitiveUlp []string

// A Word defined using Go and ULP assembly.
type WordPrimitive struct {
	Go    PrimitiveGo      // The Go function to be executed.
	Ulp   PrimitiveUlp     // The ULP assembly to be compiled.
	Entry *DictionaryEntry // The associated dictionary entry.
}

func (w *WordPrimitive) Execute(vm *VirtualMachine) error {
	return w.Go(vm, w.Entry)
}
