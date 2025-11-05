package reader

import (
	"bufio"
	"slices"
)

type PeekReader struct {
	Reader *bufio.Reader

	peekBuffer []byte
}

func NewPeekReader(r *bufio.Reader) *PeekReader {
	return &PeekReader{Reader: r}
}

/*
Continue allows the PeekReader to continue reading from the underlying reader,
it will continue after any data previously peeked.

Calling this method multiple times is safe and has no additional effect.
*/
func (p *PeekReader) Continue() {
	p.peekBuffer = nil
}

func (p *PeekReader) Peek(n int) (bytes []byte, err error) {
	if len(p.peekBuffer) == 0 {
		bytes = make([]byte, n)

		return
	}

	bytes = make([]byte, n)
	copy(bytes, p.peekBuffer)
	_, err = p.Reader.Read(bytes[len(p.peekBuffer):])

	return
}

func (p *PeekReader) PeekBytes(r byte) (bytes []byte, err error) {
	idx := slices.Index(p.peekBuffer, r)

	if idx >= 0 {
		bytes = make([]byte, idx)
		copy(bytes, p.peekBuffer[0:idx])

		return
	}

	newBytes, err := p.Reader.ReadBytes(r)
	p.peekBuffer = append(p.peekBuffer, newBytes...)
	bytes = make([]byte, len(p.peekBuffer))
	copy(bytes, p.peekBuffer)

	return
}

func (p *PeekReader) ReadBytes(r byte) (bytes []byte, err error) {
	bytes = make([]byte, 0, 512)

	if len(p.peekBuffer) == 0 {
		return p.Reader.ReadBytes(r)
	}

	idx := slices.Index(p.peekBuffer, r)

	if idx >= 0 {
		copy(bytes, p.peekBuffer[0:idx])
		p.peekBuffer = p.peekBuffer[idx:]

		return
	}

	copy(bytes, p.peekBuffer)
	p.peekBuffer = nil

	extra, err := p.Reader.ReadBytes(r)
	bytes = append(bytes, extra...)

	return
}
