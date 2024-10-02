package forth

// The input parse area.
type ParseArea struct {
	area []byte
}

// Set up the parse area.
func (p *ParseArea) Setup() error {
	p.area = make([]byte, 0)
	return nil
}
