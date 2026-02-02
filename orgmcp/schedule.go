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

var (
	DeadlineValue  = "DEADLINE"
	ScheduledValue = "SCHEDULED"
	ClosedValue    = "CLOSED"
)

func NewScheduleStatus(str string) (ScheduleStatus, error) {
	switch strings.ToUpper(strings.TrimSpace(str)) {
	case "DEADLINE":
		return Deadline, nil
	case "SCHEDULED":
		return Scheduled, nil
	case "CLOSED":
		return Closed, nil
	default:
		return 0, nil
	}
}

func (s *ScheduleStatus) UnmarshalJSON(data []byte) (err error) {
	schedule, err := NewScheduleStatus(strings.Trim(string(data), "\""))
	s = &schedule

	return
}

func (s ScheduleStatus) String() string {
	return ScheduleKeywords[s]
}

var ScheduleKeywords = []string{"DEADLINE", "SCHEDULED", "CLOSED"}
var OrderedSchedules = []ScheduleStatus{Scheduled, Deadline, Closed}

type Schedule struct {
	Values map[ScheduleStatus]struct {
		T        time.Time
		withTime bool
	}

	parent *Header
}

func NewScheduleFromReader(reader *reader.PeekReader) option.Option[Schedule] {
	var err error
	schedule := Schedule{}
	schedule.Values = make(map[ScheduleStatus]struct {
		T        time.Time
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
			schedule.Values[Scheduled] = struct {
				T        time.Time
				withTime bool
			}{T: t, withTime: false}
		}
	}

	r = regexp.MustCompile(".*DEADLINE: <(.*?)( .*?)?>.*")
	matches = r.FindStringSubmatch(content)
	// fmt.Fprintf(os.Stderr, "matches: %v", matches)

	if len(matches) != 0 {
		if t, err := time.Parse("2006-01-02", matches[1]); err == nil {
			schedule.Values[Deadline] = struct {
				T        time.Time
				withTime bool
			}{T: t, withTime: false}
		}
	}

	r = regexp.MustCompile(`.*CLOSED: \[(.*? )(.*? )?(.*?)?\].*`)
	matches = r.FindStringSubmatch(content)

	if len(matches) != 0 {
		if t, err := time.Parse("2006-01-02", matches[1]); err == nil {
			schedule.Values[Closed] = struct {
				T        time.Time
				withTime bool
			}{T: t, withTime: false}
		}
	}

	if len(matches) > 3 {
		if t, err := time.Parse("2006-01-02 15:04", matches[1]+matches[3]); err == nil {
			schedule.Values[Closed] = struct {
				T        time.Time
				withTime bool
			}{T: t, withTime: true}
		}
	}

	if len(schedule.Values) == 0 {
		return option.None[Schedule]()
	}

	reader.Continue()
	return option.Some(schedule)
}

func (s *Schedule) Render(builder *strings.Builder) {
	// Indent according to parent's child indent level
	// subtract 1 to account for the space before the schedule keywords bound to a minimum of 0
	builder.WriteString(strings.Repeat(" ", max(s.parent.ChildIndentLevel()-1, 0)))

	for _, status := range OrderedSchedules {
		t := s.Values[status]
		if t.T.IsZero() {
			continue
		}

		builder.WriteString(" ")
		builder.WriteString(status.String())
		builder.WriteString(": ")

		if status == Closed {
			builder.WriteRune('[')
		} else {
			builder.WriteRune('<')
		}

		builder.WriteString(t.T.Format("2006-01-02"))
		builder.WriteRune(' ')
		builder.WriteString(t.T.Weekday().String()[:3])

		if t.withTime {
			builder.WriteRune(' ')
			builder.WriteString(t.T.Format("15:04"))
		}

		if status == Closed {
			builder.WriteRune(']')
		} else {
			builder.WriteRune('>')
		}
	}

	builder.WriteString("\n")
}
