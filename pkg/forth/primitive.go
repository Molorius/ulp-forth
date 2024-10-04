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
			name:   "--'",
			goFunc: primitiveFuncTick,
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
			name:   "U.",
			goFunc: primitiveUDot,
			ulpAsm: PrimitiveUlp{
				"ld r0, r3, 0",                   // load the value we want to print
				"mv r1, 2",                       // indicate that the host should do the printu16 method
				"st r0, r2, __boot_data_start+4", // set the value
				"st r1, r2, __boot_data_start+3", // set the method indicator
				"jump next",
			},
		},
		{
			name:   "EXIT",
			goFunc: primitiveFuncExit,
			ulpAsm: PrimitiveUlp{
				"ld r0, r2, __rsp", // load the return stack pointer
				"ld r1, r0, 0",     // load the return address into r0
				"sub r0, r0, 1",    // decrement pointer
				"st r0, r2, __rsp", // store the updated return stack pointer
				"jump next",
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
				"st r0, r0, 0",
				"jump next",
			},
		},
		{
			name:   "DROP",
			goFunc: primitiveFuncDrop,
			ulpAsm: PrimitiveUlp{
				"add r3, r3, 1",
				"jump next",
			},
		},
		{
			name:   "BYE",
			goFunc: primitiveFuncBye,
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
	fmt.Printf("%s", vm.Stack)
	return nil
}

func primitiveFuncWords(vm *VirtualMachine, entry *DictionaryEntry) error {
	fmt.Println("")
	for i := len(vm.Dictionary.Entries) - 1; i >= 0; i-- {
		name := vm.Dictionary.Entries[i].Name
		if len(name) > 0 {
			fmt.Print(name)
			fmt.Print(" ")
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
	fmt.Println()
	fmt.Print(wordEntry.Details())
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

func primitiveFuncTick(vm *VirtualMachine, entry *DictionaryEntry) error {
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

func primitiveUDot(vm *VirtualMachine, entry *DictionaryEntry) error {
	cell, err := vm.Stack.Pop()
	if err != nil {
		return errors.Join(fmt.Errorf("%s could not pop from stack.", entry), err)
	}
	fmt.Printf("%s ", cell)
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
