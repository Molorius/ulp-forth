package forth

import (
	"errors"
	"fmt"
)

type primitive struct {
	name   string
	goFunc PrimitiveGo
	ulpAsm PrimitiveUlp
	flag   Flag
}

func primitiveAdd(vm *VirtualMachine, name string, goFunc PrimitiveGo, ulpAsm PrimitiveUlp, flag Flag) error {
	var entry DictionaryEntry
	entry = DictionaryEntry{
		Name: name,
		Word: &WordPrimitive{
			Go:    goFunc,
			Ulp:   ulpAsm,
			Entry: &entry,
		},
		Flag: flag,
	}
	return vm.Dictionary.AddEntry(&entry)
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
					return errors.Join(fmt.Errorf("%s could not pop name.", entry), err)
				}
				cellString, ok := cell.(CellString)
				if !ok {
					return fmt.Errorf("%s requires a counted string.", entry)
				}
				name := cellString.Memory[cellString.Offset:]
				wordEntry, err := vm.Dictionary.FindName(string(name))
				if err != nil {
					return errors.Join(fmt.Errorf("%s could not find word.", entry), err)
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
					return errors.Join(fmt.Errorf("%s could not pop delimiter.", entry), err)
				}
				word, err := vm.ParseArea.Word(byte(n))
				if err != nil {
					return errors.Join(fmt.Errorf("%s could not parse.", entry), err)
				}
				c := CellString{
					Memory: word,
					Offset: 0,
				}
				return vm.Stack.Push(c)
			},
		},
		{
			name: "--CREATE-FORTH",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				cell, err := vm.Stack.Pop()
				if err != nil {
					return errors.Join(fmt.Errorf("%s could not pop name.", entry), err)
				}
				cellString, ok := cell.(CellString)
				if !ok {
					return fmt.Errorf("%s requires a counted string.", entry)
				}
				name := cellString.Memory[cellString.Offset:]
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
					return errors.Join(fmt.Errorf("%s could not add new dictionary entry.", entry), err)
				}
				return nil
			},
		},
		{
			name: "[",
			flag: Flag{Immediate: true},
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				return vm.State.Set(StateInterpret)
			},
		},
		{
			name: "]",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				return vm.State.Set(StateCompile)
			},
		},
		{
			name: "LAST",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				last := vm.Dictionary.Entries[len(vm.Dictionary.Entries)-1]
				c := CellEntry{last}
				return vm.Stack.Push(c)
			},
		},
		{
			name: ">C",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				c, err := vm.Stack.Pop()
				if err != nil {
					return err
				}
				return vm.ControlFlowStack.Push(c)
			},
		},
		{
			name: "C>",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				c, err := vm.ControlFlowStack.Pop()
				if err != nil {
					return err
				}
				return vm.Stack.Push(c)
			},
		},
		{
			name: ">R",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				c, err := vm.Stack.Pop()
				if err != nil {
					return err
				}
				return vm.ReturnStack.Push(c)
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
		},
		{
			name: "R>",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				c, err := vm.ReturnStack.Pop()
				if err != nil {
					return err
				}
				return vm.Stack.Push(c)
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
		},
		{
			name: "BRANCH",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				return vm.Stack.Push(&CellBranch{})
			},
		},
		{
			name: "BRANCH0",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				return vm.Stack.Push(&CellBranch0{})
			},
		},
		{
			name: "DEST",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				return vm.Stack.Push(&CellDestination{})
			},
		},
		{
			name: "RESOLVE-BRANCH",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				destCell, err := vm.Stack.Pop()
				if err != nil {
					return err
				}
				branchCell, err := vm.Stack.Pop()
				if err != nil {
					return err
				}
				dest, ok := destCell.(*CellDestination)
				if !ok {
					return fmt.Errorf("%s expected a destination: %s", entry, dest)
				}
				switch b := branchCell.(type) {
				case *CellBranch:
					b.dest = dest
				case *CellBranch0:
					b.dest = dest
				default:
					return fmt.Errorf("%s expected a branch: %s", entry, branchCell)
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
					return errors.Join(fmt.Errorf("%s could not pop from stack.", entry), err)
				}
				newCell := CellLiteral{c}
				last, err := vm.Dictionary.LastForthWord()
				if err != nil {
					return errors.Join(fmt.Errorf("%s could not get last forth word.", entry), err)
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
					return errors.Join(fmt.Errorf("%s could not pop from stack.", entry), err)
				}
				cellString, ok := cell.(CellString)
				if !ok {
					return fmt.Errorf("%s requires a counted string.", entry)
				}
				name := cellString.Memory[cellString.Offset:]
				found, err := vm.Dictionary.FindName(string(name))
				if err != nil {
					return errors.Join(fmt.Errorf("%s could not find name: %s", entry, name), err)
				}
				newCell := CellEntry{found}
				return vm.Stack.Push(newCell)
			},
		},
		{
			name: "COMPILE,",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				c, err := vm.Stack.Pop()
				if err != nil {
					return errors.Join(fmt.Errorf("%s could not pop from stack.", entry), err)
				}
				last, err := vm.Dictionary.LastForthWord()
				if err != nil {
					return errors.Join(fmt.Errorf("%s could not get last forth word.", entry), err)
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
					return err
				}
				cellEntry, ok := c.(CellEntry)
				if !ok {
					return fmt.Errorf("unable to execute cell %s", c)
				}
				return cellEntry.Execute(vm)
			},
			ulpAsm: PrimitiveUlp{
				"ld r0, r3, 0",   // load the token into r0
				"add r3, r3, 1",  // decrement stack pointer
				"jump __ins_asm", // start execution of the token
			},
		},
		{
			name: "ALLOCATE",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				n, err := vm.Stack.PopNumber()
				if err != nil {
					return err
				}
				cell := CellData{
					Data: &Data{
						Cells: make([]Cell, n),
					},
					Offset: 0,
				}
				// set all values to 0 at start
				for i := range cell.Data.Cells {
					cell.Data.Cells[i] = CellNumber{0}
				}
				err = vm.Stack.Push(cell)
				if err != nil {
					return err
				}
				// also push 0 because allocation worked
				return vm.Stack.Push(CellNumber{0})
			},
		},
		{
			name: "@",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				cell, err := vm.Stack.Pop()
				if err != nil {
					return err
				}
				data, ok := cell.(CellData)
				if !ok {
					return fmt.Errorf("%s can only read data cells: %T", entry, data)
				}
				if data.Offset >= len(data.Data.Cells) {
					return fmt.Errorf("%s reading outside of data range, offset %d", entry, data.Offset)
				}
				return vm.Stack.Push(data.Data.Cells[data.Offset])
			},
			ulpAsm: PrimitiveUlp{
				"ld r0, r3, 0", // get the address from stack
				"ld r0, r0, 0", // load the value
				"st r0, r3, 0", // store the address on stack
				"jump __next_skip_r2",
			},
		},
		{
			name: "!",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				cell, err := vm.Stack.Pop()
				if err != nil {
					return err
				}
				data, ok := cell.(CellData)
				if !ok {
					return fmt.Errorf("%s can only write data cells: %T", entry, data)
				}
				if data.Offset >= len(data.Data.Cells) {
					return fmt.Errorf("%s writing outside of data range, offset %d", entry, data.Offset)
				}
				n, err := vm.Stack.Pop()
				if err != nil {
					return err
				}
				data.Data.Cells[data.Offset] = n
				return nil
			},
			ulpAsm: PrimitiveUlp{
				"ld r0, r3, 0",  // load the address
				"ld r1, r3, 1",  // load the value
				"st r1, r0, 0",  // store the value in address
				"add r3, r3, 2", // decrement the stack
				"jump __next_skip_r2",
			},
		},
		{
			name: "--POSTPONE",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				cell, err := vm.Stack.Pop()
				if err != nil {
					return errors.Join(fmt.Errorf("%s could not pop from stack.", entry), err)
				}
				cellEntry, ok := cell.(CellEntry)
				if !ok {
					return fmt.Errorf("%s requires an entry cell: %s", entry, cell)
				}
				last, err := vm.Dictionary.LastForthWord()
				if err != nil {
					return errors.Join(fmt.Errorf("%s could not get last forth word.", entry), err)
				}
				if cellEntry.Entry.Flag.Immediate {
					last.Cells = append(last.Cells, cellEntry)
					return nil
				} else {
					compile, err := vm.Dictionary.FindName("COMPILE,")
					if !ok {
						return errors.Join(fmt.Errorf("%s requires the COMPILE, word.", entry), err)
					}
					newCells := []Cell{CellLiteral{cellEntry}, CellEntry{compile}}
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
					return errors.Join(fmt.Errorf("%s could not get entry cell.", entry), err)
				}
				cellEntry, ok := cell0.(CellEntry)
				if !ok {
					return fmt.Errorf("%s expected an entry cell: %s", entry, cellEntry)
				}
				cellNum, err := vm.Stack.PopNumber()
				if err != nil {
					return errors.Join(fmt.Errorf("%s could not get bool cell.", entry), err)
				}
				flag := cellNum > 0
				cellEntry.Entry.Flag.Hidden = flag
				return nil
			},
		},
		{
			name: "SET-IMMEDIATE",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				cell0, err := vm.Stack.Pop()
				if err != nil {
					return errors.Join(fmt.Errorf("%s could not get entry cell.", entry), err)
				}
				cellEntry, ok := cell0.(CellEntry)
				if !ok {
					return fmt.Errorf("%s expected an entry cell: %s", entry, cellEntry)
				}
				cellNum, err := vm.Stack.PopNumber()
				if err != nil {
					return errors.Join(fmt.Errorf("%s could not get bool cell.", entry), err)
				}
				flag := cellNum > 0
				cellEntry.Entry.Flag.Immediate = flag
				return nil
			},
		},
		{
			name: "EXIT",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				ret, err := vm.ReturnStack.Pop()
				if err != nil {
					return errors.Join(fmt.Errorf("%s could not pop return address.", entry), err)
				}
				addr, ok := ret.(*CellAddress)
				if !ok {
					return fmt.Errorf("%s expected a return address on the return stack.: %v", entry, ret)
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
		},
		{
			name: "+",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				right, err := vm.Stack.PopNumber()
				if err != nil {
					return errors.Join(fmt.Errorf("%s could not get right value.", entry), err)
				}
				left, err := vm.Stack.PopNumber()
				if err != nil {
					return errors.Join(fmt.Errorf("%s could not get left value.", entry), err)
				}
				return vm.Stack.Push(CellNumber{left + right})
			},
			ulpAsm: PrimitiveUlp{
				"ld r0, r3, 0",
				"ld r1, r3, 1",
				"add r0, r0, r1",
				"add r3, r3, 1",
				"st r0, r3, 0",
				"jump __next_skip_r2",
			},
		},
		{
			name: "-",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				right, err := vm.Stack.PopNumber()
				if err != nil {
					return errors.Join(fmt.Errorf("%s could not get right value.", entry), err)
				}
				left, err := vm.Stack.PopNumber()
				if err != nil {
					return errors.Join(fmt.Errorf("%s could not get left value.", entry), err)
				}
				return vm.Stack.Push(CellNumber{left - right})
			},
			ulpAsm: PrimitiveUlp{
				"ld r0, r3, 1",
				"ld r1, r3, 0",
				"sub r0, r0, r1",
				"add r3, r3, 1",
				"st r0, r3, 0",
				"jump __next_skip_r2",
			},
		},
		{
			name: "SWAP",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				right, err := vm.Stack.Pop()
				if err != nil {
					return errors.Join(fmt.Errorf("%s could not get right value.", entry), err)
				}
				left, err := vm.Stack.Pop()
				if err != nil {
					return errors.Join(fmt.Errorf("%s could not get left value.", entry), err)
				}
				err = vm.Stack.Push(right)
				if err != nil {
					return err
				}
				return vm.Stack.Push(left)
			},
			ulpAsm: PrimitiveUlp{
				"ld r1, r3, 0",
				"ld r0, r3, 1",
				"st r1, r3, 1",
				"st r0, r3, 0",
				"jump __next_skip_r2",
			},
		},
		{
			name: "DUP",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				c, err := vm.Stack.Pop()
				if err != nil {
					return err
				}
				err = vm.Stack.Push(c)
				if err != nil {
					return err
				}
				return vm.Stack.Push(c)
			},
			ulpAsm: PrimitiveUlp{
				"ld r0, r3, 0",
				"sub r3, r3, 1",
				"st r0, r3, 0",
				"jump __next_skip_r2",
			},
		},
		{
			name: "PICK",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				n, err := vm.Stack.PopNumber()
				if err != nil {
					return err
				}
				if int(n) >= len(vm.Stack.stack) {
					return fmt.Errorf("%s number out of range: %d", entry, n)
				}
				index := len(vm.Stack.stack) - int(n) - 1
				cell := vm.Stack.stack[index]
				return vm.Stack.Push(cell)
			},
			ulpAsm: PrimitiveUlp{
				"ld r0, r3, 0",
				"add r0, r0, r3",
				"ld r0, r0, 1",
				"st r0, r3, 0",
				"jump __next_skip_r2",
			},
		},
		{
			name: "ROT",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				c, err := vm.Stack.Pop()
				if err != nil {
					return errors.Join(fmt.Errorf("%s could not get c value.", entry), err)
				}
				b, err := vm.Stack.Pop()
				if err != nil {
					return errors.Join(fmt.Errorf("%s could not get b value.", entry), err)
				}
				a, err := vm.Stack.Pop()
				if err != nil {
					return errors.Join(fmt.Errorf("%s could not get a value.", entry), err)
				}
				err = vm.Stack.Push(b)
				if err != nil {
					return err
				}
				err = vm.Stack.Push(c)
				return vm.Stack.Push(a)
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
		},
		{
			name: "DROP",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				_, err := vm.Stack.Pop()
				if err != nil {
					return errors.Join(fmt.Errorf("%s could not pop from stack.", entry), err)
				}
				return nil
			},
			ulpAsm: PrimitiveUlp{
				"add r3, r3, 1",
				"jump __next_skip_r2",
			},
		},
		{
			name: "BYE",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				return vm.State.Set(StateExit)
			},
		},
		{
			name:   "VM.STACK.INIT", // initialize the ulp stack
			goFunc: notImplemented,
			ulpAsm: PrimitiveUlp{
				"move r3, __stack_end", // set the stack pointer to the end of the stack
				"jump __next_skip_r2",
			},
		},
		{
			name:   "VM.STOP", // stop executing
			goFunc: notImplemented,
			ulpAsm: PrimitiveUlp{
				"jump .", // loop indefinitely
			},
		},
		{
			name: "ESP.FUNC.UNSAFE", // use one of the custom esp32/host functions
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				funcType, err := vm.Stack.PopNumber()
				if err != nil {
					return errors.Join(fmt.Errorf("%s could not get function type from stack.", entry), err)
				}
				cell, err := vm.Stack.Pop()
				if err != nil {
					return errors.Join(fmt.Errorf("%s could not pop from stack.", entry), err)
				}
				switch funcType {
				case 0: // nothing
				case 1: // done
				case 2: // print unsigned number (we're not quite accurate here for convenience)
					fmt.Fprint(vm.Out, cell, " ")
				case 3: // print char
					num, ok := cell.(CellNumber)
					if !ok {
						return fmt.Errorf("%s expected a number: %s", entry, num)
					}
					c := byte(num.Number)
					fmt.Fprintf(vm.Out, "%c", c)
				default:
					return fmt.Errorf("%s unknown function type %d", entry, funcType)
				}
				return nil
			},
			ulpAsm: PrimitiveUlp{
				"ld r0, r3, 1",                   // load the value we want to print
				"ld r1, r3, 0",                   // load the method number
				"st r0, r2, __boot_data_start+4", // set the value
				"st r1, r2, __boot_data_start+3", // set the method indicator
				"add r3, r3, 2",                  // decrease the stack by 2
				"jump __next_skip_r2",
			},
		},
		{
			name: "ESP.FUNC.READ.UNSAFE",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				return vm.Stack.Push(CellNumber{0})
			},
			ulpAsm: PrimitiveUlp{
				"ld r0, r2, __boot_data_start+3",
				"sub r3, r3, 1",
				"st r0, r3, 0",
				"jump __next_skip_r2",
			},
		},
		{
			name:   "MUTEX.TAKE",
			goFunc: nop,
			ulpAsm: PrimitiveUlp{
				"move r1, 1",
				"st r1, r2, __boot_data_start",   // flag0 = 1
				"st r1, r2, __boot_data_start+2", // turn = 1
				"mutex.take.0:",                  // //while flag1>0 && turn>0
				"ld r0, r2, __boot_data_start+1", // read flag1
				"jumpr mutex.take.1, 1, lt",      // exit if flag1<1
				"ld r0, r2, __boot_data_start+2", // read turn
				"jumpr mutex.take.0, 0, gt",      // loop if turn>0
				"mutex.take.1:",
				"jump __next_skip_r2",
			},
		},
		{
			name:   "MUTEX.GIVE",
			goFunc: nop,
			ulpAsm: PrimitiveUlp{
				"st r2, r2, __boot_data_start", // flag0 = 0
				"jump __next_skip_r2",
			},
		},
		{
			name:   "DEBUG.PAUSE", // temporarily used while we don't have jumps
			goFunc: nop,
			ulpAsm: PrimitiveUlp{
				"wait 0xFFFF",
				"jump __next_skip_r2",
			},
		},
	}
	for _, p := range prims {
		err := primitiveAdd(vm, p.name, p.goFunc, p.ulpAsm, p.flag)
		if err != nil {
			return err
		}
	}
	return nil
}

func notImplemented(vm *VirtualMachine, entry *DictionaryEntry) error {
	return fmt.Errorf("%s cannot be executed on the host.", entry)
}

func nop(vm *VirtualMachine, entry *DictionaryEntry) error {
	return nil
}
