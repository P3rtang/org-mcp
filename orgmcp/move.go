package orgmcp

import "errors"

type MoveOperationKind interface {
	SwapOperation | IndexOperation | IndexRelativeOperation
}

type MoveOperation struct {
	Swap          *SwapOperation
	Index         *IndexOperation
	IndexRelative *IndexRelativeOperation
}

func NewMoveOperation[T MoveOperationKind](kind T) MoveOperation {
	switch v := any(kind).(type) {
	case SwapOperation:
		return MoveOperation{Swap: &v}
	case IndexOperation:
		return MoveOperation{Index: &v}
	case IndexRelativeOperation:
		return MoveOperation{IndexRelative: &v}
	default:
		panic("invalid move operation kind")
	}
}

func (op MoveOperation) MoveSlice(slice []Render) (res []Render, err error) {
	if op.Swap != nil {
		leftIndex := -1
		rightIndex := -1

		for i, item := range slice {
			if item.Uid() == op.Swap.uidLeft {
				leftIndex = i
			}
			if item.Uid() == op.Swap.uidRight {
				rightIndex = i
			}
		}

		if leftIndex == -1 || rightIndex == -1 {
			err = errors.New("Uids not found in slice")
			return
		}

		slice[leftIndex], slice[rightIndex] = slice[rightIndex], slice[leftIndex]
		res = slice
	} else if op.Index != nil {
		index := -1

		for i, item := range slice {
			if item.Uid() == op.Index.uid {
				index = i
				break
			}
		}

		if index == -1 {
			err = errors.New("Uid not found in slice")
			return
		}

		item := slice[index]
		slice = append(slice[:index], slice[index+1:]...)
		res = append(slice[:op.Index.to], append([]Render{item}, slice[op.Index.to:]...)...)
	} else if op.IndexRelative != nil {
		index := -1

		for i, item := range slice {
			if item.Uid() == op.IndexRelative.uid {
				index = i
				break
			}
		}

		if index == -1 {
			err = errors.New("Uid not found in slice")
			return
		}

		newIndex := index + op.IndexRelative.offset
		if newIndex < 0 || newIndex >= len(slice) {
			err = errors.New("New index out of bounds")
			return
		}

		item := slice[index]
		slice = append(slice[:index], slice[index+1:]...)
		res = append(slice[:newIndex], append([]Render{item}, slice[newIndex:]...)...)
	}

	return
}

type SwapOperation struct {
	uidLeft  Uid
	uidRight Uid
}

type IndexOperation struct {
	uid Uid
	to  int
}

type IndexRelativeOperation struct {
	uid    Uid
	offset int
}

func (of *OrgFile) Move(op MoveOperation) (err error) {
	c, err := op.MoveSlice(of.children)
	if err != nil {
		return
	}

	of.children = c
	return
}

func (h *Header) Move(op MoveOperation) (err error) {
	c, err := op.MoveSlice(h.children)
	if err != nil {
		return
	}

	h.children = c
	return
}

func (b *Bullet) Move(op MoveOperation) (err error) {
	c, err := op.MoveSlice(b.children)
	if err != nil {
		return
	}

	b.children = c
	return
}

func (p *PlainText) Move(op MoveOperation) (err error) {
	return errors.New("PlainText cannot have children")
}
