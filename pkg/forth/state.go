package forth

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
	state StateType
}

func (m *State) Setup(vm *VirtualMachine) error {
	m.state = StateInterpret
	return nil
}

func (m *State) Get() (StateType, error) {
	return m.state, nil
}

func (m *State) Set(stateType StateType) error {
	m.state = stateType
	return nil
}
