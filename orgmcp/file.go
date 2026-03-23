package orgmcp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"os"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/p3rtang/org-mcp/embeddings"
	"github.com/p3rtang/org-mcp/utils/itertools"
	"github.com/p3rtang/org-mcp/utils/option"
	"github.com/p3rtang/org-mcp/utils/reader"
	"github.com/p3rtang/org-mcp/utils/result"
	"github.com/p3rtang/org-mcp/utils/slice"
)

type SearchScore struct {
	score  float64
	render Render
}

type OrgFile struct {
	name     string
	children []Render

	items       map[Uid]Render
	locationMap map[Uid]int
}

// Enforce that OrgFile implements the Render interface at compile time
var _ Render = (*OrgFile)(nil)

func OrgFileFromReader(ctx context.Context, r io.Reader) result.Result[OrgFile] {
	startTime := time.Now()

	org_file := OrgFile{
		items: map[Uid]Render{},
	}

	org_file.items[NewUid(0)] = &org_file

	peek_reader := reader.NewPeekReader(bufio.NewReader(r))

	current_line := 1

	var currentParent map[int]Render = map[int]Render{
		0: &org_file,
	}
	currentParentIdx := 1

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
				h.Parent = option.Some(currentParent[h.Level()-1])
				h.location = current_line
				currentParent[h.Level()-1].AddChildren(&h)
				org_file.items[h.Uid()] = &h
				current_line += 1
				currentParent[h.Level()] = &h
				currentParentIdx = h.Level()
				currentContentIndex = 0
				currentContentIndent = 0
			})
		case ' ':
			indent := len(val) - len(strings.TrimLeft(string(val), " "))
			ParseIndentedLine(peek_reader, currentParent[currentParentIdx]).Then(func(r Render) {
				if currentContentIndent == 0 {
				} else if currentContentIndent < indent {
					currentContentIndent = indent
					currentContentIndex += 1
				} else if currentContentIndent > indent {
					currentContentIndent = indent
					currentContentIndex -= 1
				}
				currentContentIndent = indent

				currentContent[currentContentIndex] = r
				current_line += 1

				if currentContentIndex >= 1 {
					currentContent[currentContentIndex-1].AddChildren(r)
				} else {
					currentParent[currentParentIdx].AddChildren(r)
				}

				org_file.items[r.Uid()] = r
			})
		default:
			panic("unreachable")
		}
	}

	org_file.BuildLocationTable()

	peek_reader.Continue()

	elapsed := time.Since(startTime)

	if logger, ok := ctx.Value("logger").(*slog.Logger); ok {
		logger.Info(fmt.Sprintf("Finished parsing org file with %d items in %d.%dms", len(org_file.items), elapsed.Milliseconds(), elapsed.Microseconds()))
	}

	ordered := slices.Collect(maps.Values(org_file.items))

	slices.SortFunc(ordered, func(i Render, j Render) int {
		return org_file.locationMap[i.Uid()] - org_file.locationMap[j.Uid()]
	})

	fmt.Fprintf(os.Stderr, "%s", PrintTable(ordered, []Column{
		ColType,
		ColUid,
		ColStatus,
		ColPreview,
		ColProgress,
		ColScheduled,
		ColDeadline,
		ColClosed,
	}))

	return result.Ok(org_file)
}

func ParseIndentedLine(r *reader.PeekReader, parent Render) option.Option[Render] {
	// errors have already been handled at this point
	bytes, _ := r.PeekBytes('\n')
	trimmed := strings.TrimSpace(string(bytes))
	// fmt.Fprintf(os.Stderr, "Indented: %s\n", string(bytes))

	switch trimmed[0] {
	case '-':
		fallthrough
	case '*':
		return option.Cast[*Bullet, Render](NewBulletFromReader(r))
	default:
		return option.Cast[*PlainText, Render](NewPlainTextFromReader(r))
	}
}

func (of *OrgFile) SetName(name string) {
	of.name = name
}

func (of *OrgFile) GenerateEmbeddings() error {
	headers := []Render{}
	contents := []string{}
	embedMap := make(map[string]embeddings.Embedding)

	for _, r := range of.items {
		builder := strings.Builder{}
		r.Render(&builder, 0)
		contents = append(contents, builder.String())
		headers = append(headers, r)
	}

	embeds, err := embeddings.Generate(contents)

	if err != nil {
		return err
	}

	for i, header := range headers {
		embedMap[header.Uid().String()] = embeds[i]
	}

	var fileName string
	if of.name == "" {
		fileName = "embeddings.json"
	} else {
		fileName = strings.TrimSuffix(of.name, ".org") + "_embeddings.json"
	}

	embedFile, err := os.Create(fileName)
	if err != nil {
		return err
	}

	defer embedFile.Close()

	encoder := json.NewEncoder(embedFile)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(embedMap); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to encode embeddings: %v\n", err)
		return err
	}

	return nil
}

func (of *OrgFile) VectorSearch(query string, top_n int) (h []Render, err error) {
	loadedEmbeddings := make(map[string]embeddings.Embedding)

	var fileName string
	if of.name == "" {
		fileName = "embeddings.json"
	} else {
		fileName = strings.TrimSuffix(of.name, ".org") + "_embeddings.json"
	}

	embedFile, err := os.Open(fileName)
	if err != nil {
		return
	}

	defer embedFile.Close()

	decoder := json.NewDecoder(embedFile)
	if err = decoder.Decode(&loadedEmbeddings); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to decode embeddings: %v\n", err)
		return
	}

	query_embed, err := embeddings.Generate([]string{query})
	if err != nil {
		return
	}

	scores := []SearchScore{}
	for uid, embed := range loadedEmbeddings {
		scores = append(scores, SearchScore{
			render: of.items[NewUid(uid)],
			score:  embeddings.Similarity(query_embed[0][:], embed[:]),
		})
	}

	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	h = slice.Map(scores, func(s SearchScore) Render {
		return s.render
	})[:top_n]

	return
}

func (of *OrgFile) Render(builder *strings.Builder, depth int) {
	if depth == 0 {
		return
	}

	var headers []Render
	var content []Render

	for _, child := range of.children {
		if _, ok := child.(*Header); ok {
			headers = append(headers, child)
		} else {
			content = append(content, child)
		}
	}

	for _, child := range content {
		child.Render(builder, depth-1)
	}

	for _, child := range headers {
		child.Render(builder, depth-1)
	}
}

func (of *OrgFile) Location(_ map[Uid]int) int {
	return 0
}

func (of *OrgFile) CheckProgress() option.Option[Progress] {
	return option.None[Progress]()
}

func (of *OrgFile) IndentLevel() int {
	return 0
}

func (of *OrgFile) ChildIndentLevel() int {
	return 0
}

func (of *OrgFile) Level() int {
	return 0
}

func (of *OrgFile) Children() []Render {
	return of.children
}

func (of *OrgFile) ChildrenRec(depth int) []Render {
	children := []Render{}

	if depth == 0 {
		return children
	}

	for _, child := range of.Children() {
		children = append(children, child)
		children = append(children, child.ChildrenRec(depth-1)...)
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
		return !slices.Contains(uids, r.Uid())
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

func (of *OrgFile) GetHeaderByStatus(status RenderStatus) []*Header {
	headers := []*Header{}

	for _, child := range of.items {
		if header, ok := child.(*Header); ok && header.Status() == status {
			headers = append(headers, header)
		}
	}

	return headers
}

type StatusReport struct {
	Count int   `json:"count"`
	Ids   []Uid `json:"ids"`
}

func (of *OrgFile) GetStatusOverview() map[RenderStatus]StatusReport {
	overview := make(map[RenderStatus]StatusReport)

	for _, child := range of.items {
		if header, ok := child.(*Header); ok {
			if header.Status() != RenderStatus(None) {
				if _, exists := overview[header.Status()]; !exists {
					overview[header.Status()] = StatusReport{Count: 0, Ids: []Uid{}}
				}

				overview[header.Status()] = StatusReport{
					Count: overview[header.Status()].Count + 1,
					Ids:   append(overview[header.Status()].Ids, header.Uid()),
				}
			}
		}
	}

	return overview
}

func (of *OrgFile) GetTagOverview() map[string]int {
	tagMap := make(map[string]int)

	for _, child := range of.items {
		for _, item := range child.TagList() {
			tagMap[item] += 1
		}
	}

	return tagMap
}

func (of *OrgFile) Uid() Uid {
	return NewUid(0)
}

func GetByType[T Render](of *OrgFile) map[Uid]T {
	mapping := make(map[Uid]T)

	for uid, r := range of.items {
		if item, ok := r.(T); ok {
			mapping[uid] = item
		}
	}

	return mapping
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

func (of *OrgFile) BuildLocationTable() *map[Uid]int {
	location_table := make(map[Uid]int)

	for _, r := range of.items {
		location_table[r.Uid()] = r.Location(location_table)
	}

	of.locationMap = location_table

	return &of.locationMap
}

func (of *OrgFile) GetLocationTable() *map[Uid]int {
	return &of.locationMap
}

func (of *OrgFile) Status() RenderStatus {
	return ""
}

func (of *OrgFile) TagList() (list TagList) {
	return
}

func (of *OrgFile) Preview(_ int) string {
	return ""
}

func (of *OrgFile) Path() string {
	return ""
}
