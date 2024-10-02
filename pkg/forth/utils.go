package forth

import "fmt"

func GetCellNumber(c Cell) (CellNumber, error) {
	cellNumber, ok := c.(CellNumber)
	if !ok {
		return CellNumber{}, fmt.Errorf("Attempted to convert cell to number: %v", c)
	}
	return cellNumber, nil
}
