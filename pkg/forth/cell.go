/*
Copyright 2024-2025 Blake Felt blake.w.felt@gmail.com

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/
package forth

import (
	"fmt"
	"strings"
)

func safeCall() string {
	return "__safe_call: "
}

// A Cell is the smallest unit of address within Forth.
// Cells are what constitute words, stack entries, etc.
type Cell interface {
	Execute(*VirtualMachine) error
	// Compile(*Ulp) (string, error) // The compiled string for the output assembly.
	AddToList(*Ulp) error
	// Build the assembly necessary to execute this cell.
	BuildExecution(*Ulp) (string, error)
	// The string reference for this cell, used for literals
	// and when the cell is stored as data.
	OutputReference(*Ulp) (string, error)
	// If this cell in some way refers to the input word, return true.
	IsRecursive(*WordForth) bool
}

// A Cell representing a number.
type CellNumber struct {
	Number uint16 // The unsigned number that this cell represents.
}

func (c CellNumber) Execute(vm *VirtualMachine) error {
	return fmt.Errorf("Cannot directly execute a number")
}

func (c CellNumber) String() string {
	// return fmt.Sprintf("#%d", c.Number)
	return fmt.Sprintf("%d", c.Number)
}

func (c CellNumber) AddToList(u *Ulp) error {
	// Numbers directly in a word are not executed,
	// so we don't add to a list.
	return nil
}

func (c CellNumber) BuildExecution(u *Ulp) (string, error) {
	return "", fmt.Errorf("Cannot directly execute number")
}

func (c CellNumber) OutputReference(u *Ulp) (string, error) {
	return fmt.Sprintf("%d", c.Number), nil
}

func (c CellNumber) IsRecursive(check *WordForth) bool {
	return false
}

// A Cell representing an address in the dictionary. Used for pointers such
// as return addresses.
type CellAddress struct {
	Entry     *DictionaryEntry // The dictionary entry with the address.
	Offset    int              // The offset within that entry.
	UpperByte bool             // If we're aligned to the upper byte.
}

func (c CellAddress) Execute(vm *VirtualMachine) error {
	switch w := c.Entry.Word.(type) {
	case *WordForth:
		if c.Offset >= len(w.Cells) {
			return fmt.Errorf("Trying to get data from outside of allocated data.")
		}
		if c.Entry.Flag.Data {
			return vm.Stack.Push(w.Cells[c.Offset])
		}
		return w.ExecuteOffset(vm, c.Offset)
	case *WordPrimitive:
		if c.Offset != 0 {
			return fmt.Errorf("Cannot execute a primitive word at an offset.")
		}
		return w.Execute(vm)
	default:
		return fmt.Errorf("Cannot execute word type %t", c.Entry.Word)
	}
}

func (c CellAddress) AddToList(u *Ulp) error {
	return c.Entry.AddToList(u)
}

func (c CellAddress) BuildExecution(u *Ulp) (string, error) {
	c.Entry.Flag.inToken = true
	name, err := c.OutputReference(u)
	if err != nil {
		return "", err
	}

	switch u.compileTarget {
	case UlpCompileTargetToken:
		return fmt.Sprintf(".int %s", name), nil
	case UlpCompileTargetSubroutine:
		return fmt.Sprintf("jump %s", name), nil
	default:
		return "", fmt.Errorf("Unknown compile target %d, please file a bug report", u.compileTarget)
	}
}

func (c CellAddress) OutputReference(u *Ulp) (string, error) {
	name := c.Entry.ulpName
	if c.Offset != 0 {
		name = fmt.Sprintf("%s+%d", name, c.Offset)
	}
	if c.UpperByte {
		name = name + "+0x8000"
	}
	return name, nil
}

func (c CellAddress) IsRecursive(check *WordForth) bool {
	return c.Entry.Word.IsRecursive(check)
}

func (c CellAddress) String() string {
	s := ""
	if c.Entry.Name == "" {
		s = fmt.Sprintf("Address{%p", c.Entry)
	} else {
		s = fmt.Sprintf("Address{%s", c.Entry)
	}
	if c.Offset != 0 {
		s += fmt.Sprintf(" %d", c.Offset)
	}
	if c.UpperByte {
		s += " upper"
	}
	s += "}"
	return s
}

// A Cell that places the underlying cell on the stack.
// Used for execution tokens.
type CellLiteral struct {
	cell Cell
}

func (c CellLiteral) Execute(vm *VirtualMachine) error {
	return vm.Stack.Push(c.cell)
}

func (c CellLiteral) String() string {
	return fmt.Sprintf("Literal(%s)", c.cell)
}

func (c CellLiteral) AddToList(u *Ulp) error {
	err := c.cell.AddToList(u)
	if err != nil {
		return err
	}
	ref, err := c.cell.OutputReference(u)
	if err != nil {
		return err
	}
	_, ok := c.cell.(CellAddress)
	if ok {
		ref = "__body" + ref
	}
	litref := c.reference(ref)
	u.literals[litref] = ref
	cellAddress, ok := c.cell.(CellAddress)
	if ok {
		cellAddress.Entry.Flag.inToken = true
	}
	return nil
}

func (c CellLiteral) reference(s string) string {
	return "__literal_" + strings.ReplaceAll(s, "+", "_plus_")
}

func (c CellLiteral) BuildExecution(u *Ulp) (string, error) {
	name, err := c.cell.OutputReference(u)
	if err != nil {
		return "", err
	}
	_, ok := c.cell.(CellAddress)
	if ok {
		name = "__body" + name
	}
	switch u.compileTarget {
	case UlpCompileTargetToken:
		ref := c.reference(name)
		return fmt.Sprintf(".int %s", ref), nil
	case UlpCompileTargetSubroutine:
		return fmt.Sprintf("move r0, %s\r\n%sjump __add_to_stack", name, safeCall()), nil
	default:
		return "", fmt.Errorf("Unknown compile target %d, please file a bug report", u.compileTarget)
	}
}

func (c CellLiteral) OutputReference(u *Ulp) (string, error) {
	return "", fmt.Errorf("Cannot refer to a CellLiteral, please file a bug report")
}

func (c CellLiteral) IsRecursive(check *WordForth) bool {
	return c.cell.IsRecursive(check)
}

// A destination to branch to. Only used during compilation.
type CellDestination struct {
	ulpName string      // the name we're going to compile this into
	Addr    CellAddress // the address of this destination
}

func (c *CellDestination) Execute(vm *VirtualMachine) error {
	return nil
}

func (c *CellDestination) AddToList(u *Ulp) error {
	// Add the destination word to a list.
	return c.Addr.AddToList(u)
}

func (c *CellDestination) BuildExecution(u *Ulp) (string, error) {
	return c.name(u) + ":", nil
}

func (c *CellDestination) OutputReference(u *Ulp) (string, error) {
	return "", fmt.Errorf("Cannot refer to a destination, please file a bug report")
}

func (c *CellDestination) IsRecursive(check *WordForth) bool {
	return false
}

func (c *CellDestination) copyAddress() CellAddress {
	return CellAddress{
		Entry:  c.Addr.Entry,
		Offset: c.Addr.Offset,
	}
}

func (c *CellDestination) name(u *Ulp) string {
	if c.ulpName != "" {
		return c.ulpName
	}
	c.ulpName = u.name("dest", "", true)
	return c.ulpName
}

func (c *CellDestination) String() string {
	return fmt.Sprintf("Dest{%p}", c)
}

// A definite branch.
type CellBranch struct {
	dest *CellDestination
}

func (c *CellBranch) Execute(vm *VirtualMachine) error {
	addr := c.dest.copyAddress() // probably unsafe, yay for gc!
	vm.IP = &addr
	return nil
}

func (c *CellBranch) AddToList(u *Ulp) error {
	// Add the destination to the list.
	return c.dest.AddToList(u)
}

func (c *CellBranch) BuildExecution(u *Ulp) (string, error) {
	switch u.compileTarget {
	case UlpCompileTargetToken:
		return fmt.Sprintf(".int %s + 0x8000", c.dest.name(u)), nil
	case UlpCompileTargetSubroutine:
		return fmt.Sprintf("move r2, %s\r\njump r2", c.dest.name(u)), nil
	default:
		return "", fmt.Errorf("Unknown compile target %d, please file a bug report", u.compileTarget)
	}
}

func (c *CellBranch) OutputReference(u *Ulp) (string, error) {
	return "", fmt.Errorf("Cannot refer to a branch, please file a bug report")
}

func (c *CellBranch) IsRecursive(check *WordForth) bool {
	return false
}

func (c *CellBranch) String() string {
	return fmt.Sprintf("Branch{%p}", c.dest)
}

// A conditional branch.
type CellBranch0 struct {
	dest *CellDestination
}

func (c *CellBranch0) Execute(vm *VirtualMachine) error {
	cellValue, err := vm.Stack.Pop()
	if err != nil {
		return err
	}
	switch value := cellValue.(type) {
	case CellNumber:
		if value.Number == 0 {
			addr := c.dest.copyAddress() // probably unsafe, yay for gc!
			vm.IP = &addr
		}
	default:
	}
	return nil
}

func (c *CellBranch0) AddToList(u *Ulp) error {
	// Add the destination to the list.
	return c.dest.AddToList(u)
}

func (c *CellBranch0) BuildExecution(u *Ulp) (string, error) {
	switch u.compileTarget {
	case UlpCompileTargetToken:
		return fmt.Sprintf(".int %s + 0x4000", c.dest.name(u)), nil
	case UlpCompileTargetSubroutine:
		return fmt.Sprintf("move r1, %s\r\n%sjump __branch_if", c.dest.name(u), safeCall()), nil
	default:
		return "", fmt.Errorf("Unknown compile target %d, please file a bug report", u.compileTarget)
	}
}

func (c *CellBranch0) OutputReference(u *Ulp) (string, error) {
	return "", fmt.Errorf("Cannot refer to a conditional branch, please file a bug report")
}

func (c *CellBranch0) IsRecursive(check *WordForth) bool {
	return false
}

func (c *CellBranch0) String() string {
	return fmt.Sprintf("Branch0{%p}", c.dest)
}

// Used during an optimization pass to add
// tail calls.
type CellTailCall struct {
	dest *WordForth
}

func (c *CellTailCall) Execute(vm *VirtualMachine) error {
	return fmt.Errorf("Cannot directly execute a tail call, please file a bug repot")
}

func (c *CellTailCall) AddToList(u *Ulp) error {
	// Add the destination to the list.
	return c.dest.AddToList(u)
}

func (c *CellTailCall) BuildExecution(u *Ulp) (string, error) {
	switch u.compileTarget {
	case UlpCompileTargetToken:
		return fmt.Sprintf(".int %s + 0x8000", c.dest.Entry.BodyLabel()), nil
	case UlpCompileTargetSubroutine:
		// put the address after the docol
		return fmt.Sprintf("move r2, %s\r\njump r2", c.dest.Entry.BodyLabel()), nil
	default:
		return "", fmt.Errorf("Unknown compile target %d, please file a bug report", u.compileTarget)
	}
}

func (c *CellTailCall) OutputReference(u *Ulp) (string, error) {
	return "", fmt.Errorf("Cannot refer to a tail call, please file a bug report")
}

func (c *CellTailCall) IsRecursive(check *WordForth) bool {
	return c.dest.IsRecursive(check)
}

func (c *CellTailCall) String() string {
	return fmt.Sprintf("TailCall{%s}", c.dest.Entry.Name)
}
