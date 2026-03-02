package orgmcp

import (
	"context"
	"errors"
	"fmt"
	"os"
)

var (
	ErrUidNotFound         error = errors.New("Uid not found in slice")
	ErrNewIndexOutOfBounds error = errors.New("New index out of bounds")
	ErrOrgFileNotFound     error = errors.New("orgFile not found in context")
	ErrNewParentNotFound   error = errors.New("New parent not found.")
	ErrItemNotFound        error = errors.New("Item not found.")
	ErrOldParentNotFound   error = errors.New("Parent of the item to move not found.")
	ErrPlainTextNoChildren error = errors.New("PlainText cannot have children")
)

type Movable interface {
	Uid() Uid
}

type MoveOperationKind interface {
	SwapOperation | IndexOperation | IndexRelativeOperation | SwapParentOperation
}

type MoveOperation[M Movable] struct {
	Swap          *SwapOperation
	Index         *IndexOperation
	IndexRelative *IndexRelativeOperation
	SwapParent    *SwapParentOperation
}

func NewMoveOperation[T MoveOperationKind, M Movable](kind T) MoveOperation[M] {
	switch v := any(kind).(type) {
	case SwapOperation:
		return MoveOperation[M]{Swap: &v}
	case IndexOperation:
		return MoveOperation[M]{Index: &v}
	case IndexRelativeOperation:
		return MoveOperation[M]{IndexRelative: &v}
	case SwapParentOperation:
		return MoveOperation[M]{SwapParent: &v}
	default:
		panic("invalid move operation kind")
	}
}

func (op MoveOperation[M]) MoveSlice(ctx context.Context, slice []M) (res []M, err error) {
	slice, err = op.swap(ctx, slice)
	slice, err = op.index(ctx, slice)
	slice, err = op.indexRelative(ctx, slice)
	slice, err = op.swapParent(ctx, slice)

	if err != nil {
		return nil, err
	}

	return slice, nil
}

func (op MoveOperation[M]) swap(_ context.Context, slice []M) (res []M, err error) {
	res = slice

	if op.Swap == nil {
		return
	}

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
		err = ErrUidNotFound
		return
	}

	slice[leftIndex], slice[rightIndex] = slice[rightIndex], slice[leftIndex]
	res = slice

	return
}

func (op MoveOperation[M]) index(_ context.Context, slice []M) (res []M, err error) {
	res = slice

	if op.Index == nil {
		return
	}

	index := -1

	for i, item := range slice {
		if item.Uid() == op.Index.Uid {
			index = i
			break
		}
	}

	if index == -1 {
		err = ErrUidNotFound
		return
	}

	item := slice[index]
	slice = append(slice[:index], slice[index+1:]...)
	res = append(slice[:op.Index.To], append([]M{item}, slice[op.Index.To:]...)...)

	return
}

func (op MoveOperation[M]) indexRelative(_ context.Context, slice []M) (res []M, err error) {
	res = slice
	if op.IndexRelative == nil {
		return
	}

	index := -1

	for i, item := range slice {
		if item.Uid() == op.IndexRelative.Uid {
			index = i
			break
		}
	}

	if index == -1 {
		err = ErrUidNotFound
		return
	}

	newIndex := index + op.IndexRelative.Offset
	if newIndex < 0 || newIndex >= len(slice) {
		err = ErrNewIndexOutOfBounds
		return
	}

	fmt.Fprintln(os.Stderr, "index:", index, "newIndex:", newIndex)

	item := slice[index]
	slice = append(slice[:index], slice[index+1:]...)
	res = append(slice[:newIndex], append([]M{item}, slice[newIndex:]...)...)

	return
}

func (op MoveOperation[M]) swapParent(ctx context.Context, slice []M) (res []M, err error) {
	res = slice

	if op.SwapParent == nil {
		return
	}

	of, ok := ctx.Value(ORG_FILE_KEY).(*OrgFile)
	if !ok {
		err = ErrOrgFileNotFound
		return
	}

	newParent := of.GetUid(op.SwapParent.newParent).UnwrapNoneAnd(func() { err = ErrNewParentNotFound })
	item := of.GetUid(op.SwapParent.uid).UnwrapNoneAnd(func() { err = ErrItemNotFound })
	if err != nil {
		return
	}

	oldParent := of.GetUid(item.ParentUid()).Unwrap()

	oldParent.RemoveChildren(item.Uid())
	err = newParent.Insert(ctx, op.SwapParent.indexRight, item)

	return
}

type SwapOperation struct {
	uidLeft  Uid
	uidRight Uid
}

type IndexOperation struct {
	Uid Uid
	To  int
}

type IndexRelativeOperation struct {
	Uid    Uid
	Offset int
}

type SwapParentOperation struct {
	uid        Uid
	newParent  Uid
	indexRight int
}

func (of *OrgFile) Move(ctx context.Context, op MoveOperation[Render]) (err error) {
	c, err := op.MoveSlice(ctx, of.children)
	if err != nil {
		return
	}

	of.children = c
	return
}

func (h *Header) Move(ctx context.Context, op MoveOperation[Render]) (err error) {
	c, err := op.MoveSlice(ctx, h.children)
	if err != nil {
		return
	}

	h.children = c
	return
}

func (b *Bullet) Move(ctx context.Context, op MoveOperation[Render]) (err error) {
	c, err := op.MoveSlice(ctx, b.children)
	fmt.Fprintf(os.Stderr, "bullet children after move: %v, got error: %v", c, err)
	if err != nil {
		return
	}

	b.children = c
	return
}

func (p *PlainText) Move(ctx context.Context, op MoveOperation[Render]) (err error) {
	return ErrPlainTextNoChildren
}
