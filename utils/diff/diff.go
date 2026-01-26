package diff

import (
	"fmt"

	"github.com/hexops/gotextdiff"
	"github.com/hexops/gotextdiff/myers"
	"github.com/hexops/gotextdiff/span"
)

// GetDiff generates a unified diff between two strings in pure Go.
// This works across all platforms (Linux, Windows, macOS) without external dependencies.
func GetDiff(oldName, oldContent, newContent string) string {
	edits := myers.ComputeEdits(span.URIFromPath(oldName), oldContent, newContent)
	diff := gotextdiff.ToUnified(oldName, oldName, oldContent, edits)
	return fmt.Sprintf("%v\n", diff)
}
