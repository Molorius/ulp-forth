package forth

import "fmt"

// A Cell is the smallest unit of address within Forth.
// Cells are what constitute words, stack entries, etc.
type Cell interface {
	Execute(*VirtualMachine) error
	// Compile(*Ulp) (string, error) // The compiled string for the output assembly.
}

// A Cell representing a number.
type CellNumber struct {
	Number uint16 // The unsigned number that this cell represents.
}

func (c CellNumber) Execute(vm *VirtualMachine) error {
	return fmt.Errorf("Cannot directly execute a number")
}

func (c CellNumber) String() string {
	return fmt.Sprintf("#%d", c.Number)
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

func (c *CellBranch0) String() string {
	return fmt.Sprintf("Branch0{%p}", c.dest)
}
