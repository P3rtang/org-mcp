package tools

import (
	"context"
	"time"

	"github.com/p3rtang/org-mcp/mcp"
)

type TestTaskInput struct{}

var TestTaskTool = mcp.GenericTool[TestTaskInput]{
	Name: "test_task_input",
	Description: `
Test tool for tasks capability in the mcp spec
`,
	Callback: func(_ context.Context, input TestTaskInput, options mcp.FuncOptions) ([]any, error) {
		return []any{
			mcp.NewTask(mcp.NewTaskStore(nil), func() ([]any, error) {
				time.Sleep(time.Second * 60)

				return []any{"DONE"}, nil
			}),
		}, nil
	},
}
