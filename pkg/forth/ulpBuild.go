package forth

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type ulpAsm struct {
	name string
	asm  []string
}

func (uAsm ulpAsm) build() string {
	var sb strings.Builder
	sb.WriteString(uAsm.name)
	sb.WriteString(":\r\n")
	for _, asm := range uAsm.asm {
		if !strings.Contains(asm, ":") {
			sb.WriteString("    ")
		}
		sb.WriteString(asm)
		sb.WriteString("\r\n")
	}
	return sb.String()
}

type ulpForthCell struct {
	cell Cell
	name string
}

type ulpForth struct {
	name  string
	cells []ulpForthCell
}

func (uForth ulpForth) build() string {
	var sb strings.Builder
	sb.WriteString(uForth.name)
	sb.WriteString(":\r\n")
	for _, cell := range uForth.cells {
		if !strings.Contains(cell.name, ":") {
			sb.WriteString("    .int ")
		}
		sb.WriteString(cell.name)
		sb.WriteString("\r\n")
	}
	return sb.String()
}

type Ulp struct {
	assembly []ulpAsm
	forth    []ulpForth
	data     map[string]string
	outCount int
}

func (u *Ulp) build() string {
	var sb strings.Builder
	// header
	sb.WriteString(u.buildInterpreter())
	// assembly
	sb.WriteString("\r\n.text\r\n")
	sb.WriteString("__asmwords_start:\r\n")
	for _, asm := range u.assembly {
		sb.WriteString(asm.build())
	}
	sb.WriteString("__asmwords_end:\r\n\r\n")
	// forth
	sb.WriteString(".data\r\n")
	sb.WriteString("__forthwords_start:\r\n")
	for _, forth := range u.forth {
		sb.WriteString(forth.build())
	}
	sb.WriteString("__forthwords_end:\r\n\r\n")
	// data
	sb.WriteString("__forthdata_start:\r\n")
	for _, d := range u.data {
		sb.WriteString(d)
		sb.WriteString("\r\n")
	}
	sb.WriteString("__forthdata_end:\r\n")
	return sb.String()
}

// Build the assembly using the word passed in as the main function.
// Note that the virtual machine will be unusable after this.
func (u *Ulp) BuildAssembly(vm *VirtualMachine, word string) (string, error) {
	// put back into interpret state and compile the main ulp program
	// u.Entries = make([]*DictionaryEntry, 0)
	u.assembly = make([]ulpAsm, 0)
	u.forth = make([]ulpForth, 0)
	u.data = make(map[string]string)

	vm.State.Set(StateInterpret)
	// err := vm.execute([]byte(" : VM.INIT VM.STACK.INIT 0 2 MUTEX.TAKE --ESP.FUNC MUTEX.GIVE DEBUG.PAUSE 0 1 MUTEX.TAKE --ESP.FUNC MUTEX.GIVE VM.STOP ; "))
	err := vm.execute([]byte(" : VM.INIT VM.STACK.INIT " + word + " ESP.DONE VM.STOP ; "))
	if err != nil {
		return "", errors.Join(fmt.Errorf("could not compile the supporting words for ulp cross-compiling."), err)
	}
	vmInitEntry := vm.Dictionary.Entries[len(vm.Dictionary.Entries)-1]
	_, err = u.findUsedEntry(vmInitEntry)
	str := u.build()
	return str, err
}

// recursively find all dictionary entries that the entry uses
func (u *Ulp) findUsedEntry(entry *DictionaryEntry) (string, error) {
	if entry.ulpName != "" { // name is already set, we've already found this
		return entry.ulpName, nil
	}
	switch w := entry.Word.(type) {
	case *WordForth:
		name := u.name("forth", entry.Name, true)
		entry.ulpName = name
		forth := ulpForth{
			name:  name,
			cells: make([]ulpForthCell, len(w.Cells)),
		}

		for i, c := range w.Cells {
			str, err := u.findUsedCell(c)
			if err != nil {
				return "", err
			}
			forth.cells[i] = ulpForthCell{
				cell: c,
				name: str,
			}
		}
		u.forth = append(u.forth, forth)
		return name, nil
	case *WordPrimitive:
		name := u.name("asm", entry.Name, true)
		entry.ulpName = name
		asm := ulpAsm{
			name: name,
			asm:  w.Ulp,
		}
		u.assembly = append(u.assembly, asm)
		return name, nil
	default:
		return "", fmt.Errorf("Type %T not supported for cross compile, word: %s", w, w)
	}
}

func (u *Ulp) findUsedCell(cell Cell) (string, error) {
	switch c := cell.(type) {
	case CellEntry:
		return u.findUsedEntry(c.Entry)
	case CellNumber:
		return strconv.Itoa(int(c.Number)), nil
	case CellLiteral:
		pointedName, err := u.findUsedCell(c.cell)
		if err != nil {
			return "", err
		}
		var name string
		_, isNum := c.cell.(CellNumber)
		if isNum {
			name = u.name("number", pointedName, false)
		} else {
			name = u.name("literal", pointedName, false)
		}
		_, ok := u.data[name]
		if !ok {
			u.data[name] = fmt.Sprintf("%s: .int %s", name, pointedName)
		}
		return name, nil
	case CellData:
		name := c.Data.ulpName
		if name == "" {
			name = u.name("data", "unnamed", true)
			c.Data.ulpName = name
		}
		_, ok := u.data[name]
		if !ok {
			// build the definition
			var sb strings.Builder
			sb.WriteString(name)
			sb.WriteString(":")
			if len(c.Data.Cells) > 0 {
				sb.WriteString(" .int ")
				val0, err := u.findUsedCell(c.Data.Cells[0])
				if err != nil {
					return "", err
				}
				sb.WriteString(val0)
				for i := 1; i < len(c.Data.Cells); i++ {
					val, err := u.findUsedCell(c.Data.Cells[i])
					if err != nil {
						return "", err
					}
					sb.WriteString(", ")
					sb.WriteString(val)
				}
			}
			u.data[name] = sb.String()
		}
		return fmt.Sprintf("%s+%d", name, c.Offset), nil
	case *CellBranch0:
		return fmt.Sprintf("%s + 0x4000", c.dest.name(u)), nil
	case *CellBranch:
		return fmt.Sprintf("%s + 0x8000", c.dest.name(u)), nil
	case *CellDestination:
		return fmt.Sprintf("%s:", c.name(u)), nil
	default:
		return "", fmt.Errorf("Type %T not supported for cross compile, cell: %s", c, c)
	}
}

func (u *Ulp) name(middle string, word string, addSuffix bool) string {
	// Forth words can have any character, replace them all
	fixed := u.replaceOtherChars(word)
	if word == "VM.INIT" { // keep this static so we can keep the vm code static
		return "__forth_VM.INIT"
	}
	if addSuffix {
		n := fmt.Sprintf("__%s_%s_%d", middle, fixed, u.outCount)
		u.outCount++
		return n
	} else {
		return fmt.Sprintf("__%s_%s", middle, fixed)
	}
}

// replace all non-alphanumeric chars.
func (u *Ulp) replaceOtherChars(str string) string {
	var sb strings.Builder
	bytes := []byte(str)

	for _, b := range bytes {
		switch true {
		case b >= '0' && b <= '9',
			b >= 'a' && b <= 'z',
			b >= 'A' && b <= 'Z',
			b == '.',
			b == '_':
			sb.WriteByte(b)
		default:
			sb.WriteString(fmt.Sprintf("_special%d_", int(b)))
		}
	}
	return sb.String()
}

func (u *Ulp) buildInterpreter() string {
	i := []string{
		// required data, will be placed at the start of .data
		".boot.data",
		".int 0, 0, 0",               // mutex
		".int 0, 0",                  // send to host
		"__ip: .int __forth_VM.INIT", // instruction pointer starts at word VM.INIT
		"__rsp: .int __stack_start",  // return stack pointer starts at the beginning of the stack section

		// boot labels
		".boot",
		".global entry",
		"entry:",
		"next:",

		// load the instruction
		"move r2, 0",        // r2 is 0 at the start of every loop as a global pointer
		"__next_skip_r2:",   // address to skip loading r2
		"ld r1, r2, __ip",   // load the instruction pointer
		"__next_skip_load:", // address to skip loading IP
		"add r1, r1, 1",     // increment the pointer to the next instruction
		"st r1, r2, __ip",   // store the pointer
		"ld r0, r1, -1",     // load the current instruction

		// determine instruction type
		"__ins_asm:",
		"jumpr __ins_forth, __forthwords_start, ge",
		// it's assembly, execute immediately
		"jump r0",

		"__ins_forth:",
		"jumpr __ins_num, __forthdata_start, ge",
		// it's forth
		"st r0, r2, __ip",  // put the address into the instruction pointer
		"ld r0, r2, __rsp", // load the return stack pointer
		"add r0, r0, 1",    // increment the rsp
		"st r1, r0, 0",     // store the instruction we were about to execute onto the return stack
		"st r0, r2, __rsp", // store the rsp
		"jump next",        // then start the vm again at the defined instruction

		"__ins_num:",
		"jumpr __ins_branch0, __forthdata_end, gt",
		// it's a number or variable
		"ld r0, r0, 0",  // load the number
		"sub r3, r3, 1", // increase the stack by 1
		"st r0, r3, 0",  // store the number
		"jump next",     // next!

		"__ins_branch0:",
		"jumpr __ins_branch, 0x8000, ge",
		// it's a conditional branch, check stack
		"ld r1, r3, 0",            // get value from stack
		"add r3, r3, 1",           // decrement stack
		"add r1, r1, 0xFFFF",      // add number with all bits set
		"jump __next_skip_r2, ov", // if overflow then number wasn't 0, continue vm. Otherwise jump.

		"__ins_branch:",
		// it's a definite branch
		"and r1, r0, 0x3FFF",    // get the lowest 14 bits
		"jump __next_skip_load", // then continue vm at this newer address
		"",                      // newline at end
	}
	return strings.Join(i, "\r\n")
}
