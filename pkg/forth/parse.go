package forth

// The input parse area.
type ParseArea struct {
	area  []byte
	index int
}

// Set up the parse area.
func (p *ParseArea) Setup() error {
	p.area = make([]byte, 0)
	p.index = 0
	return nil
}

func (p *ParseArea) Fill(bytes []byte) error {
	p.index = 0                       // reset the index
	p.area = p.area[:0]               // remove everything in parse area
	p.area = append(p.area, bytes...) // refill with the new bytes
	return nil
}

func (p *ParseArea) Word() ([]byte, error) {
	// trim starting whitespace
	startIndex := p.index
	for ; startIndex < len(p.area); startIndex++ {
		if !isWhitespace(p.area[startIndex]) {
			break
		}
	}

	// find the end
	endIndex := startIndex + 1
	if endIndex > len(p.area) {
		p.index = endIndex
		return nil, nil
	}
	for ; endIndex < len(p.area); endIndex++ {
		if isWhitespace(p.area[endIndex]) {
			break
		}
	}
	name := p.area[startIndex:endIndex]
	p.index = endIndex
	return name, nil
}

func isWhitespace(b byte) bool {
	switch b {
	case ' ', '\r', '\n', '\t':
		return true
	default:
		return false
	}
}
