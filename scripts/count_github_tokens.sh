#!/bin/bash

set -e

PAT="${1:-${GITHUB_MCP_PAT}}"

if [ -z "$PAT" ]; then
    echo "Usage: $0 <github_pat>"
    echo "Or set GITHUB_MCP_PAT environment variable"
    exit 1
fi

# Step 1: Initialize
echo "=== Initialize ==="
INIT=$(curl -s -X POST \
  "https://api.githubcopilot.com/mcp/" \
  -H "Authorization: Bearer $PAT" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"token-counter","version":"1.0.0"}}}')

echo "$INIT" | tokens
echo ""

# Step 2: Try tools/list
echo "=== tools/list ==="
LIST=$(curl -s -X POST \
  "https://api.githubcopilot.com/mcp/" \
  -H "Authorization: Bearer $PAT" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}')

echo "$LIST" | tokens
echo ""

# Step 3: Try tools/capabilities
echo "=== tools/capabilities ==="
CAPS=$(curl -s -X POST \
  "https://api.githubcopilot.com/mcp/" \
  -H "Authorization: Bearer $PAT" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":3,"method":"tools/capabilities","params":{}}')

echo "$CAPS" | tokens
echo ""

# Step 4: Try server/discovery
echo "=== server/discover ==="
LIST=$(curl -s -X POST \
  "https://api.githubcopilot.com/mcp/" \
  -H "Authorization: Bearer $PAT" \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","id":2,"method":"server/discover","params":{}}')

echo "$LIST" | tokens
echo "$LIST"
echo ""
