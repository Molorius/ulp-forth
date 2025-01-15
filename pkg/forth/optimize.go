package forth

import (
	"errors"
	"fmt"
)

type Optimizer struct {
	u *Ulp
}

func (o *Optimizer) Optimize() error {
	// mark all recursive words for later passes
	err := o.tagRecursion()
	if err != nil {
		return errors.Join(fmt.Errorf("could not tag recursion during optimization"), err)
	}
	err = o.removeDeferred()
	if err != nil {
		return errors.Join(fmt.Errorf("could not remove deferred word"), err)
	}
	// change calls at end of words to tail calls
	err = o.putTailCalls()
	if err != nil {
		return errors.Join(fmt.Errorf("could not create tail calls, please file a bug report"), err)
	}
	return nil
}

// Mark the flag of every word that has recursion.
func (o *Optimizer) tagRecursion() error {
	// unmark all of the words
	for _, w := range o.u.forthWords {
		w.Entry.Flag.recursive = false
	}
	for _, w := range o.u.forthWords {
		o.clearVisited()            // clear every word every time
		for _, c := range w.Cells { // check every cell in that word
			if c.IsRecursive(w) {
				w.Entry.Flag.recursive = true
				break
			}
		}
	}
	return nil
}

// Put tail calls as an optimization.
func (o *Optimizer) putTailCalls() error {
	for _, w := range o.u.forthWords {
		length := len(w.Cells) - 1
		for i := 0; i < length; i++ {
			// check if the first cell is a forth word
			firstAddress, ok := w.Cells[i].(CellAddress)
			if !ok {
				continue
			}
			word, ok := firstAddress.Entry.Word.(*WordForth)
			if !ok {
				continue
			}
			// check if the second cell is an EXIT
			secondAddress, ok := w.Cells[i+1].(CellAddress)
			if !ok {
				continue
			}
			if !secondAddress.Entry.Flag.isExit {
				continue
			}
			// replace both cells with the tail call!
			tailCall := CellTailCall{word}     // create the tail call
			w.Cells[i] = &tailCall             // replace the word
			before := w.Cells[:i+1]            // get the cells before, including the tail call
			after := w.Cells[i+2:]             // get the cells after, excluding the exit
			w.Cells = append(before, after...) // recreate the list
			length -= 1                        // shift the length
		}
	}
	return nil
}

func (o *Optimizer) removeDeferred() error {
	for _, w := range o.u.forthWords {
		f := w.Entry.Flag
		// if this word was made with DEFER and it cannot be altered
		if f.isDeferred && !f.inToken {
			literal, ok := w.Cells[0].(CellLiteral)
			if !ok {
				return fmt.Errorf("could not read literal %s in %s", w.Cells[0], w)
			}
			address, ok := literal.cell.(CellAddress)
			if !ok {
				return fmt.Errorf("could not read address in %s", w)
			}
			data, ok := address.Entry.Word.(*WordForth)
			if !ok {
				return fmt.Errorf("error reading a deferred word, please file a bug report.")
			}
			if len(data.Cells) < 1 {
				return fmt.Errorf("deferred word not allocated, please file a bug report.")
			}
			embedded := data.Cells[0]
			exit := w.Cells[len(w.Cells)-1]
			w.Cells = w.Cells[0:2]
			w.Cells[0] = embedded
			w.Cells[1] = exit
		}
	}
	return nil
}

func (o *Optimizer) clearVisited() {
	for _, w := range o.u.forthWords {
		w.Entry.ClearVisited()
	}
}
