# org-mcp

<p>
  <a href="https://glama.ai/mcp/servers"><img src="https://glama.ai/mcp/servers/org-mcp/badge" alt="MCP Server"></a>
  <a href="https://github.com/p3rtang/org-mcp/releases"><img src="https://img.shields.io/github/v/release/p3rtang/org-mcp" alt="Release"></a>
  <a href="https://github.com/p3rtang/org-mcp/blob/main/LICENSE"><img src="https://img.shields.io/github/license/p3rtang/org-mcp" alt="License"></a>
</p>

An MCP (Model Context Protocol) server that gives AI models direct access to your Emacs Org-mode files. Build AI-powered workflows on top of one of the most powerful plain-text task management systems in existence.

## Why org-mcp?

If you already use Emacs Org-mode, you know its power. But now you can supercharge it with AI:

- **AI-Assisted Task Management** - Let AI help organize, prioritize, and manage your org files
- **Semantic Search** - Find relevant headers using natural language queries
- **Automated Metadata** - Automatically set timestamps, properties, and status changes
- **Bridge to Modern AI Tools** - Connect Claude, GPT, or any MCP-compatible AI to your existing org workflow
- **Persistent memory** - Stable UIDs and structured metadata ensures a consistent experience across interactions

## Two Modes

org-mcp operates in two distinct modes depending on your needs:

### MCP Server Mode (`serve`)

The default mode. Connect org-mcp to your AI editor (Zed, VS Code, Cursor, etc.) and interact with your org files through AI conversation.

```bash
# Run as MCP server
./org-mcp serve
```

When running in MCP mode, the server listens for JSON-RPC messages on stdin and responds on stdout. This is the mode to use when integrating with AI assistants.

### Export Mode (`export`)

Convert Org files to Markdown for sharing or publishing.

```bash
# Export to markdown (default input: .tasks.org, default output: out.md)
./org-mcp export

# Custom input/output
./org-mcp export --input my-tasks.org --output README.md
```

## Features

| Tool | Description |
|------|-------------|
| `manage_header` | Create, update, remove headers with full status tracking (TODO -> PROG -> DONE) |
| `manage_bullet` | Add, remove, complete, toggle checklist items |
| `manage_text` | Add or update plain text content within headers |
| `query_items` | Query headers with filters and return token-efficient CSV |
| `vector_search` | Semantic search across all headers using embeddings |
| `status` | Quick status changes with automatic CLOSED timestamps |

- **Full support for basic org mode items**: Properties, tags, scheduled/deadline dates, CLOSED timestamps
- **CSV Output**: All query results return CSV for maximum token efficiency
- **ID Persistence**: Stable UIDs for headers that survive across operations
- **Structured Metadata**: Automatic property drawer management

## Quick Start

### 1. Build the Server

```bash
go build -o org-mcp .
```

### 2. Configure Your Editor

#### Zed

Add to your `~/.config/zed/settings.json`:

```json
{
  "mcp_servers": {
    "org-mcp": {
      "command": "/path/to/org-mcp",
      "args": ["serve"]
    }
  }
}
```

#### VS Code / Cursor

Use the MCP client extension and configure:

```json
{
  "mcpServers": {
    "org-mcp": {
      "command": "/path/to/org-mcp",
      "args": ["serve"]
    }
  }
}
```

### 3. Start Chatting with AI

```
AI: Show me my TODO tasks
[org-mcp returns CSV of all TODO items]

AI: Mark task 12345 as DONE
[org-mcp updates status and sets CLOSED timestamp]

AI: Add a new task for reviewing PRs
[org-mcp creates new header with TODO status]

AI: Could you find the most relevant notes about "debugging tips"?
[org-mcp performs vector search and returns top results]

AI: What is scheduled for next week?
[org-mcp queries for items with scheduled dates in the next 7 days]
```

## Usage Examples

### Query Tasks by Status

```json
{
  "items": [{"status": "TODO"}],
  "columns": ["UID", "STATUS", "PREVIEW", "TAGS"]
}
```

### Create a New Header

```json
{
  "headers": [{
    "content": "Review PR #42",
    "status": "TODO",
    "method": "add",
    "uid": "root"
  }]
}
```

### Update Header Status

```json
{
  "headers": [{
    "method": "update",
    "status": "DONE",
    "uid": "12345"
  }]
}
```

### Semantic Search

```json
{
  "query": "debugging tips",
  "top_n": 5
}
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `MARKDOWN_COLOR` | Render markdown with github compatible colors |

## Command-Line Options

### serve

```
Usage: org-mcp serve [flags]

Options:
  -h, --help          Help for serve
```

### export

```
Usage: org-mcp export [flags]

Options:
  -i, --input string    Input Org file (default: .tasks.org)
  -o, --output string   Output Markdown file (default: out.md)
  -h, --help           Help for export
```

## Example Use Cases

- **AI Project Manager**: Let AI read your org file, suggest priorities, and update tasks
- **Meeting Notes**: AI summarizes conversations and creates action items in org format
- **Knowledge Base**: Vector search your org files for instant answers
- **Automated Workflows**: AI can trigger status changes, add tags, or schedule tasks
- **Writing Assistant**: Use AI to help structure and organize org documents
- **Persistent Memory**: With stable UIDs and structured metadata, AI can maintain context across interactions, making it ideal for long-term task management and knowledge retention.

## Related

- [Emacs Org-mode](https://orgmode.org/) - The world's best plain-text task management system
- [Model Context Protocol](https://spec.modelcontextprotocol.io/) - Open protocol for AI tool integration
- [Zed Editor](https://zed.dev/) - AI-native code editor with MCP support
