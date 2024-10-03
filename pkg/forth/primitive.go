package forth

import (
	"errors"
	"fmt"
)

type primitive struct {
	name   string
	goFunc PrimitiveGo
	ulpAsm PrimitiveUlp
}

func primitiveAdd(vm *VirtualMachine, name string, goFunc PrimitiveGo, ulpAsm PrimitiveUlp) error {
	var entry DictionaryEntry
	entry = DictionaryEntry{
		Name: name,
		Word: WordPrimitive{
			Go:    goFunc,
			Ulp:   ulpAsm,
			Entry: &entry,
		},
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
			name:   "WORD",
			goFunc: primitiveFuncWord,
		},
		{
			name:   "\\",
			goFunc: primitiveFuncBackslash,
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
		err := primitiveAdd(vm, p.name, p.goFunc, p.ulpAsm)
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
	for i := len(vm.Dictionary.Entries) - 1; i >= 0; i-- {
		name := vm.Dictionary.Entries[i].Name
		if len(name) > 0 {
			fmt.Print(name)
			fmt.Print(" ")
		}
	}
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

func primitiveFuncBackslash(vm *VirtualMachine, entry *DictionaryEntry) error {
	_, err := vm.ParseArea.Word('\n')
	if err != nil {
		return errors.Join(fmt.Errorf("%s error while skipping through newline.", entry), err)
	}
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
