/*
Copyright 2024-2025 Blake Felt blake.w.felt@gmail.com

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/
package forth

import "fmt"

func GetCellNumber(c Cell) (CellNumber, error) {
	cellNumber, ok := c.(CellNumber)
	if !ok {
		return CellNumber{}, fmt.Errorf("attempted to convert cell to number: %v", c)
	}
	return cellNumber, nil
}

func bytesToCells(bytes []byte, countedString bool) ([]Cell, error) {
	in := make([]byte, 0)
	if countedString {
		in = append(in, byte(len(bytes)))
	}
	in = append(in, bytes...)
	c := make([]Cell, 0)
	end := len(in) - len(in)%2
	for i := 0; i < end; i += 2 {
		lower := uint16(in[i])
		upper := uint16(in[i+1])
		val := (upper << 8) | lower
		c = append(c, CellNumber{val})
	}
	if len(in)%2 == 1 { // if there is an extra
		c = append(c, CellNumber{uint16(in[len(in)-1])})
	}
	return c, nil
}

func cellsToBytes(cells []Cell) ([]byte, error) {
	out := make([]byte, 0)
	for _, c := range cells {
		n, ok := c.(CellNumber)
		if !ok {
			return nil, fmt.Errorf("can only convert CellNumber array to bytes")
		}
		upper := byte(n.Number >> 8)
		lower := byte(n.Number)
		out = append(out, lower, upper)
	}
	return out, nil
}

func addrToString(addr CellAddress) (string, error) {
	word, ok := addr.Entry.Word.(*WordForth)
	if !ok {
		return "", fmt.Errorf("can only read string of forth word found %T", addr.Entry.Word)
	}
	cells := word.Cells
	offset := addr.Offset
	upper := addr.UpperByte
	return countedCellsToString(cells, offset, upper)
}

func countedCellsToString(cells []Cell, offset int, upper bool) (string, error) {
	bytes, err := cellsToBytes(cells[offset:])
	if err != nil {
		return "", err
	}
	if upper {
		bytes = bytes[1:]
	}
	length := bytes[0]
	return bytesToString(bytes[1:], int(length))
}

func cellsToString(cells []Cell, length int, upper bool) (string, error) {
	bytes, err := cellsToBytes(cells)
	if err != nil {
		return "", err
	}
	if upper {
		return bytesToString(bytes[1:], length)
	}
	return bytesToString(bytes, length)
}

func bytesToString(bytes []byte, length int) (string, error) {
	if length < 0 || length > len(bytes) {
		return "", fmt.Errorf("string length is invalid: %d", length)
	}
	return string(bytes[:length]), nil
}
