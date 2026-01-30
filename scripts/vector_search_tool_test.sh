#!/bin/bash

go build .

json='{ "method":"tools/call", "params":{ "name":"vector_search", "arguments": {"query": "testing", "top_n": 3} } }'

echo "$json" | ./org-mcp 2>/dev/null
