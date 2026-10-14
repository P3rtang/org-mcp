package reader

import (
	"bufio"
	"bytes"
	"errors"
	"io"
)

type PeekReader struct {
	reader     *bufio.Reader
	peekBuffer []byte
}

func NewPeekReader(r *bufio.Reader) *PeekReader {
	return &PeekReader{reader: r}
}

/*
Discard clears the peek buffer so subsequent reads pull fresh data from the
underlying reader. The underlying reader's read position has already advanced
past any data returned by PeekBytes or Peek, so the next read will see the
data that came after the peeked content.

Calling this method multiple times is safe and has no additional effect.
*/
func (p *PeekReader) Discard() {
	p.peekBuffer = nil
}

/*
Peek returns up to n bytes from the stream without advancing the underlying
reader's position. The data is cached in the internal peek buffer so that
subsequent PeekBytes or ReadBytes calls see the same content.

If the stream has fewer than n bytes remaining, the returned slice contains
whatever is available along with io.EOF.
*/
func (p *PeekReader) Peek(n int) ([]byte, error) {
	if n <= 0 {
		return nil, nil
	}

	if len(p.peekBuffer) >= n {
		return p.peekBuffer[:n:n], nil
	}

	need := n - len(p.peekBuffer)
	chunk := make([]byte, need)
	read, err := io.ReadFull(p.reader, chunk)
	p.peekBuffer = append(p.peekBuffer, chunk[:read]...)

	if len(p.peekBuffer) >= n {
		return p.peekBuffer[:n:n], nil
	}

	if errors.Is(err, io.ErrUnexpectedEOF) {
		err = io.EOF
	}

	return p.peekBuffer, err
}

/*
PeekBytes returns the bytes from the current position up to and including
the first occurrence of delim. The data is cached in the internal peek
buffer; subsequent PeekBytes calls with the same delim return the same
content until Discard or ReadBytes is called.

If delim is not found before EOF, the returned slice contains whatever
bytes remain along with io.EOF.
*/
func (p *PeekReader) PeekBytes(delim byte) ([]byte, error) {
	if idx := bytes.IndexByte(p.peekBuffer, delim); idx >= 0 {
		return p.peekBuffer[: idx+1 : idx+1], nil
	}

	chunk, err := p.reader.ReadBytes(delim)
	p.peekBuffer = append(p.peekBuffer, chunk...)

	if idx := bytes.IndexByte(p.peekBuffer, delim); idx >= 0 {
		return p.peekBuffer[: idx+1 : idx+1], nil
	}

	return p.peekBuffer, err
}

/*
ReadBytes reads until the first occurrence of delim and returns the bytes
including the delimiter. The consumed bytes are removed from the peek
buffer; any data in the underlying reader that follows the delimiter
becomes the new peek buffer for subsequent calls.

If delim is not found before EOF, the returned slice contains whatever
bytes remain along with io.EOF.
*/
func (p *PeekReader) ReadBytes(delim byte) ([]byte, error) {
	if idx := bytes.IndexByte(p.peekBuffer, delim); idx >= 0 {
		result := p.peekBuffer[: idx+1 : idx+1]
		p.peekBuffer = p.peekBuffer[idx+1:]
		return result, nil
	}

	var result []byte
	if len(p.peekBuffer) > 0 {
		result = p.peekBuffer
		p.peekBuffer = nil
	}

	chunk, err := p.reader.ReadBytes(delim)
	if err != nil {
		return chunk, err
	}

	result = append(result, chunk...)

	if idx := bytes.IndexByte(result, delim); idx >= 0 {
		return result[: idx+1 : idx+1], nil
	}

	return result, err
}
