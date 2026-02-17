package orgmcp

import "errors"

func (of *OrgFile) Insert(index int, render Render) (err error) {
	children := of.children[:index]
	children = append(children, render)
	of.children = append(children, of.children[index:]...)

	return
}

func (h *Header) Insert(index int, render Render) (err error) {
	children := h.children[:index]
	children = append(children, render)
	h.children = append(children, h.children[index:]...)

	return
}

func (b *Bullet) Insert(index int, render Render) (err error) {
	children := b.children[:index]
	children = append(children, render)
	b.children = append(children, b.children[index:]...)

	return
}

func (p *PlainText) Insert(index int, render Render) (err error) {
	return errors.New("PlainText cannot have children")
}
