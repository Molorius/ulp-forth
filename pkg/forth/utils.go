package forth

import "fmt"

func GetCellNumber(c Cell) (CellNumber, error) {
	cellNumber, ok := c.(CellNumber)
	if !ok {
		return CellNumber{}, fmt.Errorf("Attempted to convert cell to number: %v", c)
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

func cellsToString(cells []Cell) (string, error) {
	bytes, err := cellsToBytes(cells)
	if err != nil {
		return "", err
	}
	length := bytes[0]
	if length < 0 || int(length+1) > len(bytes) {
		return "", fmt.Errorf("counted string length is invalid: %d", length)
	}
	return string(bytes[1 : length+1]), nil
}
