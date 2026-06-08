#!/bin/bash

set -e

# Build if needed
if [ ! -f "./org-mcp" ]; then
    go build .
fi

# Run initialize and count tokens
RESULT=$(echo '{ "method":"initialize", "params":{} }' | ./org-mcp serve 2>/dev/null)
COUNT=$(echo "$RESULT" | tokens)

echo "Initialize response: $COUNT tokens"

# Optionally show breakdown by tool
echo ""
echo "Tool schemas:"
echo "$RESULT" | jq -r '.capabilities.tools[].inputSchema.properties | to_entries[] | .key' 2>/dev/null || echo "JSON parsing not available"