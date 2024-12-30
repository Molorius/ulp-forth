/*
Copyright 2024 Blake Felt blake.w.felt@gmail.com

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/
package forth

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strconv"
	"strings"

	"github.com/ergochat/readline"
)

// The Forth virtual machine.
type VirtualMachine struct {
	Dictionary       Dictionary   // The Forth dictionary.
	Stack            Stack        // The data stack.
	ReturnStack      Stack        // The return stack.
	ControlFlowStack Stack        // The control flow stack.
	DoStack          Stack        // A stack for compiling DO loops.
	ParseArea        ParseArea    // The input parse area.
	State            VMNumber     // The execution state for the virtual machine. Convert to type State when using.
	IP               *CellAddress // The interpreter pointer.
	Base             VMNumber     // The number base.
	Out              io.Writer
}

// Set up the virtual machine.
func (vm *VirtualMachine) Setup() error {
	if vm.Out == nil {
		vm.Out = os.Stdout
	}

	err := vm.Dictionary.Setup(vm)
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

	err = vm.ControlFlowStack.Setup()
	if err != nil {
		return err
	}
	err = vm.DoStack.Setup()
	if err != nil {
		return err
	}

	err = vm.ParseArea.Setup()
	if err != nil {
		return err
	}

	vm.IP = nil

	err = PrimitiveSetup(vm)
	if err != nil {
		return err
	}

	err = vm.Base.Setup(vm, "BASE", true)
	if err != nil {
		return err
	}
	vm.Base.Set(10)

	err = vm.State.Setup(vm, "STATE", false) // no need to share state as it shouldn't change
	if err != nil {
		return err
	}

	err = vm.builtin()
	if err != nil {
		return err
	}

	return nil
}

//go:embed builtin/*.f
var builtins embed.FS

//go:embed esp32/*.f
var builtinsEsp32 embed.FS

func (vm *VirtualMachine) buildEmbed(f embed.FS, name string) error {
	dirEntries, err := f.ReadDir(name)
	if err != nil {
		return errors.Join(fmt.Errorf("Error while opening embedded directory."), err)
	}
	for _, entry := range dirEntries {
		// don't look in subdirectories (for now)
		if !entry.IsDir() {
			file, err := f.Open(name + "/" + entry.Name())
			if err != nil {
				return err
			}
			defer file.Close()
			err = vm.ExecuteFile(file)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// builtin runs the files necessary for the Forth environment.
func (vm *VirtualMachine) builtin() error {
	return vm.buildEmbed(builtins, "builtin")
}

// BuiltinEsp32 runs the files intended for accessing esp32 hardware.
func (vm *VirtualMachine) BuiltinEsp32() error {
	return vm.buildEmbed(builtinsEsp32, "esp32")
}

// writerNoNewline is an io.Writer that does not print newlines.
// Forth implementations usually don't have a newline immediately after
// user input, this lets us do the same.
type writerNoNewline struct {
}

func (w writerNoNewline) Write(p []byte) (int, error) {
	pCopy := make([]byte, 0, len(p)) // create a new slice length 0, capacity same as input
	newlines := 0                    // number of newlines

	for _, c := range p {
		switch c {
		case '\n':
			newlines += 1 // add to the list
		default:
			pCopy = pCopy[:len(pCopy)+1] // extend the length of our slice
			pCopy[len(pCopy)-1] = c      // put in the character
		}
	}
	n, err := os.Stdout.Write(pCopy) // write to stdout
	return n + newlines, err         // tell the higher function that we printed the newlines
}

// Start up a read-eval-print loop.
func (vm *VirtualMachine) Repl() error {
	rl, err := readline.New("")
	cfg := readline.Config{
		Stdout: writerNoNewline{},
	}
	if err != nil {
		return errors.Join(fmt.Errorf("Unable to start readline, please file a bug report."), err)
	}
	defer rl.Close()

	for {
		stateUint, err := vm.State.Get()
		if err != nil {
			return err
		}
		state := StateType(stateUint)
		if state == StateExit {
			return nil
		}
		line, err := rl.ReadLineWithConfig(&cfg)
		if err != nil {
			return err
		}
		fmt.Fprint(vm.Out, " ")
		err = vm.Execute([]byte(line))
		if err != nil {
			fmt.Fprintln(vm.Out)
			return err
		}
		fmt.Fprintln(vm.Out, " ok")
	}
}

// Execute the given bytes.
func (vm *VirtualMachine) Execute(bytes []byte) error {
	err := vm.ParseArea.Fill(bytes)
	if err != nil {
		return err
	}
	for {
		word, err := vm.ParseArea.Word(' ')
		if err != nil {
			return err
		}
		if len(word) == 0 {
			return nil
		}
		cells, err := vm.getCells(string(word))
		if err != nil {
			return err
		}
		stateUint, err := vm.State.Get()
		if err != nil {
			return err
		}
		state := StateType(stateUint)
		switch state {
		case StateInterpret:
			for _, c := range cells {
				err := c.Execute(vm)
				if err != nil {
					return err
				}
			}
		case StateCompile:
			for _, cell := range cells {
				cellAddress, ok := cell.(CellAddress)
				if ok && cellAddress.Entry.Flag.Immediate {
					err = cellAddress.Execute(vm)
					if err != nil {
						return err
					}
				} else {
					last, err := vm.Dictionary.LastForthWord()
					if err != nil {
						return err
					}
					last.Cells = append(last.Cells, cell)
				}
			}
		case StateExit:
			return nil // the repl will exit after reading the state
		default:
			return fmt.Errorf("Unknown state %d", state)
		}
	}
}

func (vm *VirtualMachine) ExecuteFile(f fs.File) error {
	data, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	return vm.Execute(data)
}

func (vm *VirtualMachine) Reset() error {
	// reset all stacks
	vm.Stack.Reset()
	vm.ReturnStack.Reset()
	vm.ControlFlowStack.Reset()
	vm.DoStack.Reset()
	// and the instruction pointer
	vm.IP = nil
	return nil
}

func (vm *VirtualMachine) getCells(name string) ([]Cell, error) {
	// check in dictionary for the name
	entry, dictErr := vm.Dictionary.FindName(name)
	if dictErr == nil {
		return []Cell{CellAddress{entry, 0, false}}, nil
	}
	// dictionary lookup failed, check if this is a character
	nameSlice := []byte(name)
	if len(nameSlice) == 3 && nameSlice[0] == '\'' && nameSlice[2] == '\'' {
		return []Cell{CellLiteral{CellNumber{uint16(nameSlice[1])}}}, nil
	}
	// not a character, try to parse as a number
	name = strings.ToLower(name)
	double := false
	if strings.HasSuffix(name, ".") {
		double = true
		name = name[:len(name)-1]
	}
	isNegative := strings.HasPrefix(name, "-")
	if isNegative {
		name = name[1:]
	}
	baseuint, err := vm.Base.Get()
	if err != nil {
		return nil, err
	}
	base := int(baseuint)
	if strings.HasPrefix(name, "0x") {
		base = 16
		name = name[2:]
	} else if strings.HasPrefix(name, "0b") {
		base = 2
		name = name[2:]
	} else if strings.HasPrefix(name, "#") {
		base = 10
		name = name[1:]
	}
	n, err := strconv.ParseInt(name, base, 64)
	if err == nil {
		if isNegative {
			n = n * -1
		}
		cell := CellLiteral{CellNumber{Number: uint16(n)}}
		if double {
			cellHigh := CellLiteral{CellNumber{Number: uint16(n >> 16)}}
			return []Cell{cell, cellHigh}, nil
		} else {
			return []Cell{cell}, nil
		}
	}
	// could not parse as a number, return lookup failure
	return nil, dictErr
}

type VMNumber struct {
	Word WordForth
}

func (n *VMNumber) Setup(vm *VirtualMachine, name string, shared bool) error {
	// first create the allocated memory
	var allocEntry DictionaryEntry
	n.Word = WordForth{
		Cells: make([]Cell, 1),
		Entry: &allocEntry,
	}
	n.Word.Cells[0] = CellNumber{0} // initialize to 0
	allocEntry = DictionaryEntry{
		Name: name,
		Word: &n.Word,
		Flag: Flag{Data: true},
	}
	// then create the actual dictionary entry
	docol, err := vm.Dictionary.FindName("DOCOL")
	if err != nil {
		return err
	}
	exit, err := vm.Dictionary.FindName("EXIT")
	if err != nil {
		return err
	}
	var dEntry DictionaryEntry
	dWord := WordForth{
		Cells: make([]Cell, 3),
		Entry: &dEntry,
	}
	dWord.Cells[0] = CellAddress{docol, 0, false}                    // run the DOCOL
	dWord.Cells[1] = CellLiteral{CellAddress{&allocEntry, 0, false}} // put address on stack
	dWord.Cells[2] = CellAddress{exit, 0, false}                     // then exit
	dEntry = DictionaryEntry{
		Name: name,
		Word: &dWord,
	}
	// add the actual entry to the dictionary
	vm.Dictionary.Entries = append(vm.Dictionary.Entries, &dEntry)
	return nil
}

func (n *VMNumber) Get() (uint16, error) {
	c := n.Word.Cells[0]
	num := c.(CellNumber) // instant fail if this isn't a number
	return num.Number, nil
}

func (n *VMNumber) Set(num uint16) error {
	n.Word.Cells[0] = CellNumber{num}
	return nil
}
