package orgmcp

import (
	"main/utils/option"
	"strings"
)

type Render interface {
	Location() int
	CheckProgress() option.Option[Progress]
	Render(builder *strings.Builder)
	IndentLevel() int
	AddChild(Render) error
	Children() []Render
}
