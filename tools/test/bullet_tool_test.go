package test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/p3rtang/org-mcp/mcp"
	"github.com/p3rtang/org-mcp/tools"
)

func TestBulletTool(t *testing.T) {
	tests := []ManageBulletTest{
		{
			name: "MoveDown",
			file: "./test_org/move_bullet.org",
			input: tools.BulletInput{
				Bullets: []tools.BulletValue{
					{
						Uid:       "95718930.b0",
						Method:    "move_relative",
						MoveValue: 1,
					},
				},
			},
		},
	}

	ctx := context.TODO()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, err := tools.BulletTool.Callback(ctx, test.input, mcp.FuncOptions{DefaultPath: test.file})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			fmt.Fprintf(os.Stderr, "%v", resp)

			// for i := range resp {
			// 	if resp[i] != test.expected[i] {
			// 		t.Errorf("expected response[%d] to be %v, got %v", i, test.expected[i], resp[i])
			// 	}
			// }
		})
	}
}
