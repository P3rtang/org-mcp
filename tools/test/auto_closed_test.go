package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/p3rtang/org-mcp/mcp"
	"github.com/p3rtang/org-mcp/orgmcp"
	"github.com/p3rtang/org-mcp/tools"
)

func TestAutoClosed(t *testing.T) {
	of, err := mcp.LoadOrgFile(context.TODO(), "./test.org")
	if err != nil {
		t.Fatalf("failed to load org file: %v", err)
	}

	doneUid := NewTestHeader(&of)
	revwUid := NewTestHeader(&of)
	todoUid := NewTestHeader(&of)

	mcp.WriteOrgFileToDisk(context.TODO(), of, "./test.org")

	tests := []ManageHeaderTest{
		{
			name: "DoneAutoClose",
			input: tools.HeaderInput{
				Headers: []tools.HeaderValue{
					{Uid: doneUid.String(), Status: "DONE", Method: "update"},
				},
				Columns: []*orgmcp.Column{
					&orgmcp.ColUidValue, &orgmcp.ColStatusValue, &orgmcp.ColClosedValue,
				},
			},
			expected: []any{fmt.Sprintf("UID,STATUS,CLOSED\n%s,DONE,%s\n", doneUid, time.Now().Format("2006-01-02 15:04"))},
		},
		{
			name: "ReviewAutoClose",
			input: tools.HeaderInput{
				Headers: []tools.HeaderValue{
					{Uid: revwUid.String(), Status: "REVW", Method: "update"},
				},
				Columns: []*orgmcp.Column{
					&orgmcp.ColUidValue, &orgmcp.ColStatusValue, &orgmcp.ColClosedValue,
				},
			},
			expected: []any{fmt.Sprintf("UID,STATUS,CLOSED\n%s,REVW,%s\n", revwUid, time.Now().Format("2006-01-02 15:04"))},
		},
		{
			name: "NoAutoClose",
			input: tools.HeaderInput{
				Headers: []tools.HeaderValue{
					{Uid: todoUid.String(), Status: "TODO", Method: "update"},
				},
				Columns: []*orgmcp.Column{
					&orgmcp.ColUidValue, &orgmcp.ColStatusValue, &orgmcp.ColClosedValue,
				},
			},
			expected: []any{fmt.Sprintf("UID,STATUS,CLOSED\n%s,TODO,\n", todoUid)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := tools.HeaderTool.Callback(context.TODO(), tt.input, mcp.FuncOptions{DefaultPath: "./test.org"})
			if err != nil {
				t.Errorf("HeaderTool failed: %v", err)
			}

			for _, expectedStr := range tt.expected {
				found := false
				for _, v := range res {
					str, ok := v.(string)
					if !ok {
						t.Errorf("expected string response, got %T", v)
					}

					if EqualString(str, expectedStr.(string)) {
						found = true
						break
					}
				}

				if !found {
					t.Errorf("expected string not found in response: %s", expectedStr.(string))
				}
			}
		})
	}

	// Clean up
	err = RemoveTestHeader(&of, doneUid)
	err = RemoveTestHeader(&of, revwUid)
	err = RemoveTestHeader(&of, todoUid)
	_, err = mcp.WriteOrgFileToDisk(context.TODO(), of, "./test.org")

	if err != nil {
		t.Fatalf("failed to remove test header: %v", err)
	}
}
