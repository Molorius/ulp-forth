package forth

import "fmt"

// The execution state.
type StateType int

const (
	StateInterpret StateType = iota // Interpret parsed words.
	StateCompile                    // Compile parsed words to Forth.
	StateExit                       // Exit the virtual machine.
	StateUnknown                    // Unknown what state the vm is supposed to be in.
)

func (m StateType) String() string {
	switch m {
	case StateInterpret:
		return "Interpret"
	case StateCompile:
		return "Compile"
	case StateExit:
		return "Exit"
	default:
		return "Unknown"
	}
}

type State struct {
	entry *DictionaryEntry
}

func (m *State) Setup(vm *VirtualMachine) error {
	var entry DictionaryEntry
	entry = DictionaryEntry{
		Name: "STATE",
		Word: &WordMemory{
			Memory: []Cell{
				CellNumber{uint16(StateInterpret)},
			},
			Entry: &entry,
		},
	}
	m.entry = &entry
	return vm.Dictionary.AddEntry(m.entry)
}

func (m *State) Get() (StateType, error) {
	cell, err := m.getCell()
	if err != nil {
		return StateUnknown, err
	}
	state := cell.Number
	return StateType(state), nil
}

func (m *State) getCell() (CellNumber, error) {
	if m.entry == nil {
		return CellNumber{}, fmt.Errorf("State not initialized, please file a bug report.")
	}
	word := m.entry.Word
	memory, ok := word.(*WordMemory)
	if !ok {
		return CellNumber{}, fmt.Errorf("State is not a memory region, please file a bug report.")
	}
	if memory.Memory == nil {
		return CellNumber{}, fmt.Errorf("State memory not initialized, please file a bug report.")
	}
	if len(memory.Memory) != 1 {
		return CellNumber{}, fmt.Errorf("State memory not correct size (size %d), please file a bug report.", len(memory.Memory))
	}
	cell := memory.Memory[0]
	cellNumber, ok := cell.(CellNumber)
	if !ok {
		return CellNumber{}, fmt.Errorf("State memory is not a number, please file a bug report.")
	}
	return cellNumber, nil
}

func (m *State) Set(stateType StateType) error {
	cell, err := m.getCell()
	if err != nil {
		return err
	}
	cell.Number = uint16(stateType)
	word, ok := m.entry.Word.(*WordMemory)
	if !ok { // we already checked this but want to stay safe
		return fmt.Errorf("State memory is not a number while trying to set, please file a bug report.")
	}
	word.Memory[0] = cell // mildly unsafe but we already checked length earlier
	return nil
}
