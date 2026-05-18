package table

import (
	. "github.com/p3rtang/org-mcp/orgmcp/types"
	"github.com/p3rtang/org-mcp/utils/option"
)

func (t *Table) Status() (s RenderStatus)                   { return }
func (t *Table) CheckProgress() (o option.Option[Progress]) { return }
func (t *Table) TagList() TagList                           { return t.parent.TagList() }
