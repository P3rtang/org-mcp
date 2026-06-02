package codeblock

import (
	"fmt"
	"slices"
	"strings"

	"github.com/p3rtang/org-mcp/utils/itertools"
	"github.com/p3rtang/org-mcp/utils/option"
	"github.com/p3rtang/org-mcp/utils/reader"
	"github.com/p3rtang/org-mcp/utils/slice"
)

const (
	CODE_BLOCK_BEGIN = "#+BEGIN_SRC"
	CODE_BLOCK_END   = "#+END_SRC"
	UNEXPECTED_END   = "Unexpected end of tokens while parsing a `%s`"
	INVALID_PREFIX   = "The prefix `%s` is not valid for a `%s`"
)

type CodeBlockBegin struct {
	lang string
	// TODO: add <switches>
	// TODO: add <header arguments>
}

func parseBeginSrc(r *reader.PeekReader) (*CodeBlockBegin, error) {
	bytes, err := r.ReadBytes('\n')
	line := strings.TrimSpace(string(bytes))
	if err != nil {
		return nil, fmt.Errorf(UNEXPECTED_END, "CodeBlock")
	}

	line = strings.TrimSpace(string(line))

	parts := slice.Extend(
		slices.Collect(
			itertools.Filter(
				strings.SplitSeq(line, " "),
				func(str string) bool { return str != "" },
			),
		),
		4,
	)

	if CODE_BLOCK_BEGIN != strings.ToUpper(parts[0]) {
		return nil, fmt.Errorf(INVALID_PREFIX, parts[0], "CodeBlock")
	}

	return &CodeBlockBegin{lang: parts[1]}, nil
}

func NewCodeBlockFromReader(r *reader.PeekReader) (cb CodeBlock, err error) {
	var begin *CodeBlockBegin
	begin, err = parseBeginSrc(r)

	builder := strings.Builder{}

	for line, err := r.ReadBytes('\n'); err != nil; line, err = r.ReadBytes('\n') {
		if strings.ToUpper(strings.TrimSpace(string(line))) == CODE_BLOCK_END {
			var lang option.Option[string]
			if begin.lang == "" {
				lang = option.None[string]()
			} else {
				lang = option.Some(begin.lang)
			}

			return NewCodeBlock(builder.String(), option.None[string](), lang), nil
		}

		builder.Write(line)
	}

	return
}
