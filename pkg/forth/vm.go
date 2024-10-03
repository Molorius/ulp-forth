package forth

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"os"
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

	err = vm.Builtin()
	if err != nil {
		return err
	}

	return nil
}

//go:embed builtin/*.f
var builtins embed.FS

func (vm *VirtualMachine) Builtin() error {
	base := "builtin"
	dirEntries, err := builtins.ReadDir(base)
	if err != nil {
		return errors.Join(fmt.Errorf("Error while opening embedded directory."), err)
	}
	for _, entry := range dirEntries {
		// don't look in subdirectories (for now)
		if !entry.IsDir() {
			name := entry.Name()
			file, err := builtins.Open(base + "/" + name)
			if err != nil {
				return err
			}
			defer file.Close()
			data, err := io.ReadAll(file)
			if err != nil {
				return err
			}
			err = vm.execute(data)
			if err != nil {
				return err
			}
		}
	}
	return nil
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

	fmt.Println("ulp-forth")
	for {
		state, err := vm.State.Get()
		if err != nil {
			return err
		}
		if state == StateExit {
			return nil
		}
		line, err := rl.ReadLineWithConfig(&cfg)
		if err != nil {
			return err
		}
		fmt.Print(" ")
		err = vm.execute([]byte(line))
		if err != nil {
			fmt.Println("")
			return err
		}
		fmt.Println(" ok")
	}
}

// Execute the given bytes.
func (vm *VirtualMachine) execute(bytes []byte) error {
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
		state, err := vm.State.Get()
		if err != nil {
			return err
		}
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
				cellEntry, ok := cell.(CellEntry)
				if ok && cellEntry.Entry.Flag.Immediate {
					err = cellEntry.Execute(vm)
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

func (vm *VirtualMachine) getCells(name string) ([]Cell, error) {
	// check in dictionary for the name
	entry, dictErr := vm.Dictionary.FindName(name)
	if dictErr == nil {
		return []Cell{CellEntry{entry}}, nil
	}
	// dictionary lookup failed, check if this is a character
	nameSlice := []byte(name)
	if len(nameSlice) == 3 && nameSlice[0] == '\'' && nameSlice[2] == '\'' {
		return []Cell{CellNumber{uint16(nameSlice[1])}}, nil
	}
	// not a character, try to parse as a number
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
		if double {
			cellHigh := CellNumber{Number: uint16(n >> 16)}
			return []Cell{cell, cellHigh}, nil
		} else {
			return []Cell{cell}, nil
		}
	}
	// could not parse as a number, return lookup failure
	return nil, dictErr
}
