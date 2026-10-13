package mcp

import "strings"

// HarnessQuirk list of harness specific quirks; matched against clientInfo
// from the MCP initialize request and appended to the base instructions.
type harnessQuirk struct {
	names   []string
	section string
}

	var harnessQuirks = []harnessQuirk{
		{
			names: []string{"opencode"},
			section: `
## Known harness quirks (opencode)
The client may mangle escape sequences in tool-call string args before this server receives them (backticks can arrive as close brackets, escaped quotes can be dropped).
When quoting inside bullet text, table cells, or header content, prefer smart quotes over escaped straight quotes.
`,
		},
	}

// buildInstructions composes the initialize instructions: the base always
// applies; harness quirks are appended when clientInfo matches.
func buildInstructions(base string, clientName string) string {
	instructions := base
	for _, entry := range harnessQuirks {
		for _, name := range entry.names {
			if strings.Contains(strings.ToLower(clientName), strings.ToLower(name)) {
				instructions += entry.section
				break
			}
		}
	}
	return instructions
}

// baseInstructions is the always-included session guidance (columns doc,
// single-source-of-truth rules). Harness quirk sections are appended on top.
const baseInstructions = `When working with org-mcp, you can use the tool to parse / change org-mode files.
This should be used to increase your understanding of the project, tasks and notes.

Tooling to traverse headers, extract metadata, and update statuses are available.
Using these tools has the benifit of keeping the org-mode file consistent and reducing the amount of data in the context window.
Since targeted updates and queries are possible, there is no need to pass the entire file contents back and forth.

The org-mode file is the single source of truth for both the programmer and the LLM.
Every task, bullet point, or step must be added to the document.
Implementation begins with extracting the relevant metadata and statuses from the document.

Direct modification of the org file is discouraged. Use the tooling provided by the org-mcp server to ensure consistency and maintain integrity.
Direct reading should be avoided as it may lead to blowing up the context window and losing important information.

An org file serves as a long-term memory and organizational tool for the project. Always refer to it as the main reference point.
It also functions as long term memory between session, this means that any information not stored in the org file will be lost between sessions.
Use this together with the programmer to ensure that all important information is captured in the org file.

## Table format contract
Tables parse correctly only in the on-disk layout the tools produce:
- table rows are indented under their header (2 spaces); flush-left tables are not parsed
- the trailing =#+NAME: <name>= and empty =#+TYPE:= lines sit directly after the last table row; nothing may be inserted between the last row and these lines or the table loses its stable uid on the next load
- orphaned metadata (NAME/TYPE with no table rows above them) parses as plain text, not a table
- a table without a NAME gets a random uid that changes on every load until a tool call persists one
- two tables under one header need a blank line between them; contiguous | rows are one table by definition

## Error conventions
Top-level errors are reserved for server failures. Tool-level mistakes (bad op, not found, duplicate, out of range) arrive as response items, each describing what failed, the actual state, and the remedy in real op names.
An empty response usually means the operation legitimately had nothing to report — verify with a query before retrying.

## Known response shapes
Batched tools return one result chunk per operation with neighborhood context; rows can repeat across chunks and earlier chunks show superseded state. The final on-disk result counts — re-query when uncertain instead of reconstruing from batched chunks.

## Columns
!!IMPORTANT!!
Tools with output that can be large have the option to specify columns to return in the response.
This allows you to get only the relevant information without blowing up the context window.
Possible column to specify in a return are the following
	- UID: Unique identifier for the org item.
	- PREVIEW: A short preview of the header content (first 60 characters).
	- CONTENT: The full content of the header, including all text and metadata. Use with caution as it can be very large.
	- PARENT: The UID of the parent header, if any.
	- LEVEL: The depth level of the header in the org file (e.g. 1 for top-level headers, 2 for their children, etc.).
	- STATUS: The TODO or checkbox status of the header or bullet (e.g. TODO, DONE, CHECKED, UNCHECKED).
	- TAGS: A comma-separated list of tags associated with the header.
	- PROGRESS: The progress of the header/bullet formatted as "X/Y" where X is the number of completed child items and Y is the total number of child items.
	- CHILDREN_COUNT: The number of child headers under this header or bullet.
	- PATH: The hierarchical path to the header in the org file, represented as a string of header/bullet/etc. uids separated by "/".
	- SCHEDULED: The scheduled date of the header, if any.
	- DEADLINE: The deadline date of the header, if any.
	- CLOSED: The closed date of the header, if any.`
