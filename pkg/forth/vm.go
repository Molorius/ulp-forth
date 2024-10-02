package forth

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/ergochat/readline"
)

// The Forth virtual machine.
type VirtualMachine struct {
	Dictionary  Dictionary   // The Forth dictionary.
	Stack       Stack        // The data stack.
	ReturnStack Stack        // The return stack.
	ParseArea   ParseArea    // The input parse area.
	State       State        // The execution state for the virtual machine.
	IP          *CellAddress // The interpreter pointer.
}

// Set up the virtual machine.
func (vm *VirtualMachine) Setup() error {
	err := vm.Dictionary.Setup()
	if err != nil {
		return err
	}

	err = vm.Stack.Setup()
	if err != nil {
		return err
	}

	err = vm.ReturnStack.Setup()
	if err != nil {
		return err
	}

	err = vm.ParseArea.Setup()
	if err != nil {
		return err
	}

	err = vm.State.Setup(vm)
	if err != nil {
		return err
	}

	vm.IP = nil

	err = PrimitiveSetup(vm)
	if err != nil {
		return err
	}

	return nil
}

// Start up a read-eval-print loop.
func (vm *VirtualMachine) Repl() error {
	rl, err := readline.New("")
	if err != nil {
		return errors.Join(fmt.Errorf("Unable to start readline, please file a bug report."), err)
	}
	defer rl.Close()

	fmt.Println("ulp-forth")
	for {
		line, err := rl.ReadSlice()
		if err != nil {
			return err
		}
		err = vm.execute(line)
		if err != nil {
			fmt.Println("")
			return err
		}
		fmt.Println("  ok")
	}
}

// Execute the given bytes.
func (vm *VirtualMachine) execute(bytes []byte) error {
	// trim starting whitespace
	startIndex := -1
	for i, b := range bytes {
		if !isWhitespace(b) {
			startIndex = i
			break
		}
	}
	if startIndex == -1 { // if it's all whitespace
		return nil // then exit early
	}
	// find the end of the word, default to end of the bytes
	endIndex := len(bytes)
	for i := startIndex; i < len(bytes); i++ {
		if isWhitespace(bytes[i]) {
			endIndex = i
			break
		}
	}
	nameBytes := bytes[startIndex:endIndex] // isolate the name
	name := string(nameBytes)               // create a string with it
	err := vm.executeName(name)             // execute that string
	if err != nil {
		return err
	}
	remaining := bytes[endIndex:] // get the remaining bytes
	return vm.execute(remaining)  // and process them
}

func isWhitespace(b byte) bool {
	switch b {
	case ' ', '\r', '\n', '\t':
		return true
	default:
		return false
	}
}

func (vm *VirtualMachine) executeName(name string) error {
	entry, dictErr := vm.Dictionary.FindName(name)
	if dictErr == nil {
		return entry.Word.Execute(vm)
	}
	// dictionary lookup failed, try to parse as a number
	name = strings.ToLower(name)
	double := false
	if strings.HasSuffix(name, ".") {
		double = true
		name = name[:len(name)-1]
	}
	base := 0
	n, err := strconv.ParseInt(name, base, 64)
	if err == nil {
		cell := CellNumber{Number: uint16(n)}
		vm.Stack.Push(cell)
		if double {
			cell = CellNumber{Number: uint16(n >> 16)}
			vm.Stack.Push(cell)
		}
		return nil
	}
	// could not parse as a number, return lookup failure
	return dictErr
}
