package orgmcp

import (
	"fmt"
	"github.com/p3rtang/org-mcp/utils/option"
	"strconv"
	"strings"
	"unicode"
)

type Progress struct {
	Total    int
	Complete int
	done     option.Option[bool]
}

func ProgressFromString(str string) option.Option[Progress] {
	complete := ""
	total := ""
	seenSlash := false

	if !strings.HasPrefix(str, "[") || !strings.HasSuffix(str, "]") {
		return option.None[Progress]()
	}

	for _, char := range str {
		if unicode.IsDigit(char) {
			if seenSlash {
				total += string(char)
			} else {
				complete += string(char)
			}
		}

		if char == '/' {
			seenSlash = true
		}
	}

	if !seenSlash {
		return option.None[Progress]()
	}

	totalInt, _ := strconv.ParseInt(total, 0, 0)
	completeInt, _ := strconv.ParseInt(complete, 0, 0)

	return option.Some(Progress{
		Total:    int(totalInt),
		Complete: int(completeInt),
	})
}

func (p *Progress) Render(builder *strings.Builder) {
	fmt.Fprintf(builder, "[%d/%d]", p.Complete, p.Total)
}

// Prog returns true if the progress is partially complete (at least one but not all).
func (p *Progress) Prog() bool {
	return p.Complete > 0 && p.Complete < p.Total
}

func (p *Progress) Done() bool {
	return p.done.UnwrapOrElse(func() bool { return p.Total == p.Complete })
}
