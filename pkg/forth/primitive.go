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
			name:   ".S",
			goFunc: primitiveFuncDotS,
		},
		{
			name:   "WORDS",
			goFunc: primitiveFuncWords,
		},
		{
			name:   "--SEE", // ( )
			goFunc: primitiveFuncSee,
		},
		{
			name:   "WORD",
			goFunc: primitiveFuncWord,
		},
		{
			name:   "--CREATE-FORTH",
			goFunc: primitiveFuncCreateForth,
		},
		{
			name:   "[",
			goFunc: primitiveFuncLeftBracket,
			flag:   Flag{Immediate: true},
		},
		{
			name:   "]",
			goFunc: primitiveFuncRightBracket,
		},
		{
			name:   "LAST",
			goFunc: primitiveFuncLast,
		},
		{
			name:   "LITERAL",
			goFunc: primitiveFuncLiteral,
			flag:   Flag{Immediate: true},
		},
		{
			name:   "FIND-WORD",
			goFunc: primitiveFindWord,
		},
		{
			name:   "COMPILE,",
			goFunc: primitiveFuncCompile,
		},
		{
			name:   "--POSTPONE",
			goFunc: primitiveFuncPostpone,
		},
		{
			name:   "SET-HIDDEN",
			goFunc: primitiveSetHidden,
		},
		{
			name:   "SET-IMMEDIATE",
			goFunc: primitiveSetImmediate,
		},
		{
			name:   "EXIT",
			goFunc: primitiveFuncExit,
			ulpAsm: PrimitiveUlp{
				"ld r0, r2, __rsp",      // load the return stack pointer
				"ld r1, r0, 0",          // load the return address into r1
				"sub r0, r0, 1",         // decrement pointer
				"st r0, r2, __rsp",      // store the updated return stack pointer
				"jump __next_skip_load", // skip loading, r1 and r2 are already fine
			},
		},
		{
			name:   "+",
			goFunc: primitiveFuncPlus,
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
			name:   "-",
			goFunc: primitiveFuncMinus,
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
			name:   "SWAP",
			goFunc: primitiveFuncSwap,
			ulpAsm: PrimitiveUlp{
				"ld r1, r3, 0",
				"ld r0, r3, 1",
				"st r1, r3, 1",
				"st r0, r3, 0",
				"jump __next_skip_r2",
			},
		},
		{
			name:   "ROT",
			goFunc: primitiveFuncRot,
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
			name:   "DROP",
			goFunc: primitiveFuncDrop,
			ulpAsm: PrimitiveUlp{
				"add r3, r3, 1",
				"jump __next_skip_r2",
			},
		},
		{
			name:   "BYE",
			goFunc: primitiveFuncBye,
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
			name:   "ESP.FUNC.UNSAFE", // use one of the custom esp32/host functions
			goFunc: primitiveEspFunc,
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
			name:   "ESP.FUNC.READ.UNSAFE",
			goFunc: primitiveFuncEspRead,
			ulpAsm: PrimitiveUlp{
				"ld r0, r2, __boot_data_start+3",
				"sub r3, r3, 1",
				"st r0, r2, 0",
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

func primitiveFuncDotS(vm *VirtualMachine, entry *DictionaryEntry) error {
	fmt.Fprintf(vm.Out, "%s", vm.Stack)
	return nil
}

func primitiveFuncWords(vm *VirtualMachine, entry *DictionaryEntry) error {
	fmt.Fprintln(vm.Out)
	for i := len(vm.Dictionary.Entries) - 1; i >= 0; i-- {
		name := vm.Dictionary.Entries[i].Name
		if len(name) > 0 {
			fmt.Fprint(vm.Out, name, " ")
		}
	}
	return nil
}

func primitiveFuncSee(vm *VirtualMachine, entry *DictionaryEntry) error {
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
}

func primitiveFuncWord(vm *VirtualMachine, entry *DictionaryEntry) error {
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
}

func primitiveFuncCreateForth(vm *VirtualMachine, entry *DictionaryEntry) error {
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
}

func primitiveFuncLeftBracket(vm *VirtualMachine, entry *DictionaryEntry) error {
	return vm.State.Set(StateInterpret)
}

func primitiveFuncRightBracket(vm *VirtualMachine, entry *DictionaryEntry) error {
	return vm.State.Set(StateCompile)
}

func primitiveFuncLast(vm *VirtualMachine, entry *DictionaryEntry) error {
	last := vm.Dictionary.Entries[len(vm.Dictionary.Entries)-1]
	c := CellEntry{last}
	return vm.Stack.Push(c)
}

func primitiveFuncLiteral(vm *VirtualMachine, entry *DictionaryEntry) error {
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
}

func primitiveFindWord(vm *VirtualMachine, entry *DictionaryEntry) error {
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
}

func primitiveFuncPostpone(vm *VirtualMachine, entry *DictionaryEntry) error {
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
}

func primitiveSetImmediate(vm *VirtualMachine, entry *DictionaryEntry) error {
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
}

func primitiveSetHidden(vm *VirtualMachine, entry *DictionaryEntry) error {
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
}

func primitiveFuncCompile(vm *VirtualMachine, entry *DictionaryEntry) error {
	c, err := vm.Stack.Pop()
	if err != nil {
		return errors.Join(fmt.Errorf("%s could not pop from stack.", entry), err)
	}
	last, err := vm.Dictionary.LastForthWord()
	if err != nil {
		return errors.Join(fmt.Errorf("%s could not get last forth word.", entry), err)
	}
	last.Cells = append(last.Cells, c)
	return nil
}

func primitiveEspFunc(vm *VirtualMachine, entry *DictionaryEntry) error {
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
}

func primitiveFuncExit(vm *VirtualMachine, entry *DictionaryEntry) error {
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
}

func primitiveFuncPlus(vm *VirtualMachine, entry *DictionaryEntry) error {
	right, err := vm.Stack.PopNumber()
	if err != nil {
		return errors.Join(fmt.Errorf("%s could not get right value.", entry), err)
	}
	left, err := vm.Stack.PopNumber()
	if err != nil {
		return errors.Join(fmt.Errorf("%s could not get left value.", entry), err)
	}
	return vm.Stack.Push(CellNumber{left + right})
}

func primitiveFuncMinus(vm *VirtualMachine, entry *DictionaryEntry) error {
	right, err := vm.Stack.PopNumber()
	if err != nil {
		return errors.Join(fmt.Errorf("%s could not get right value.", entry), err)
	}
	left, err := vm.Stack.PopNumber()
	if err != nil {
		return errors.Join(fmt.Errorf("%s could not get left value.", entry), err)
	}
	return vm.Stack.Push(CellNumber{left - right})
}

func primitiveFuncSwap(vm *VirtualMachine, entry *DictionaryEntry) error {
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
}

func primitiveFuncRot(vm *VirtualMachine, entry *DictionaryEntry) error {
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
}

func primitiveFuncDrop(vm *VirtualMachine, entry *DictionaryEntry) error {
	_, err := vm.Stack.Pop()
	if err != nil {
		return errors.Join(fmt.Errorf("%s could not pop from stack.", entry), err)
	}
	return nil
}

func primitiveFuncBye(vm *VirtualMachine, entry *DictionaryEntry) error {
	return vm.State.Set(StateExit)
}

func primitiveFuncEspRead(vm *VirtualMachine, entry *DictionaryEntry) error {
	return vm.Stack.Push(CellNumber{0})
}

func notImplemented(vm *VirtualMachine, entry *DictionaryEntry) error {
	return fmt.Errorf("%s cannot be executed on the host.", entry)
}

func nop(vm *VirtualMachine, entry *DictionaryEntry) error {
	return nil
}
