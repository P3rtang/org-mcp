package tools

import (
	"context"
	"fmt"
	"maps"
	"os"
	"regexp"
	"slices"
	"time"

	"github.com/p3rtang/org-mcp/mcp"
	"github.com/p3rtang/org-mcp/orgmcp"
	"github.com/p3rtang/org-mcp/utils/itertools"
)

type ViewItem struct {
	Uid     string               `json:"uid,omitempty" jsonschema:"description=UID of the header to view. If not provided, all headers are considered."`
	Status  *orgmcp.RenderStatus `json:"status,omitempty" jsonschema:"description=Filter headers by status (e.g. TODO ; DONE). Case insensitive. As well as bullets by their checkbox status (e.g. CHECKED ; UNCHECKED)."`
	Content string               `json:"content,omitempty" jsonschema:"description=Filter headers with a regex match on content. It will only consider the preview of the header content and not any metadata; children; status or other information."`
	Tags    []string             `json:"tags,omitempty" jsonschema:"description=Filter headers by tags. Only headers containing all specified tags will be returned."`
	Depth   *int                 `json:"depth,omitempty" jsonschema:"description=Depth of child headers to include. Default is 1 (only direct children)."`
	Date    *DateFilter          `json:"date,omitempty" jsonschema:"description=Filter headers by date criteria. Any text containing dates is not yet supported."`
}

type ViewInput struct {
	Items   []ViewItem `json:"items" jsonschema:"description=List of items to view based on their UIDs and filters.,required=true"`
	Columns ColumnList `json:"columns,omitempty"`
	Path    string     `json:"path,omitempty" jsonschema:"description=An optional file path; will default to ./.tasks.org"`
}

type ColumnList []*orgmcp.Column

func (cl ColumnList) GetSchema() map[string]any {
	return map[string]any{
		"description": `
## Description
List of columns to include in the output. If not specified defaults to [UID ; PREVIEW].

## Available columns
  - TYPE: The type of the item (e.g. headline, list item, etc.)
  - UID: The unique identifier of the item.
  - PREVIEW: A short preview of the item's content (first 100 characters).
  - CONTENT: The full content of the item, including all text and metadata.
  - STATUS: The status of the item (e.g. TODO, DONE, CHECKED, UNCHECKED).
  - PROGRESS: For items with children, the percentage of completed children.
  - PARENT: The UID of the parent item, if any.
  - CHILDREN_COUNT: The number of direct children this item has.
  - LEVEL: The level of the item in the hierarchy (1 for top-level headers, 2 for their children, etc.)
  - PATH: The file path of the item.
  - SCHEDULED: The scheduled date of the item, if any.
  - DEADLINE: The deadline date of the item, if any.
  - CLOSED: The closed date of the item, if any.
`,
		"type": "array",
		"items": map[string]any{
			"type": "string",
		},
	}
}

type DateFilter struct {
	Match      string  `json:"match,omitempty" jsonschema:"description=The type of date match to perform.,enum=SCHEDULED;DEADLINE;CLOSED,required=true"`
	ShowClosed bool    `json:"show_closed,omitempty" jsonschema:"description=Whether to include closed dates in the filter. Handy when using DEADLINE or SCHEDULED match types to show overdue and current tasks."`
	Date       *string `json:"date,omitempty" jsonschema:"description=The date to match against in YYYY-MM-DD format. Will default to today."`
	Range      *int    `json:"range,omitempty" jsonschema:"description=The range in days to consider. For example you could request all deadlines this week by setting date to today and range to 7. Range can be negative and can be used in combination with the ommited date to get the week ahead or behind. For example, setting range to -7 will get all deadlines in the past week. When a negative range is used, the date will be considered the end date of the range and not included but up to."`
}

var ViewTool = mcp.GenericTool[ViewInput]{
	Name: "query_items",
	Description: `
# query_items
  View and filter Org items.

## Arguments
  - items: Array of filters (OR logic between items).
    - uid: string (optional)
    - status: "TODO" | "DONE" | "CHECKED" | ...
    - content: string (regex match)
    - tags: Array<string> (all tags must be present)
    - date_filter:
      - match: "SCHEDULED" | "DEADLINE" | "CLOSED"
      - range: number (days, negative for past)
      - show_closed: boolean
    - depth: number (optional, defaults to 1), determines how many levels of children to include in the CSV.
  - path: string (defaults to ./.tasks.org), unless you encounter errors about file not found or otherwise specified leave this empty.
  - columns: Array of column names to include in the output CSV. Defaults to [UID ; PREVIEW]. See the columns section below for available columns.

## Summary
  Returns a CSV of matching items. See the columns section in the common instructions for what columns you can specify.
`,
	Callback: func(ctx context.Context, input ViewInput, options mcp.FuncOptions) (resp []any, err error) {
		var path string
		if input.Path == "" {
			path = options.DefaultPath
		} else {
			path = input.Path
		}

		orgFile, err := mcp.LoadOrgFile(ctx, path)
		if err != nil {
			return
		}

		if len(input.Columns) == 0 {
			input.Columns = []*orgmcp.Column{&orgmcp.ColUidValue, &orgmcp.ColPreviewValue}
		}

		results := map[orgmcp.Uid]orgmcp.Render{}

		for _, item := range input.Items {
			depth := 1
			if item.Depth != nil {
				depth = *item.Depth
			}

			for _, render := range orgFile.ChildrenRec(-1) {
				if item.Uid != "" && render.Uid().String() != item.Uid {
					continue
				}

				if item.Status != nil && render.Status() != *item.Status {
					continue
				}

				if item.Content != "" {
					reg, err := regexp.Compile(item.Content)
					if err != nil {
						return nil, err
					}

					preview := render.Preview(-1)
					if !reg.MatchString(preview) {
						continue
					}
				}

				if item.Date != nil {
					match, err := FilterDate(render, item.Date)
					if err != nil {
						return nil, err
					}

					if !match {
						continue
					}
				}

				if len(item.Tags) > 0 {
					foundAll := true
					for _, tag := range item.Tags {
						if !slices.Contains(render.TagList(), tag) {
							foundAll = false
							break
						}
					}

					if !foundAll {
						continue
					}
				}

				results[render.Uid()] = render
				for _, child := range render.ChildrenRec(depth) {
					results[child.Uid()] = child
				}
			}
		}

		locationTable := orgFile.GetLocationTable()

		ordered := slices.Collect(maps.Values(results))
		slices.SortFunc(ordered, func(a, b orgmcp.Render) int {
			return (*locationTable)[a.Uid()] - (*locationTable)[b.Uid()]
		})

		resp = append(resp, orgmcp.PrintCsv(ordered, input.Columns))
		_, err = mcp.WriteOrgFileToDisk(ctx, orgFile, path)

		return
	},
}

func FilterDate(r orgmcp.Render, dateFilter *DateFilter) (match bool, err error) {
	if dateFilter == nil {
		return true, err
	}

	var header *orgmcp.Header
	var ok bool

	if header, ok = r.(*orgmcp.Header); !ok {
		return
	}

	var schedule *orgmcp.Schedule
	if schedule, ok = header.Schedule().Split(); !ok {
		return
	}

	var filterDate = time.Now()
	if dateFilter.Date != nil {
		var parsed time.Time
		if parsed, err = time.Parse("2006-01-02", *dateFilter.Date); err != nil {
			return
		}

		filterDate = parsed
	}

	var startDate, endDate time.Time
	if dateFilter.Range != nil {
		if *dateFilter.Range < 0 {
			startDate = filterDate.AddDate(0, 0, *dateFilter.Range)
			endDate = filterDate
		} else {
			startDate = filterDate
			endDate = filterDate.AddDate(0, 0, *dateFilter.Range)
		}
	} else {
		startDate = filterDate
	}

	scheduleStatus, err := orgmcp.NewScheduleStatus(dateFilter.Match)
	if err != nil {
		return
	}
	date := schedule.Values[scheduleStatus]

	fmt.Fprintf(os.Stderr, "Filtering date %v between %v and %v\n", date.T, startDate, endDate)

	if !dateFilter.ShowClosed && slices.Contains(itertools.Collect(maps.Keys(schedule.Values)), orgmcp.Closed) {
		return
	}

	if endDate.IsZero() && date.T.Before(startDate) {
		return true, nil
	}

	if date.T.Equal(startDate) || date.T.Before(startDate) || date.T.After(endDate) {
		return false, nil
	}

	return true, nil
}
