package orgmcp

import (
	"github.com/p3rtang/org-mcp/utils/option"
	"strings"
)

type Tag string

type TagList []string

func TagListFromString(str string) option.Option[TagList] {
	if !(strings.HasPrefix(str, ":") && strings.HasSuffix(str, ":")) {
		return option.None[TagList]()
	}

	var list TagList = strings.Split(str, ":")
	return option.Some(list[1 : len(list)-1])
}

func (tl TagList) Render(builder *strings.Builder) {
	builder.WriteRune(':')

	for _, tag := range tl {
		builder.WriteString(tag)
		builder.WriteRune(':')
	}
}
