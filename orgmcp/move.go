package orgmcp

import (
	"errors"

	. "github.com/p3rtang/org-mcp/orgmcp/types"
)

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
