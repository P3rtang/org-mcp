package orgmcp

import (
	"bufio"
	"main/utils/option"
	"slices"
	"strings"
	"time"
)

type ScheduleStatus int

const (
	Deadline ScheduleStatus = iota
	Scheduled
	Closed
)

func newStatusFromKeyword(kw string) ScheduleStatus {
	switch kw {
	case "DEADLINE":
		return Deadline
	case "SCHEDULED":
		return Scheduled
	case "CLOSED":
		return Closed
	default:
		panic("unreachable")
	}
}

func (s ScheduleStatus) String() string {
	return ScheduleKeywords[s]
}

var ScheduleKeywords = []string{"DEADLINE", "SCHEDULED", "CLOSED"}

type Schedule struct {
	date    time.Time
	content string
	status  ScheduleStatus

	indent int
}

func NewScheduleFromReader(reader *bufio.Reader) option.Option[Schedule] {
	var err error
	var depth = 1
	var bytes []byte

	for bytes, err = reader.Peek(depth); !slices.Contains(bytes, '\n') && err == nil; bytes, err = reader.Peek(depth) {
		depth += 1
	}

	if err != nil {
		return option.None[Schedule]()
	}

	for _, keyword := range ScheduleKeywords {
		idx := strings.Index(string(bytes), keyword)

		if idx >= 0 {
			// fmt.Fprintf(os.Stderr, "Found schedule keyword (%s) at %d\n", keyword, idx)
			schedule, _ := reader.ReadBytes('\n')
			var content string

			if idx+len(keyword)+1 < len(schedule) {
				content = string(schedule[idx+len(keyword)+1:])
			}

			return option.Some(Schedule{
				status:  newStatusFromKeyword(keyword),
				content: strings.TrimSpace(content),
				indent:  idx,
			})
		}
	}

	return option.None[Schedule]()
}

func (s *Schedule) Render(builder *strings.Builder) {
	builder.WriteString(strings.Repeat(" ", s.indent))
	builder.WriteString(s.status.String())
	builder.WriteString(": ")
	builder.WriteString(s.content)
}
