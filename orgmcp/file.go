package orgmcp

import (
	"bufio"
	"io"
	"main/utils/itertools"
	"main/utils/option"
	"main/utils/reader"
	"main/utils/result"
	"main/utils/slice"
	"maps"
	"slices"
	"strings"
)

type OrgFile struct {
	name     string
	children []Render

	items map[int]Render
}

// Enforce that OrgFile implements the Render interface at compile time
var _ Render = (*OrgFile)(nil)

func OrgFileFromReader(r io.Reader) result.Result[OrgFile] {
	org_file := OrgFile{
		items: map[int]Render{},
	}

	org_file.items[0] = &org_file

	peek_reader := reader.NewPeekReader(bufio.NewReader(r))

	var current_parent map[int]Render = map[int]Render{
		0: &org_file,
	}
	current_parent_idx := 0
	var current_line = 1

	for val, err := peek_reader.PeekBytes('\n'); true; val, err = peek_reader.PeekBytes('\n') {
		if err == io.EOF {
			break
		}

		if err != nil {
			return result.Err[OrgFile](err)
		}

		if len(strings.TrimSpace(string(val))) == 0 {
			peek_reader.Continue()
			continue
		}

		switch val[0] {
		case '*':
			peek_reader.Continue()
			NewHeaderFromString(string(val), peek_reader.Reader).Then(func(h Header) {
				h.Parent = option.Some(current_parent[h.Level])
				h.location = current_line
				current_parent[h.Level].AddChildren(&h)
				org_file.items[h.Uid()] = &h
				current_line += 1
				current_parent[h.Level+1] = &h
				current_parent_idx = h.Level + 1
			})
		case ' ':
			ParseIndentedLine(peek_reader, current_parent[current_parent_idx]).Then(func(r Render) {
				org_file.items[r.Uid()] = r
				current_line += 1
				current_parent[current_parent_idx].AddChildren(r)
			})
		default:
			panic("unreachable")
		}
	}

	peek_reader.Continue()

	return result.Ok(org_file)
}

func ParseIndentedLine(r *reader.PeekReader, parent Render) option.Option[Render] {
	// errors have already been handled at this point
	bytes, _ := r.PeekBytes('\n')
	trimmed := strings.TrimSpace(string(bytes))

	switch trimmed[0] {
	case '-':
		fallthrough
	case '*':
		r.Continue()
		return option.Map(NewBulletFromString(trimmed, parent), func(b Bullet) Render { return &b })
	default:
		return option.Cast[PlainText, Render](NewPlainTextFromReader(r))
	}
}

func (of *OrgFile) Render(builder *strings.Builder, depth int) {
	if depth == 0 {
		return
	}

	for _, child := range of.children {
		child.Render(builder, depth-1)
	}
}

func (of *OrgFile) Location() int {
	return 0
}

func (of *OrgFile) CheckProgress() option.Option[Progress] {
	return option.None[Progress]()
}

func (of *OrgFile) IndentLevel() int {
	return 0
}

func (of *OrgFile) Children() []Render {
	return of.children
}

func (of *OrgFile) ChildrenRec() []Render {
	children := []Render{}

	for _, child := range of.Children() {
		children = append(children, child.ChildrenRec()...)
	}

	return children
}

func (of *OrgFile) AddChildren(r ...Render) error {
	of.children = append(of.children, r...)

	return nil
}

func (of *OrgFile) RemoveChildren(uids ...int) error {
	of.children = slice.Filter(of.children, func(r Render) bool {
		return slices.Contains(uids, of.Uid())
	})

	return nil
}

func (of *OrgFile) GetLine(line int) option.Option[Render] {
	if item := of.items[line]; item != nil {
		return option.Some(item)
	}

	return option.None[Render]()
}

func (of *OrgFile) GetUid(uid int) option.Option[Render] {
	if uid == 0 {
		return option.Some[Render](of)
	}

	if child, found := of.items[uid]; found {
		return option.Some(child)
	}

	return option.None[Render]()
}

func (of *OrgFile) ParentUid() int {
	return 0
}

func (of *OrgFile) GetTag(tag string) option.Option[*Header] {
	return option.Cast[Render, *Header](
		itertools.Find(
			maps.Values(of.items),
			func(r Render) bool {
				header, ok := r.(*Header)
				if !ok {
					return false
				}

				return option.Map(header.Tags, func(t TagList) bool {
					return slices.Contains(t, tag)
				}).UnwrapOr(false)
			},
		),
	)
}

func (of *OrgFile) GetHeaderByStatus(status HeaderStatus) []*Header {
	headers := []*Header{}

	for _, child := range of.items {
		if header, ok := child.(*Header); ok && header.Status == status {
			headers = append(headers, header)
		}
	}

	return headers
}

func (of *OrgFile) Uid() int {
	return 0
}

func GetHeaderRec(header *Header, predicate func(*Header) bool, headers []*Header) []*Header {
	if predicate(header) {
		headers = append(headers, header)
		return headers
	}

	for _, child := range header.children {
		if header, ok := child.(*Header); ok {
			headers = GetHeaderRec(header, predicate, headers)
		}
	}

	return headers
}
