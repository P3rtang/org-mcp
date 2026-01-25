package orgmcp

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"sort"
	"strings"

	"github.com/p3rtang/org-mcp/embeddings"
	"github.com/p3rtang/org-mcp/utils/itertools"
	"github.com/p3rtang/org-mcp/utils/option"
	"github.com/p3rtang/org-mcp/utils/reader"
	"github.com/p3rtang/org-mcp/utils/result"
	"github.com/p3rtang/org-mcp/utils/slice"
)

type SearchScore struct {
	score  float64
	header *Header
}

type OrgFile struct {
	name     string
	children []Render

	items map[Uid]Render
}

// Enforce that OrgFile implements the Render interface at compile time
var _ Render = (*OrgFile)(nil)

func OrgFileFromReader(r io.Reader) result.Result[OrgFile] {
	org_file := OrgFile{
		items: map[Uid]Render{},
	}

	org_file.items[NewUid(0)] = &org_file

	peek_reader := reader.NewPeekReader(bufio.NewReader(r))

	current_line := 1

	var currentParent map[int]Render = map[int]Render{
		0: &org_file,
	}
	currentParentIdx := 0

	var currentContent map[int]Render = map[int]Render{}
	currentContentIndex := 0
	currentContentIndent := 0

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
			NewHeaderFromString(string(val), peek_reader).Then(func(h Header) {
				h.Parent = option.Some(currentParent[h.Level])
				h.location = current_line
				currentParent[h.Level].AddChildren(&h)
				org_file.items[h.Uid()] = &h
				current_line += 1
				currentParent[h.Level+1] = &h
				currentParentIdx = h.Level + 1
				currentContentIndex = 0
				currentContentIndent = 0
			})
		case ' ':
			ParseIndentedLine(peek_reader, currentParent[currentParentIdx]).Then(func(r Render) {
				indent := r.IndentLevel()
				if currentContentIndent == 0 {
					currentContentIndent = indent
				} else if currentContentIndent < indent {
					currentContentIndent = indent
					currentContentIndex += 1
				}

				currentContent[currentContentIndex] = r
				current_line += 1

				fmt.Fprintf(os.Stderr, "Current content index: %d\n", currentContentIndex)

				if currentContentIndex >= 1 {
					currentContent[currentContentIndex-1].AddChildren(r)
				} else {
					currentParent[currentParentIdx].AddChildren(r)
				}

				org_file.items[r.Uid()] = r

				fmt.Fprintf(os.Stderr, "------------\nuid: %s\n------------\n", r.Uid())
			})
		default:
			panic("unreachable")
		}
	}

	peek_reader.Continue()

	fmt.Fprintf(os.Stderr, "Finished parsing org file with %d items\nMap: %v", len(org_file.items), org_file.items)

	return result.Ok(org_file)
}

func ParseIndentedLine(r *reader.PeekReader, parent Render) option.Option[Render] {
	// errors have already been handled at this point
	bytes, _ := r.PeekBytes('\n')
	trimmed := strings.TrimSpace(string(bytes))
	fmt.Fprintf(os.Stderr, "Indented: %s\n", string(bytes))

	switch trimmed[0] {
	case '-':
		fallthrough
	case '*':
		return option.Cast[*Bullet, Render](NewBulletFromReader(r))
	default:
		return option.Cast[*PlainText, Render](NewPlainTextFromReader(r))
	}
}

func (of *OrgFile) GenerateEmbeddings() error {
	headers := []*Header{}
	contents := []string{}

	for _, r := range of.items {
		if h, ok := r.(*Header); ok {
			headers = append(headers, h)
			builder := strings.Builder{}
			h.Render(&builder, 1)
			contents = append(contents, builder.String())
		}
	}

	embeds, err := embeddings.Generate(contents)

	if err != nil {
		return err
	}

	for i, header := range headers {
		header.embedding = option.Some(embeds[i])
	}

	return nil
}

func (of *OrgFile) VectorSearch(query string, top_n int) (h []*Header, err error) {
	headers := []*Header{}
	contents := []string{query}

	for _, r := range of.items {
		if h, ok := r.(*Header); ok {
			headers = append(headers, h)
			builder := strings.Builder{}
			h.Render(&builder, 1)
			contents = append(contents, builder.String())
		}
	}

	embeds, err := embeddings.Generate(contents)

	if err != nil {
		return
	}

	query_embed := embeds[0]
	header_embeds := embeds[1:]

	scores := []SearchScore{}
	for i, embed := range header_embeds {
		scores = append(scores, SearchScore{
			header: headers[i],
			score:  embeddings.Similarity(query_embed[:], embed[:]),
		})
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	h = slice.Map(scores, func(s SearchScore) *Header {
		return s.header
	})[:top_n]

	return
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
	for _, child := range r {
		child.SetParent(of)
	}

	of.children = append(of.children, r...)

	return nil
}

func (of *OrgFile) SetParent(render Render) error {
	return errors.New("OrgFile cannot have a parent")
}

func (of *OrgFile) RemoveChildren(uids ...Uid) error {
	of.children = slice.Filter(of.children, func(r Render) bool {
		return slices.Contains(uids, of.Uid())
	})

	return nil
}

func (of *OrgFile) GetUid(uid Uid) option.Option[Render] {
	if uid == NewUid(0) || uid == NewUid("root") {
		return option.Some[Render](of)
	}

	if child, found := of.items[uid]; found {
		return option.Some(child)
	}

	return option.None[Render]()
}

func (of *OrgFile) ParentUid() Uid {
	return NewUid(0)
}

func (of *OrgFile) GetTag(tag string) option.Option[*Header] {
	return option.Map(
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
		func(r Render) *Header { return r.(*Header) },
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

func (of *OrgFile) Uid() Uid {
	return NewUid(0)
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
