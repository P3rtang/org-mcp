package orgmcp

import (
	"context"
	"errors"
)

func (of *OrgFile) Insert(ctx context.Context, index int, render Render) (err error) {
	children := of.children[:index]
	children = append(children, render)
	of.children = append(children, of.children[index:]...)

	return
}

func (of *OrgFile) InsertAfter(ctx context.Context, after Uid, render Render) (err error) {
	insertIdx := -1
	for i, child := range of.children {
		if child.Uid() == after {
			insertIdx = i + 1
			break
		}
	}

	if insertIdx == -1 {
		return errors.New("after Uid not found among children")
	}

	return of.Insert(ctx, insertIdx, render)
}

func (of *OrgFile) InsertBefore(ctx context.Context, before Uid, render Render) (err error) {
	insertIdx := -1
	for i, child := range of.children {
		if child.Uid() == before {
			insertIdx = i
			break
		}
	}

	if insertIdx == -1 {
		return errors.New("before Uid not found among children")
	}

	return of.Insert(ctx, insertIdx, render)
}

func (h *Header) Insert(ctx context.Context, index int, render Render) (err error) {
	children := h.children[:index]
	children = append(children, render)
	h.children = append(children, h.children[index:]...)

	return
}

func (h *Header) InsertAfter(ctx context.Context, after Uid, render Render) (err error) {
	insertIdx := -1
	for i, child := range h.children {
		if child.Uid() == after {
			insertIdx = i + 1
			break
		}
	}

	if insertIdx == -1 {
		return errors.New("after Uid not found among children")
	}

	return h.Insert(ctx, insertIdx, render)
}

func (h *Header) InsertBefore(ctx context.Context, before Uid, render Render) (err error) {
	insertIdx := -1
	for i, child := range h.children {
		if child.Uid() == before {
			insertIdx = i
			break
		}
	}

	if insertIdx == -1 {
		return errors.New("before Uid not found among children")
	}

	return h.Insert(ctx, insertIdx, render)
}

func (b *Bullet) Insert(ctx context.Context, index int, render Render) (err error) {
	children := b.children[:index]
	children = append(children, render)
	b.children = append(children, b.children[index:]...)

	return
}

func (b *Bullet) InsertAfter(ctx context.Context, after Uid, render Render) (err error) {
	insertIdx := -1
	for i, child := range b.children {
		if child.Uid() == after {
			insertIdx = i + 1
			break
		}
	}

	if insertIdx == -1 {
		return errors.New("after Uid not found among children")
	}

	return b.Insert(ctx, insertIdx, render)
}

func (b *Bullet) InsertBefore(ctx context.Context, before Uid, render Render) (err error) {
	insertIdx := -1
	for i, child := range b.children {
		if child.Uid() == before {
			insertIdx = i
			break
		}
	}

	if insertIdx == -1 {
		return errors.New("before Uid not found among children")
	}

	return b.Insert(ctx, insertIdx, render)
}

func (p *PlainText) Insert(ctx context.Context, index int, render Render) (err error) {
	return errors.New("PlainText cannot have children")
}

func (p *PlainText) InsertAfter(ctx context.Context, after Uid, render Render) (err error) {
	return errors.New("PlainText cannot have children")
}

func (p *PlainText) InsertBefore(ctx context.Context, before Uid, render Render) (err error) {
	return errors.New("PlainText cannot have children")
}
