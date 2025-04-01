/*
Copyright 2024-2025 Blake Felt blake.w.felt@gmail.com

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/
package forth

import "fmt"

// The structure for stacks.
type Stack struct {
	stack []Cell
}

func (s Stack) String() string {
	return fmt.Sprintf("%s", s.stack)
}

// Set up the stack.
func (s *Stack) Setup() error {
	s.stack = make([]Cell, 0)
	return nil
}

func (s *Stack) Reset() {
	s.stack = s.stack[:0]
}

// Push a cell onto the stack.
func (s *Stack) Push(c Cell) error {
	s.stack = append(s.stack, c)
	return nil
}

// Pop a cell from the stack.
func (s *Stack) Pop() (Cell, error) {
	last := len(s.stack) - 1
	if last < 0 {
		return nil, fmt.Errorf("attempted to pop empty stack")
	}
	c := s.stack[last]
	s.stack = s.stack[:last]
	return c, nil
}

// Pop a number from the stack.
func (s *Stack) PopNumber() (uint16, error) {
	cell, err := s.Pop()
	if err != nil {
		return 0, err
	}
	cellNumber, ok := cell.(CellNumber)
	if !ok {
		return 0, fmt.Errorf("could not convert cell to number: %s type %T", cell, cell)
	}
	return cellNumber.Number, nil
}

// Get the current depth of the stack.
func (s *Stack) Depth() int {
	return len(s.stack)
}

// Set the stack depth. Must not be greater than current depth.
func (s *Stack) SetDepth(depth int) error {
	if depth > len(s.stack) {
		return fmt.Errorf("cannot arbitrarily increase stack depth")
	}
	s.stack = s.stack[:depth]
	return nil
}
