package forth

import "fmt"

// A Cell is the smallest unit of address within Forth.
// Cells are what constitute words, stack entries, etc.
type Cell interface {
	Execute(*VirtualMachine) error
}

// A Cell representing a number.
type CellNumber struct {
	Number uint16 // The unsigned number that this cell represents.
}

func (c CellNumber) Execute(vm *VirtualMachine) error {
	return vm.Stack.Push(c) // Push copy of this cell onto stack.
}

func (c CellNumber) String() string {
	return fmt.Sprintf("%d", c.Number)
}

// A Cell representing an entry in the Dictionary such as an execution token.
type CellEntry struct {
	Entry *DictionaryEntry // The dictionary entry this represents.
}

func (c CellEntry) Execute(vm *VirtualMachine) error {
	return c.Entry.Word.Execute(vm) // Execute the underlying dictionary entry.
}

func (c CellEntry) String() string {
	return c.Entry.String()
}

// A Cell representing an address in the dictionary. Used for pointers such
// as return addresses.
type CellAddress struct {
	Entry  *DictionaryEntry // The dictionary entry with the address.
	Offset int              // The offset within that entry.
}

func (c CellAddress) Execute(vm *VirtualMachine) error {
	return fmt.Errorf("Cannot execute an address cell.")
}

func (c CellAddress) String() string {
	return fmt.Sprintf("{%s %d}", c.Entry, c.Offset)
}

// A Cell representing a string.
type CellString struct {
	Memory []byte
	Offset int
}

func (c CellString) Execute(vm *VirtualMachine) error {
	return fmt.Errorf("Cannot execute a string cell.")
}

func (c CellString) String() string {
	return fmt.Sprintf("\"%s\"", c.Memory[c.Offset:])
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

// A destination to branch to. Only used during compilation.
type CellDestination struct {
	ulpName string      // the name we're going to compile this into
	Addr    CellAddress // the address of this destination
}

func (c *CellDestination) Execute(vm *VirtualMachine) error {
	return nil
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
	c.ulpName = u.name("dest", "")
	return c.ulpName
}

func (c *CellDestination) String() string {
	return fmt.Sprintf("Dest(%p)", c)
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

// A conditional branch.
type CellBranch0 struct {
	dest *CellDestination
}

func (c *CellBranch0) Execute(vm *VirtualMachine) error {
	n, err := vm.Stack.PopNumber()
	if err != nil {
		return err
	}
	if n == 0 {
		addr := c.dest.copyAddress() // probably unsafe, yay for gc!
		vm.IP = &addr
	}
	return nil
}
