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
				cellAddr, ok := cell.(CellAddress)
				if !ok {
					return fmt.Errorf("%s requires an address cell.", cellAddr)
				}
				name, err := cellsToString(cellAddr.Entry.Word.(*WordForth).Cells)
				if err != nil {
					return err
				}
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
				str, err := vm.ParseArea.Word(byte(n))
				if err != nil {
					return errors.Join(fmt.Errorf("%s could not parse.", entry), err)
				}
				var de DictionaryEntry
				cells, err := bytesToCells(str, true)
				if err != nil {
					return err
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
				cellAddr, ok := cell.(CellAddress)
				if !ok {
					return fmt.Errorf("%s requires an address cell.", cellAddr)
				}
				name, err := cellsToString(cellAddr.Entry.Word.(*WordForth).Cells)
				if err != nil {
					return err
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
					return errors.Join(fmt.Errorf("%s could not add new dictionary entry.", entry), err)
				}
				return nil
			},
		},
		{
			name: "[",
			flag: Flag{Immediate: true},
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				return vm.State.Set(uint16(StateInterpret))
			},
		},
		{
			name: "]",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				return vm.State.Set(uint16(StateCompile))
			},
		},
		{
			name: "LAST",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				last := vm.Dictionary.Entries[len(vm.Dictionary.Entries)-1]
				c := CellAddress{last, 0, false}
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
				cellAddr, ok := cell.(CellAddress)
				if !ok {
					return fmt.Errorf("%s requires an address cell.", cellAddr)
				}
				name, err := cellsToString(cellAddr.Entry.Word.(*WordForth).Cells)
				if err != nil {
					return err
				}
				found, err := vm.Dictionary.FindName(string(name))
				if err != nil {
					return errors.Join(fmt.Errorf("%s could not find name: %s", entry, name), err)
				}
				newCell := CellAddress{found, 0, false}
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
				cellAddr, ok := c.(CellAddress)
				if !ok {
					return fmt.Errorf("unable to execute cell %s", c)
				}
				return cellAddr.Execute(vm)
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
					return err
				}
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
				switch c := cell.(type) {
				case CellAddress:
					w, ok := c.Entry.Word.(*WordForth)
					if !ok {
						return fmt.Errorf("%s can only read forth data words: %T", entry, w)
					}
					if c.Offset < 0 || c.Offset >= len(w.Cells) {
						return fmt.Errorf("%s reading outside of data range, offset %d", entry, c.Offset)
					}
					return vm.Stack.Push(w.Cells[c.Offset])
				default:
					return fmt.Errorf("%s can only write address or entry cells: %T", entry, cell)
				}
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
				n, err := vm.Stack.Pop()
				if err != nil {
					return err
				}
				switch c := cell.(type) {
				case CellAddress:
					w, ok := c.Entry.Word.(*WordForth)
					if !ok {
						return fmt.Errorf("%s can only write forth data words: %T", entry, w)
					}
					if c.Offset < 0 || c.Offset >= len(w.Cells) {
						return fmt.Errorf("%s writing outside of data range, offset %d", entry, c.Offset)
					}
					w.Cells[c.Offset] = n
					return nil
				default:
					return fmt.Errorf("%s can only write address or entry cells: %T", entry, cell)
				}
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
				cellAddr, ok := cell.(CellAddress)
				if !ok {
					return fmt.Errorf("%s requires an entry cell: %s", entry, cell)
				}
				last, err := vm.Dictionary.LastForthWord()
				if err != nil {
					return errors.Join(fmt.Errorf("%s could not get last forth word.", entry), err)
				}
				if cellAddr.Entry.Flag.Immediate {
					last.Cells = append(last.Cells, cellAddr)
					return nil
				} else {
					compile, err := vm.Dictionary.FindName("COMPILE,")
					if !ok {
						return errors.Join(fmt.Errorf("%s requires the COMPILE, word.", entry), err)
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
					return errors.Join(fmt.Errorf("%s could not get address cell.", entry), err)
				}
				cellAddr, ok := cell0.(CellAddress)
				if !ok {
					return fmt.Errorf("%s expected an address cell: %s", entry, cellAddr)
				}
				cellNum, err := vm.Stack.PopNumber()
				if err != nil {
					return errors.Join(fmt.Errorf("%s could not get bool cell.", entry), err)
				}
				flag := cellNum > 0
				cellAddr.Entry.Flag.Hidden = flag
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
				cellAddr, ok := cell0.(CellAddress)
				if !ok {
					return fmt.Errorf("%s expected an entry cell: %s", entry, cellAddr)
				}
				cellNum, err := vm.Stack.PopNumber()
				if err != nil {
					return errors.Join(fmt.Errorf("%s could not get bool cell.", entry), err)
				}
				flag := cellNum > 0
				cellAddr.Entry.Flag.Immediate = flag
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
				right, err := vm.Stack.Pop()
				if err != nil {
					return err
				}
				left, err := vm.Stack.Pop()
				if err != nil {
					return err
				}
				switch r := right.(type) {
				case CellNumber:
					switch l := left.(type) {
					case CellNumber:
						return vm.Stack.Push(CellNumber{l.Number + r.Number})
					case CellAddress:
						return vm.Stack.Push(CellAddress{l.Entry, l.Offset + int(r.Number), false})
					default:
						return fmt.Errorf("Cannot add these types")
					}
				case CellAddress:
					switch l := left.(type) {
					case CellNumber:
						return vm.Stack.Push(CellAddress{r.Entry, int(l.Number) + r.Offset, false})
					default:
						return fmt.Errorf("Cannot add these types")
					}
				default:
					return fmt.Errorf("Cannot add these types")
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
		},
		{
			name: "-",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				right, err := vm.Stack.Pop()
				if err != nil {
					return err
				}
				left, err := vm.Stack.Pop()
				if err != nil {
					return err
				}
				switch r := right.(type) {
				case CellNumber:
					switch l := left.(type) {
					case CellNumber:
						return vm.Stack.Push(CellNumber{l.Number - r.Number})
					case CellAddress:
						return vm.Stack.Push(CellAddress{l.Entry, l.Offset - int(r.Number), false})
					default:
						return fmt.Errorf("%s cannot add these types", entry)
					}
				case CellAddress:
					switch l := left.(type) {
					case CellAddress:
						if l.Entry == r.Entry {
							return vm.Stack.Push(CellNumber{uint16(l.Offset) - uint16(r.Offset)})
						}
						return fmt.Errorf("%s cannot subtract addresses from different words.", entry)
					default:
						return fmt.Errorf("%s cannot subtract these types", entry)
					}
				default:
					return fmt.Errorf("%s cannot subtract these types", entry)
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
		},
		{
			name: "AND",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				right, err := vm.Stack.PopNumber()
				if err != nil {
					return errors.Join(fmt.Errorf("%s could not get right value.", entry), err)
				}
				left, err := vm.Stack.PopNumber()
				if err != nil {
					return errors.Join(fmt.Errorf("%s could not get left value.", entry), err)
				}
				return vm.Stack.Push(CellNumber{left & right})
			},
			ulpAsm: PrimitiveUlp{
				"ld r0, r3, 1",
				"ld r1, r3, 0",
				"and r0, r0, r1",
				"add r3, r3, 1",
				"st r0, r3, 0",
				"jump __next_skip_r2",
			},
		},
		{
			name: "OR",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				right, err := vm.Stack.PopNumber()
				if err != nil {
					return errors.Join(fmt.Errorf("%s could not get right value.", entry), err)
				}
				left, err := vm.Stack.PopNumber()
				if err != nil {
					return errors.Join(fmt.Errorf("%s could not get left value.", entry), err)
				}
				return vm.Stack.Push(CellNumber{left | right})
			},
			ulpAsm: PrimitiveUlp{
				"ld r0, r3, 1",
				"ld r1, r3, 0",
				"or r0, r0, r1",
				"add r3, r3, 1",
				"st r0, r3, 0",
				"jump __next_skip_r2",
			},
		},
		{
			name: "*",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				right, err := vm.Stack.PopNumber()
				if err != nil {
					return errors.Join(fmt.Errorf("%s could not get right value.", entry), err)
				}
				left, err := vm.Stack.PopNumber()
				if err != nil {
					return errors.Join(fmt.Errorf("%s could not get left value.", entry), err)
				}
				return vm.Stack.Push(CellNumber{left * right})
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
				"jumpr __mult.0, 0, gt", // loop if y > 0
				// finalize
				"add r3, r3, 1", // decrement stack, z already in place
				"jump next",
			},
		},
		{
			name: "/MOD",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				right, err := vm.Stack.PopNumber()
				if err != nil {
					return errors.Join(fmt.Errorf("%s could not get right value.", entry), err)
				}
				left, err := vm.Stack.PopNumber()
				if err != nil {
					return errors.Join(fmt.Errorf("%s could not get left value.", entry), err)
				}
				quotient := left / right
				remainder := left % right
				err = vm.Stack.Push(CellNumber{remainder})
				if err != nil {
					return err
				}
				return vm.Stack.Push(CellNumber{quotient})
			},
			ulpAsm: PrimitiveUlp{
				// 'd' on 0
				// 'n' on 1
				// 'r' on r1
				// 'q' on r2
				// loop on stage_cnt

				"move r1, 0", // r = 0, r2 already set to 0
				"stage_rst",  // stage_cnt = 0

				"__divmod.0:",
				// shift n into r, shift r, shift q
				"lsh r1, r1, 1",                // r = r<<1
				"lsh r2, r2, 1",                // q = q<<1
				"ld r0, r3, 1",                 // load n
				"jumpr __divmod.1, 0x8000, lt", // jump if the top bit is not set
				"or r1, r1, 1",                 // if bit is set, "shift" this bit into r
				"__divmod.1:",                  // then
				"lsh r0, r0, 1",                // n = n<<1
				"st r0, r3, 1",                 // store n
				// attempt subtracting
				"ld r0, r3, 0",        // load d
				"sub r0, r1, r0",      // r0 = r - d
				"jump __divmod.2, ov", // jump ahead if that overflowed
				// no overflow
				"move r1, r0",  // store result into r
				"or r2, r2, 1", // set the lowest bit of q
				"__divmod.2:",
				"stage_inc 1",              // increase the stage counter
				"jumps __divmod.0, 16, lt", // loop over each bit

				// done! store r and q
				"st r1, r3, 1", // r
				"st r2, r3, 0", // q
				"jump next",
			},
		},
		{
			name: "LSHIFT",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				amount, err := vm.Stack.PopNumber()
				if err != nil {
					return err
				}
				num, err := vm.Stack.PopNumber()
				if err != nil {
					return err
				}
				return vm.Stack.Push(CellNumber{num << amount})
			},
			ulpAsm: PrimitiveUlp{
				"ld r0, r3, 1",
				"ld r1, r3, 0",
				"lsh r0, r0, r1",
				"add r3, r3, 1",
				"st r0, r3, 0",
				"jump __next_skip_r2",
			},
		},
		{
			name: "RSHIFT",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				amount, err := vm.Stack.PopNumber()
				if err != nil {
					return err
				}
				num, err := vm.Stack.PopNumber()
				if err != nil {
					return err
				}
				return vm.Stack.Push(CellNumber{num >> amount})
			},
			ulpAsm: PrimitiveUlp{
				"ld r0, r3, 1",
				"ld r1, r3, 0",
				"rsh r0, r0, r1",
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
			name: "U<",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				right, err := vm.Stack.PopNumber()
				if err != nil {
					return errors.Join(fmt.Errorf("%s could not get right value.", entry), err)
				}
				left, err := vm.Stack.PopNumber()
				if err != nil {
					return errors.Join(fmt.Errorf("%s could not get left value.", entry), err)
				}
				val := 0
				if left < right {
					val = 0xFFFF
				}
				return vm.Stack.Push(CellNumber{uint16(val)})
			},
			ulpAsm: PrimitiveUlp{
				"ld r2, r3, 1",            // left
				"ld r1, r3, 0",            // right
				"move r0, 0xFFFF",         // default to true
				"sub r2, r2, r1",          // subtract
				"jump __u_lessthan.0, ov", // jump if overflow
				"move r0, 0",              // no overflow: false
				"__u_lessthan.0:",         // then
				"add r3, r3, 1",           // decrement stack
				"st r0, r3, 0",            // store the result
				"jump next",
			},
		},
		{
			name: "BYE",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				return vm.State.Set(uint16(StateExit))
			},
		},
		{
			name: "DEPTH",
			goFunc: func(vm *VirtualMachine, entry *DictionaryEntry) error {
				depth := len(vm.Stack.stack)
				cell := CellNumber{uint16(depth)}
				return vm.Stack.Push(cell)
			},
			ulpAsm: PrimitiveUlp{
				"move r0, __stack_end",
				"sub r0, r0, r3",
				"sub r3, r3, 1",
				"st r0, r3, 0",
				"jump __next_skip_r2",
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
