/*
Copyright 2024-2025 Blake Felt blake.w.felt@gmail.com

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/
package forth

// The input parse area.
type ParseArea struct {
	area  []byte
	index int

	savedArea  []byte
	savedIndex int
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

func (p *ParseArea) Save() error {
	p.savedArea = p.area
	p.savedIndex = p.index
	// reset the area
	p.area = nil
	p.index = 0
	return nil
}

func (p *ParseArea) Restore() error {
	p.area = p.savedArea
	p.index = p.savedIndex
	// reset the saved area
	p.savedArea = nil
	p.savedIndex = 0
	return nil
}

func (p *ParseArea) Word(delimiter byte, escape bool) ([]byte, error) {
	// trim starting whitespace
	startIndex := p.index
	for ; startIndex < len(p.area); startIndex++ {
		if !isWhitespace(p.area[startIndex]) {
			break
		}
	}

	// find the end
	endIndex := startIndex
	if endIndex > len(p.area) {
		p.index = endIndex
		return nil, nil
	}
	escapeNext := false
L:
	for ; endIndex < len(p.area); endIndex++ {
		c := p.area[endIndex]
		switch delimiter {
		case ' ':
			if isWhitespace(c) {
				break L
			}
		default:
			if c == delimiter {
				if escape {
					if !escapeNext {
						break L
					}
				} else {
					break L
				}
			}
			if !escapeNext && c == '\\' {
				escapeNext = true
			} else {
				escapeNext = false
			}
		}
	}
	name := p.area[startIndex:endIndex]
	p.index = endIndex + 1
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
