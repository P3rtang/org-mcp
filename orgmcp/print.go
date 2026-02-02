package orgmcp

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/p3rtang/org-mcp/utils/slice"
)

type Column string

const (
	ColType          Column = "TYPE"
	ColUid           Column = "UID"
	ColPreview       Column = "PREVIEW"
	ColContent       Column = "CONTENT"
	ColStatus        Column = "STATUS"
	ColProgress      Column = "PROGRESS"
	ColParent        Column = "PARENT"
	ColChildrenCount Column = "CHILDREN_COUNT"
	ColTags          Column = "TAGS"
	ColLevel         Column = "LEVEL"
	ColPath          Column = "PATH"
	ColScheduled     Column = "SCHEDULED"
	ColDeadline      Column = "DEADLINE"
	ColClosed        Column = "CLOSED"
)

var (
	ColTypeValue          = ColType
	ColUidValue           = ColUid
	ColPreviewValue       = ColPreview
	ColContentValue       = ColContent
	ColStatusValue        = ColStatus
	ColProgressValue      = ColProgress
	ColParentValue        = ColParent
	ColChildrenCountValue = ColChildrenCount
	ColTagsValue          = ColTags
	ColLevelValue         = ColLevel
	ColPathValue          = ColPath
	ColScheduledValue     = ColScheduled
	ColDeadlineValue      = ColDeadline
	ColClosedValue        = ColClosed
)

var AllColumns = []Column{
	ColType,
	ColUid,
	ColPreview,
	ColContent,
	ColStatus,
	ColProgress,
	ColParent,
	ColChildrenCount,
	ColTags,
	ColLevel,
	ColPath,
	ColScheduled,
	ColDeadline,
	ColClosed,
}

var AllColumnsStr = strings.Join(slice.Map(AllColumns, func(c Column) string { return c.String() }), ", ")

func (c Column) Value(r Render, quoteChars string) (val string) {
	switch c {
	case ColType:
		val = string(reflect.TypeOf(r).String())
	case ColUid:
		val = string(r.Uid().String())
	case ColPreview:
		val = r.Preview(60)
		val = strings.ReplaceAll(val, "\n", "\\n")
	case ColContent:
		var contentBuilder strings.Builder
		r.Render(&contentBuilder, 0)
		val = strings.TrimRight(contentBuilder.String(), "\n")
		val = strings.ReplaceAll(val, "\n", "\\n")
		if strings.ContainsAny(val, quoteChars) {
			val = fmt.Sprintf("\"%s\"", val)
		}
	case ColStatus:
		val = r.Status().String()
	case ColProgress:
		if p, ok := r.CheckProgress().Split(); ok && p.Total > 0 {
			val = fmt.Sprintf("%d/%d", p.Complete, p.Total)
		}
	case ColParent:
		val = r.ParentUid().String()
	case ColChildrenCount:
		val = fmt.Sprint(len(r.Children()))
	case ColTags:
		val = strings.Join(r.TagList(), ",")
		if strings.ContainsAny(val, quoteChars) {
			val = fmt.Sprintf("\"%s\"", val)
		}
	case ColLevel:
		val = fmt.Sprintf("%d", r.Level())
	case ColPath:
		val = r.Path()
	case ColScheduled:
		if header, ok := r.(*Header); ok && header.Schedule().IsSome() {
			date := header.Schedule().Unwrap().Values[Scheduled]
			if !date.T.IsZero() {
				val = date.T.Format("2006-01-02")
			}
		}
	case ColDeadline:
		if header, ok := r.(*Header); ok && header.Schedule().IsSome() {
			date := header.Schedule().Unwrap().Values[Deadline]
			if !date.T.IsZero() {
				val = date.T.Format("2006-01-02")
			}
		}
	case ColClosed:
		if header, ok := r.(*Header); ok && header.Schedule().IsSome() {
			date := header.Schedule().Unwrap().Values[Closed]
			if !date.T.IsZero() {
				val = date.T.Format("2006-01-02")
			}
		}
	}

	return
}

func (c *Column) String() string {
	return string(*c)
}

func (c *Column) UnmarshalJSON(input []byte) error {
	col := strings.Trim(strings.ToUpper(string(input)), "\"")

	switch col {
	case "TYPE":
		*c = ColType
	case "UID":
		*c = ColUid
	case "PREVIEW":
		*c = ColPreview
	case "CONTENT":
		*c = ColContent
	case "STATUS":
		*c = ColStatus
	case "PROGRESS":
		*c = ColProgress
	case "PARENT":
		*c = ColParent
	case "CHILDREN_COUNT":
		*c = ColChildrenCount
	case "TAGS":
		*c = ColTags
	case "PATH":
		*c = ColPath
	case "LEVEL":
		*c = ColLevel
	case "SCHEDULED":
		*c = ColScheduled
	case "DEADLINE":
		*c = ColDeadline
	case "CLOSED":
		*c = ColClosed
	default:
		return fmt.Errorf("Unknown column type %s\n, potential values are: %s\n", col, AllColumnsStr)
	}

	return nil
}

type ColumnContent struct {
	content []string
	size    int
}

func PrintTable(r []Render, cols []Column) string {
	builder := strings.Builder{}

	columnContent := make([]ColumnContent, len(cols))

	for i, col := range cols {
		columnContent[i].size = len(string(col))

		for _, item := range r {
			val := col.Value(item, "")

			columnContent[i].content = append(columnContent[i].content, val)
			columnContent[i].size = max(columnContent[i].size, len(val))
		}
	}

	for i, col := range cols {
		builder.WriteString("| ")
		builder.WriteString(string(col))
		builder.WriteString(strings.Repeat(" ", columnContent[i].size-len(string(col))))
		builder.WriteString(" ")
	}
	builder.WriteString("|\n")

	for _, c := range columnContent {
		builder.WriteString("+")
		builder.WriteString(strings.Repeat("-", c.size+2))
	}
	builder.WriteString("+\n")

	for row := range r {
		for colIdx := range cols {
			builder.WriteString("| ")
			content := columnContent[colIdx].content[row]
			builder.WriteString(content)
			builder.WriteString(strings.Repeat(" ", columnContent[colIdx].size-len(content)))
			builder.WriteString(" ")
		}
		builder.WriteString("|\n")
	}

	builder.WriteString("\n")

	return builder.String()
}

func PrintCsv(r []Render, cols []*Column) string {
	builder := strings.Builder{}

	for i, col := range cols {
		builder.WriteString(string(*col))
		if i < len(cols)-1 {
			builder.WriteString(",")
		}
	}

	builder.WriteString("\n")

	for _, item := range r {
		for i, col := range cols {
			val := col.Value(item, ",")

			builder.WriteString(val)
			if i < len(cols)-1 {
				builder.WriteString(",")
			}
		}
		builder.WriteString("\n")
	}

	return builder.String()
}
