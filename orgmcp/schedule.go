package orgmcp

import (
	"github.com/p3rtang/org-mcp/utils/option"
	"github.com/p3rtang/org-mcp/utils/reader"
	"regexp"
	"strings"
	"time"
)

type ScheduleStatus int

const (
	Deadline ScheduleStatus = iota
	Scheduled
	Closed
)

// func newStatusFromKeyword(kw string) ScheduleStatus {
// 	switch kw {
// 	case "DEADLINE":
// 		return Deadline
// 	case "SCHEDULED":
// 		return Scheduled
// 	case "CLOSED":
// 		return Closed
// 	default:
// 		panic("unreachable")
// 	}
// }

func (s ScheduleStatus) String() string {
	return ScheduleKeywords[s]
}

var ScheduleKeywords = []string{"DEADLINE", "SCHEDULED", "CLOSED"}

type Schedule struct {
	values map[ScheduleStatus]struct {
		t        time.Time
		withTime bool
	}

	parent *Header
}

func NewScheduleFromReader(reader *reader.PeekReader) option.Option[Schedule] {
	var err error
	schedule := Schedule{}
	schedule.values = make(map[ScheduleStatus]struct {
		t        time.Time
		withTime bool
	})

	bytes, err := reader.PeekBytes('\n')

	if err != nil {
		return option.None[Schedule]()
	}

	content := string(bytes)

	r := regexp.MustCompile(".*SCHEDULED: <(.*?)( .*?)?>.*")
	matches := r.FindStringSubmatch(content)
	// fmt.Fprintf(os.Stderr, "matches: %v", matches)

	if len(matches) != 0 {
		if t, err := time.Parse("2006-01-02", matches[1]); err == nil {
			schedule.values[Scheduled] = struct {
				t        time.Time
				withTime bool
			}{t: t, withTime: false}
		}
	}

	r = regexp.MustCompile(".*DEADLINE: <(.*?)( .*?)?>.*")
	matches = r.FindStringSubmatch(content)
	// fmt.Fprintf(os.Stderr, "matches: %v", matches)

	if len(matches) != 0 {
		if t, err := time.Parse("2006-01-02", matches[1]); err == nil {
			schedule.values[Deadline] = struct {
				t        time.Time
				withTime bool
			}{t: t, withTime: false}
		}
	}

	r = regexp.MustCompile(`.*CLOSED: \[(.*? )(.*? )?(.*?)?\].*`)
	matches = r.FindStringSubmatch(content)

	if len(matches) != 0 {
		if t, err := time.Parse("2006-01-02", matches[1]); err == nil {
			schedule.values[Closed] = struct {
				t        time.Time
				withTime bool
			}{t: t, withTime: false}
		}
	}

	if len(matches) > 3 {
		if t, err := time.Parse("2006-01-02 15:04", matches[1]+matches[3]); err == nil {
			schedule.values[Closed] = struct {
				t        time.Time
				withTime bool
			}{t: t, withTime: true}
		}
	}

	if len(schedule.values) == 0 {
		return option.None[Schedule]()
	}

	reader.Continue()
	return option.Some(schedule)
}

func (s *Schedule) Render(builder *strings.Builder) {
	builder.WriteString(strings.Repeat(" ", s.parent.IndentLevel()-1))

	for status, t := range s.values {
		builder.WriteString(" ")
		builder.WriteString(status.String())
		builder.WriteString(": ")

		if status == Closed {
			builder.WriteRune('[')
		} else {
			builder.WriteRune('<')
		}

		builder.WriteString(t.t.Format("2006-01-02"))
		builder.WriteRune(' ')
		builder.WriteString(t.t.Weekday().String()[:3])

		if t.withTime {
			builder.WriteRune(' ')
			builder.WriteString(t.t.Format("15:04"))
		}

		if status == Closed {
			builder.WriteRune(']')
		} else {
			builder.WriteRune('>')
		}
	}

	builder.WriteString("\n")
}
